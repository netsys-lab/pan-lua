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
package lua

import (
	"fmt"
	"log"
	"time"

	"github.com/aarzilli/golua/lua"
	"github.com/netsec-ethz/scion-apps/pkg/pan"
	"github.com/netsys-lab/pan-lua/rpc"
)

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

			linktypes := lua.LTable{}
			for _, l := range path.Metadata.LinkType {
				linktypes.Append(lua.LNumber(l))
			}
			meta.RawSetString("LinkType", &linktypes)

			internalhops := lua.LTable{}
			for _, h := range path.Metadata.InternalHops {
				internalhops.Append(lua.LNumber(h))
			}
			meta.RawSetString("InternalHops", &internalhops)

			notes := lua.LTable{}
			for _, n := range path.Metadata.Notes {
				notes.Append(lua.LString(n))
			}
			meta.RawSetString("Notes", &notes)

			geo := lua.LTable{}
			for _, g := range path.Metadata.Geo {
				pos := lua.LTable{}
				pos.RawSetString("Latitude", lua.LNumber(g.Latitude))
				pos.RawSetString("Longitude", lua.LNumber(g.Longitude))
				pos.RawSetString("Address", lua.LString(g.Address))
				geo.Append(&pos)
			}
			meta.RawSetString("Geo", &geo)

			t.RawSetString("Metadata", &meta)
		}
	}
	return &t
}

func lua_table_slice_to_table(s []*lua.LTable) *lua.LTable {
	res := lua.LTable{}
	for _, t := range s {
		res.Append(t)
	}
	return &res
}

// help to translate lua to pan pointers back and forth
type state struct {
	lpaths map[string]map[string]*lua.LTable
	ppaths map[*lua.LTable]*pan.Path
}

func new_state() state {
	return state{
		make(map[string]map[string]*lua.LTable),
		make(map[*lua.LTable]*pan.Path),
	}
}

func (s state) get_pan_path(lpath *lua.LTable) *pan.Path {
	return s.ppaths[lpath]
}

func (s state) clear_addr(addr pan.UDPAddr) {
	raddr := addr.String()
	for _, lt := range s.lpaths[raddr] {
		s.ppaths[lt] = nil
	}
	s.lpaths[raddr] = map[string]*lua.LTable{}
}

func (s state) set_paths(addr pan.UDPAddr, ppaths []*pan.Path) (lpaths []*lua.LTable) {
	raddr := addr.String()
	lpaths = make([]*lua.LTable, len(ppaths))
	for i, ppath := range ppaths {
		lpath := newLuaPath(ppath)
		s.lpaths[raddr][string(ppath.Fingerprint)] = lpath
		s.ppaths[lpath] = ppath
		lpaths[i] = lpath
	}
	return
}

type LuaSelector struct {
	*State
	//	state
	//mod *lua.LTable
	d time.Duration
}

// func NewLuaSelector(script string) (*LuaSelector, error) {
func NewSelector(state *State) rpc.ServerSelector {
	state.Lock()
	defer state.Unlock()

	state.NewTable()

	state.PushGoFunction(func(L *lua.State) int {
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
		s := fmt.Sprintf("Function %s not implemented in script", fn)
		state.PushGoFunction(func(L *lua.State) int {
			state.GetGlobal("panapi")
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

	state.SetGlobal("panapi")

	s := &LuaSelector{state, time.Second}

	go func(s *LuaSelector) {
		old := time.Now()
		for {
			time.Sleep(s.d)
			seconds := time.Since(old).Seconds()
			s.Lock()
			s.GetGlobal("panapi")
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

	// //assume that setpaths is called with all the currently valid options
	// //meaning that anything we already know can be flushed
	// s.state.clear_addr(remote)
	// lpaths := s.set_paths(remote, paths)

	//call the "Initialize" function in the Lua script
	//with two arguments
	//and don't expect a return value

	// err := s.CallByParam(
	// 	lua.P{
	// 		Protect: true,
	// 		Fn:      s.mod.RawGetString("Initialize"),
	// 		NRet:    0,
	// 	},
	// 	newLuaPreferences(prefs),
	// 	lua.LString(local.String()),
	// 	lua.LString(remote.String()),
	// 	lua_table_slice_to_table(lpaths),
	// )
	var err error = nil
	if err != nil {
		log.Println(err)
	}
	return err

}
func (s *LuaSelector) SetPreferences(prefs map[string]string, local, remote pan.UDPAddr) error {
	s.Lock()
	defer s.Unlock()
	return nil

	/*	return s.CallByParam(lua.P{
		Protect: true,
		Fn:      s.mod.RawGetString("SetPreferences"),
		NRet:    0},
		newLuaPreferences(prefs),
		lua.LString(local.String()),
		lua.LString(remote.String()),
	)*/
}

func (s *LuaSelector) Path(local, remote pan.UDPAddr) (*pan.Path, error) {
	s.Lock()
	defer s.Unlock()
	return nil, nil

	//call the "Path" function from the Lua script
	//expect 1 return value
	/*err := s.CallByParam(lua.P{
		Protect: true,
		Fn:      s.mod.RawGetString("Path"),
		NRet:    1},
		lua.LString(local.String()),
		lua.LString(remote.String()),
	)
	if err != nil {
		return nil, err
	}
	lt := s.ToTable(-1)
	//pop element from the stack
	s.Pop(1)
	return s.state.get_pan_path(lt), nil*/
}

func (s *LuaSelector) PathDown(local, remote pan.UDPAddr, fp pan.PathFingerprint, pi pan.PathInterface) error {
	//s.l.Println("PathDown()")
	s.Lock()
	defer s.Unlock()
	//s.Printf("PathDown called with fp %v and pi %v", fp, pi)
	/*	return s.CallByParam(
		lua.P{
			Fn:      s.mod.RawGetString("PathDown"),
			NRet:    0,
			Protect: true,
		},
		lua.LString(local.String()),
		lua.LString(remote.String()),
		lua.LString(fp),
		newLuaPathInterface(pi),
	)*/
	return nil

}

func (s *LuaSelector) Refresh(local, remote pan.UDPAddr, paths []*pan.Path) error {
	s.Println("Refresh()")
	s.Lock()
	defer s.Unlock()

	/*
		//assume that setpaths is called with all the currently valid options
		//meaning that anything we already know can be flushed
		s.state.clear_addr(remote)
		lpaths := s.state.set_paths(remote, paths)

		//call the "setpaths" function in the Lua script
		//with two arguments
		//and don't expect a return value
		return s.CallByParam(
			lua.P{
				Protect: true,
				Fn:      s.mod.RawGetString("Refresh"),
				NRet:    0,
			},
			lua.LString(local.String()),
			lua.LString(remote.String()),
			lua_table_slice_to_table(lpaths),
		)*/
	return nil
}

func (s *LuaSelector) Close(local, remote pan.UDPAddr) error {
	s.Lock()
	defer s.Unlock()

	/*
		//call the "selectpath" function from the Lua script
		//expect 1 return value
		err := s.CallByParam(
			lua.P{
				Protect: true,
				Fn:      s.mod.RawGetString("Close"),
				NRet:    1,
			},
			lua.LString(local.String()),
			lua.LString(remote.String()),
		)

		if err != nil {
			log.Println("Close called on LuaSelector:", err)
		}
		//s.L.Close()
		return err
	*/
	return nil
}
