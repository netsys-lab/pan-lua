// Copyright 2021,2022,2023 Thorben Krüger (thorben.krueger@ovgu.de)
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
package luajit

import (
	//"fmt"
	"log"
	"time"

	"github.com/netsec-ethz/scion-apps/pkg/pan"
	"github.com/netsys-lab/pan-lua/rpc"
)

/*
   #cgo LDFLAGS: -lluajit-5.1
   #cgo CFLAGS: -I /usr/include/luajit-2.1
   #include <lauxlib.h>
   #include <lualib.h>
   #include <stdlib.h>
   #include <stdio.h>
   #include <lua.h>
*/
import "C"

func (s *State) pushLuaPathInterface(intf pan.PathInterface) {
	s.NewTable()
	s.PushString(intf.IA.String())
	s.SetField(-2, "IA")
	s.PushInteger(int64(intf.IfID))
	s.SetField(-2, "IfID")
}

func (s *State) pushLuaPreferences(prefs map[string]string) {
	s.NewTable()
	for k, v := range prefs {
		s.PushString(v)
		s.SetField(-2, k)
	}
}

func (s *State) pushLuaPath(path *pan.Path) {
	if path != nil {
		s.NewTable()
		s.PushString(path.Source.String())
		s.SetField(-2, "Source")
		s.PushString(path.Destination.String())
		s.SetField(-2, "Destination")
		s.PushString(string(path.Fingerprint))
		s.SetField(-2, "Fingerprint")
		s.PushString(path.Expiry.String())
		s.SetField(-2, "Expiry")

		if path.Metadata != nil {
			s.NewTable()
			s.PushInteger(int64(path.Metadata.MTU))
			s.SetField(-2, "MTU")

			s.NewTable()
			for i, intf := range path.Metadata.Interfaces {
				s.PushInteger(int64(i + 1))
				s.pushLuaPathInterface(intf)
				s.SetTable(-3)
			}
			s.SetField(-2, "Interfaces")

			s.NewTable()
			for i, latency := range path.Metadata.Latency {
				s.PushInteger(int64(i + 1))
				s.PushNumber(float64(latency))
				s.SetTable(-3)
			}
			s.SetField(-2, "Latency")

			s.NewTable()
			for i, bandwidth := range path.Metadata.Bandwidth {
				s.PushInteger(int64(i + 1))
				s.PushInteger(int64(bandwidth))
				s.SetTable(-3)
			}
			s.SetField(-2, "Bandwidth")

			s.NewTable()
			for i, linktype := range path.Metadata.LinkType {
				s.PushInteger(int64(i + 1))
				s.PushInteger(int64(linktype))
				s.SetTable(-3)
			}
			s.SetField(-2, "LinkType")

			s.NewTable()
			for i, internalHops := range path.Metadata.InternalHops {
				s.PushInteger(int64(i + 1))
				s.PushInteger(int64(internalHops))
				s.SetTable(-3)

			}
			s.SetField(-2, "InternalHops")

			s.NewTable()
			for i, note := range path.Metadata.Notes {
				s.PushInteger(int64(i + 1))
				s.PushString(note)
				s.SetTable(-3)
			}
			s.SetField(-2, "Notes")

			s.NewTable()
			for i, geo := range path.Metadata.Geo {
				s.PushInteger(int64(i + 1))
				s.NewTable()
				s.PushInteger(int64(geo.Latitude))
				s.SetField(-2, "Latitude")
				s.PushInteger(int64(geo.Longitude))
				s.SetField(-2, "Longitude")
				s.PushString(geo.Address)
				s.SetField(-2, "Address")
				s.SetTable(-3)
			}
			s.SetField(-2, "Geo")
			s.SetField(-2, "Metadata")
		}
	}
}

type state struct {
	ppaths map[string]*pan.Path
}

func new_state() state {
	return state{make(map[string]*pan.Path)}
}

type LuaSelector struct {
	*State
	state
	d time.Duration
}

var (
	PANAPI      *C.char = C.CString("panapi")
	Path        *C.char = C.CString("Path")
	Fingerprint *C.char = C.CString("Fingerprint")
)

