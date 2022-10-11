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
		var qconf quic.Config
		selector, tracer, err := lib.RPCClientHelper()
		selector.SetPreferences(map[string]string{"ConnCapacityProfile": p})
		if err != nil {
			log.Fatalln(err)
		}
		qconf.Tracer = tracer
		err = runClient(bytes, remote, selector, &qconf, &tlsConf)
		if err != nil {
			log.Println(err)
		}
	}
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

func report(c chan int, verbose bool) {
	total := 0
	subtotal := 0
	ticker := time.Tick(time.Second)
	begin := time.Now()
	for {
		select {
		case bytes := <-c:
			subtotal += bytes
		case <-ticker:
			total += subtotal
			dur := time.Since(begin)
			fmt.Printf("%d,%d,%d\n", int(dur.Seconds()), subtotal, total)
			if verbose {
				log.Printf("%.3f kb/s", float64(subtotal)/1000)
			}
			subtotal = 0

		}
	}
}

func runServer(local string, tlsconf *tls.Config) error {
	addr, err := pan.ResolveUDPAddr(local)
	if err != nil {
		return err
	}
	listener, err := pan.ListenQUIC(context.TODO(), netaddr.IPPortFrom(addr.IP, addr.Port), nil, tlsconf, nil)
	if err != nil {
		return err
	}
	running := false
	c := make(chan int)
	for {
		session, err := listener.Accept(context.TODO())
		if err != nil {
			log.Println(err)
			continue
		}
		connection, err := session.AcceptStream(context.TODO())
		if err != nil {
			log.Println(err)
			continue
		}
		if !running {
			go report(c, true)
			running = true
		}
		log.Printf("Got Connection from: %s", session.RemoteAddr())
		if err != nil {
			log.Println(err)
		} else {
			go myCopy(ioutil.Discard, connection, c)
		}
	}
}

func runClient(bytes int64, remote string, selector pan.Selector, qconf *quic.Config, tlsConf *tls.Config) error {
	addr, err := pan.ResolveUDPAddr(remote)
	if err != nil {
		return err
	}
	session, err := pan.DialQUIC(
		context.Background(),
		netaddr.IPPort{},
		addr,
		nil,
		selector,
		"",
		tlsConf,
		qconf,
	)
	if err != nil {
		return err
	}

	stream, err := session.OpenStream() //Sync(context.Background())

	if err != nil {
		return fmt.Errorf("Initate error: %s", err)
	}
	defer stream.Close()

	c := make(chan int)
	go report(c, false)

	reader := io.LimitReader(rand.Reader, bytes)
	begin := time.Now()
	n, err := myCopy(stream, reader, c)
	if err == nil {
		log.Printf("Average: %.3f kb/s", float64(n)/(1000*time.Since(begin).Seconds()))
	}
	return err
}
