// Copyright 2021,2022 Thorben Krüger (thorben.krueger@ovgu.de)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"time"

	quic "github.com/lucas-clemente/quic-go"
	"github.com/netsec-ethz/scion-apps/pkg/pan"
	lib "github.com/netsys-lab/pan-lua"
	"inet.af/netaddr"
)

func main() {
	var (
		remote, local, p                             string
		server, client, daemontracer, daemonselector bool
		bytes                                        int64
	)

	flag.StringVar(&remote, "remote", "", `[Client] Remote (i.e. the server's) Address
        (e.g. 17-ffaa:1:1,[127.0.0.1]:1337`)
	flag.StringVar(&local, "local", "", `[Server] Local Address to listen on
        (e.g. 17-ffaa:1:1,[127.0.0.1]:1337`)
	flag.StringVar(&p, "profile", "CapacitySeeking", "SCION capacity profile (Default|CapacitySeeking|Scavenger|LowLatency)")
	flag.BoolVar(&daemontracer, "daemontracer", false, "use PANAPI daemon tracer")
	flag.BoolVar(&daemonselector, "daemonselector", false, "use PANAPI daemon selector")
	flag.Int64Var(&bytes, "bytes", 1000*1000*10, "amount of bytes to transfer during experiment")

	flag.Parse()

	//log.SetFlags(log.Lshortfile)

	if len(local) > 0 {
		server = true
	}
	if len(remote) > 0 {
		client = true
	}
	if server == client {
		log.Fatalln("Either specify -port for server or -remote for client")
	}

	tlsConf := lib.DummyTLSConfig()
	if len(local) > 0 {
		log.Println(runServer(local, &tlsConf))
	} else {
		var (
			qconf    quic.Config
			selector pan.Selector
		)

		if daemontracer || daemonselector {
			s, tracer, err := lib.RPCClientHelper()
			if err != nil {
				log.Println(err)
			}
			if daemonselector && s != nil {
				s.SetPreferences(map[string]string{"ConnCapacityProfile": p})
				selector = s
			}
			if daemontracer && tracer != nil {
				qconf.Tracer = tracer
			}
		} else {
			log.Println("Skipping Daemon connection")
		}
		err := runClient(bytes, remote, selector, &qconf, &tlsConf)
		if err != nil {
			log.Println(err)
		}
	}
	log.Println("bye")
}

func myCopy(w io.Writer, r io.Reader, c chan int) (total int64, err error) {
	buf := make([]byte, 1024*32)
	for {
		var nr int
		nr, err = r.Read(buf)
		nw, erw := w.Write(buf[:nr])
		total += int64(nw)
		c <- nr
		if erw == nil && nw != nr {
			err = fmt.Errorf("short write")
			break
		}
		if err == io.EOF {
			err = nil
			break
		}
		if erw != nil {
			err = fmt.Errorf("Write error: %s", erw)
			break
		}
		if err != nil {
			err = fmt.Errorf("Read err: %s", err)
			break
		}
	}
	return
}

func report(c chan pan.ConnStats, verbose bool) {
	total := 0
	subtotal := 0
	ticker := time.Tick(time.Second)
	paths := map[string]int{}
	for {
		select {
		case stats := <-c:
			subtotal += stats.Bytes
			if stats.Path != nil {
				fp := string(stats.Path.Fingerprint)
				paths[fp] += stats.Bytes
			}
		case t := <-ticker:
			total += subtotal
			for fp, bytes := range paths {
				fmt.Printf("%s,%s,%d\n", t, fp, bytes)
			}
			if verbose {
				log.Printf("%.3f kb/s", float64(subtotal)/1000)
			}
			subtotal = 0
			paths = map[string]int{}
		}
	}
}

func runServer(local string, tlsconf *tls.Config) error {
	ctx := context.Background()
	addr, err := pan.ResolveUDPAddr(ctx, local)
	if err != nil {
		return err
	}
	stats := make(chan pan.ConnStats)
	go report(stats, false)
	listener, err := pan.ListenQUICStats(ctx, netaddr.IPPortFrom(addr.IP, addr.Port), nil, tlsconf, nil, stats)
	if err != nil {
		return err
	}
	running := false

	for {
		session, err := listener.Accept(ctx)
		if err != nil {
			log.Println(err)
			continue
		}
		connection, err := session.AcceptStream(ctx)
		if err != nil {
			log.Println(err)
			continue
		}
		if !running {
			//go report(c, true)
			running = true
		}
		log.Printf("Got Connection from: %s", session.RemoteAddr())
		if err != nil {
			log.Println(err)
		} else {
			go io.Copy(ioutil.Discard, connection)
			//go myCopy(ioutil.Discard, connection, c)
		}
	}
}

func runClient(bytes int64, remote string, selector pan.Selector, qconf *quic.Config, tlsConf *tls.Config) error {
	log.Println("running client")
	ctx := context.Background()
	addr, err := pan.ResolveUDPAddr(ctx, remote)
	if err != nil {
		log.Println("error", err)
		return err
	} else {
		log.Printf("resolved address: %s", addr)
	}

	stats := make(chan pan.ConnStats)
	go report(stats, false)

	session, err := pan.DialQUICStats(
		ctx,
		netaddr.IPPort{},
		addr,
		nil,
		selector,
		"",
		tlsConf,
		qconf,
		stats,
	)
	if err != nil {
		log.Println("error", err)
		return err
	} else {
		log.Println("dialled session")
	}

	stream, err := session.OpenStream() //Sync(context.Background())

	if err != nil {
		log.Println("error", err)
		return fmt.Errorf("Initate error: %s", err)
	}
	defer stream.Close()

	reader := io.LimitReader(rand.Reader, bytes)
	begin := time.Now()
	//n, err := myCopy(stream, reader, c)
	n, err := io.Copy(stream, reader)
	if err == nil {
		log.Printf("Average: %.3f kb/s", float64(n)/(1000*time.Since(begin).Seconds()))
	} else {
		log.Println("error", err)
	}
	return err
}