// func NewLuaSelector(script string) (*LuaSelector, error) {
func NewSelector(state *State) rpc.ServerSelector {
	state.Lock()
	defer state.Unlock()

	state.NewTable()

	/*state.PushGoFunction(func(L *lua.State) int {
		s := ""
		for i := 1; i <= L.GetTop(); i++ {
			s += L.ToString(i) + " "
		}
		state.Println(s)
		return 0
	})
	state.SetField(-2, "Log")

	for _, fn := range []string{
		"Initialize",
		"SetPreferences",
		"Path",
		"PathDown",
		"Refresh",
		"Close",
		"Periodic",
	} {
		/*s := fmt.Sprintf("Function %s not implemented in script", fn)
		state.PushGoFunction(func(L *lua.State) int {
			state.GetGlobal(PANAPI)
			state.GetField(-1, "Log")
			state.PushString(s)
			state.Call(1, 0)
			state.Pop(1)
			return 0
		})
		state.SetField(-2, fn)

	}

	state.PushGoFunction(func(L *lua.State) int {
		L.PushInteger(time.Now().UnixMicro())
		return 1
	})
	state.SetField(-2, "Now")
	*/

	state.SetCGlobal(PANAPI)

	s := &LuaSelector{state, new_state(), time.Second}

	go func(s *LuaSelector) {
		old := time.Now()
		for {
			time.Sleep(s.d)
			seconds := time.Since(old).Seconds()
			s.Lock()
			s.GetCGlobal(PANAPI)
			s.GetField(-1, "Periodic")
			s.PushNumber(seconds)
			if err := s.Call(1, 0); err != nil {
				state.Println("Periodic:", err)
			}
			s.Pop(1)
			s.Unlock()
			old = time.Now()
		}
	}(s)
	return s
}

func (s *LuaSelector) Initialize(prefs map[string]string, local, remote pan.UDPAddr, paths []*pan.Path) error {
	//s.Printf("Initialize(%s,%s,[%d]pan.Path)", local, remote, len(paths))
	s.Lock()
	defer s.Unlock()

	s.GetCGlobal(PANAPI)
	s.GetField(-1, "Initialize")
	s.Remove(-2)
	s.NewTable()
	for k, v := range prefs {
		s.PushString(v)
		s.SetField(-2, k)
	}
	s.PushString(local.String())
	s.PushString(remote.String())
	s.NewTable()
	for i, path := range paths {
		s.state.ppaths[string(path.Fingerprint)] = path
		s.PushInteger(int64(i + 1))
		s.pushLuaPath(path)
		s.SetTable(-3)
	}
	err := s.Call(4, 0)
	if err != nil {
		log.Println(err)
	}
	return err

}
func (s *LuaSelector) SetPreferences(prefs map[string]string, local, remote pan.UDPAddr) error {
	s.Lock()
	defer s.Unlock()

	s.GetCGlobal(PANAPI)
	s.GetField(-1, "SetPreferences")
	s.Remove(-2)

	s.NewTable()
	for k, v := range prefs {
		s.PushString(v)
		s.SetField(-2, k)
	}
	s.PushString(local.String())
	s.PushString(remote.String())
	err := s.Call(3, 0)
	if err != nil {
		log.Println(err)
	}
	return err
}

func (s *LuaSelector) Path(local, remote pan.UDPAddr) (*pan.Path, error) {
	s.Lock()
	defer s.Unlock()

	s.GetCGlobal(PANAPI)
	s.GetCField(-1, Path)
	s.Remove(-2)
	s.PushString(local.String())
	s.PushString(remote.String())
	err := s.Call(2, 1)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	s.GetCField(-1, Fingerprint)
	fingerprint := s.ToString(-1)
	s.Pop(2) // pop path as well as its fingerprint
	return s.state.ppaths[fingerprint], err
}

func (s *LuaSelector) PathDown(local, remote pan.UDPAddr, fp pan.PathFingerprint, pi pan.PathInterface) error {
	//s.l.Println("PathDown()")
	s.Lock()
	defer s.Unlock()

	s.GetCGlobal(PANAPI)
	s.GetField(-1, "PathDown")
	s.Remove(-2)
	s.PushString(local.String())
	s.PushString(remote.String())
	s.PushString(string(fp))
	s.pushLuaPathInterface(pi)
	err := s.Call(4, 0)
	if err != nil {
		log.Println(err)
	}
	return err
}

func (s *LuaSelector) Refresh(local, remote pan.UDPAddr, paths []*pan.Path) error {
	s.Println("Refresh()")
	s.Lock()
	defer s.Unlock()

	s.GetCGlobal(PANAPI)
	s.GetField(-1, "Refresh")
	s.Remove(-2)
	s.PushString(local.String())
	s.PushString(remote.String())
	s.NewTable()
	for i, path := range paths {
		s.state.ppaths[string(path.Fingerprint)] = path
		s.PushInteger(int64(i + 1))
		s.pushLuaPath(path)
		s.SetTable(-3)
	}
	if err := s.Call(3, 0); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (s *LuaSelector) Close(local, remote pan.UDPAddr) error {
	s.Lock()
	defer s.Unlock()

	s.GetCGlobal(PANAPI)
	s.GetField(-1, "Close")
	s.Remove(-2)
	s.PushString(local.String())
	s.PushString(remote.String())

	if err := s.Call(2, 0); err != nil {
		log.Println(err)
		return err
	}
	return nil
}
