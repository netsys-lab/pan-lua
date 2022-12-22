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
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"runtime/pprof"

	"github.com/lucas-clemente/quic-go/logging"
	"github.com/lucas-clemente/quic-go/qlog"
	"github.com/netsec-ethz/scion-apps/pkg/pan"
	"github.com/netsys-lab/pan-lua/lua"
	//"github.com/netsys-lab/pan-lua/dummy"
	"github.com/netsys-lab/pan-lua/rpc"
	"github.com/netsys-lab/pan-lua/selector"
)

func main() {
	var (
		script string
		cpulog string
		sel    rpc.ServerSelector
		err    error
	)

	flag.StringVar(&script, "script", "", "Lua script for path selection")
	flag.StringVar(&cpulog, "cpulog", "", "Write profiling information to file")
	flag.Parse()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, os.Kill, os.Interrupt)

	l, err := net.ListenUnix("unix", rpc.DefaultDaemonAddress)
	if err != nil {
		log.Fatalf("Could not start daemon: %s", err)
	}
	log.Println("Starting daemon")

	// remove the underlying socket file on close
	l.SetUnlinkOnClose(true)

	if cpulog != "" {
		f, err := os.Create(cpulog)
		if err != nil {
			log.Fatal("cpuprofile:", err)
		}
		if err = pprof.StartCPUProfile(f); err != nil {
			log.Fatal("cpuprofile:", err)
		}
	}

	lua_state := lua.NewState()
	sel = lua.NewSelector(lua_state)
	//stats := lua.NewStats(lua_state)
	err = lua_state.LoadScript(script)
	if err != nil {
		log.Printf("Could not load path-selection script: %s", err)
		log.Println("Falling back to default selector")
		sel = rpc.NewServerSelectorFunc(func(pan.UDPAddr, pan.UDPAddr) selector.Selector {
			return &selector.DefaultSelector{}
		})
	}

	tracer := qlog.NewTracer(
		func(p logging.Perspective, connectionID []byte) io.WriteCloser {
			fname := fmt.Sprintf("/tmp/quic-tracer-%d-%x.log", p, connectionID)
			log.Println("quic tracer file opened as", fname)
			f, err := os.Create(fname)
			if err != nil {
				panic(err)
			}
			return f
		})
	//serverselector := rpc.NewServerSelectorFunc(func(raddr,
	server, err := rpc.NewServer(sel, tracer, nil) //stats)
	if err != nil {
		log.Fatalln(err)
	}
	go func() {
		log.Println("Started listening for rpc calls")
		server.Accept(l)
	}()
	sig := <-c
	log.Printf("Got signal [%s]: running defered cleanup and exiting.", sig)
	err = l.Close()
	if err != nil {
		log.Println(err)
	}
	//should be NOP if profiler is not running
	pprof.StopCPUProfile()
	os.Exit(0)
}
