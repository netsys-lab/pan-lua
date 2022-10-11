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

	"github.com/netsec-ethz/scion-apps/pkg/pan"
	"github.com/netsys-lab/pan-lua/rpc"
	"github.com/yuin/gopher-lua"
)

func newLuaPathInterface(intf pan.PathInterface) *lua.LTable {
	iface := lua.LTable{}
	iface.RawSetString("IA", lua.LString(intf.IA.String()))
	iface.RawSetString("IfID", lua.LNumber(intf.IfID))
	return &iface

}

func newLuaPreferences(prefs map[string]string) *lua.LTable {
	t := lua.LTable{}
	for k, v := range prefs {
		t.RawSetString(k, lua.LString(v))
	}
	return &t
}

func newLuaPath(path *pan.Path) *lua.LTable {
	t := lua.LTable{}
	if path != nil {
		t.RawSetString("Source", lua.LString(path.Source.String()))
		t.RawSetString("Destination", lua.LString(path.Destination.String()))
		t.RawSetString("Fingerprint", lua.LString(path.Fingerprint))
		t.RawSetString("Expiry", lua.LString(path.Expiry.String()))

		if path.Metadata != nil {
			meta := lua.LTable{}
			meta.RawSetString("MTU", lua.LNumber(path.Metadata.MTU))

			ifaces := lua.LTable{}
			for _, i := range path.Metadata.Interfaces {
				ifaces.Append(newLuaPathInterface(i))
			}
			meta.RawSetString("Interfaces", &ifaces)

			latencies := lua.LTable{}
			for _, l := range path.Metadata.Latency {
				latencies.Append(lua.LNumber(l))
			}
			meta.RawSetString("Latency", &latencies)

			bandwidths := lua.LTable{}
			for _, b := range path.Metadata.Bandwidth {
				bandwidths.Append(lua.LNumber(b))
			}
			meta.RawSetString("Bandwidth", &bandwidths)

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
	state
	mod *lua.LTable
	d   time.Duration
}

// func NewLuaSelector(script string) (*LuaSelector, error) {
func NewSelector(state *State) rpc.ServerSelector {
	state.Lock()
	defer state.Unlock()

	mod := map[string]lua.LGFunction{}
	for _, fn := range []string{
		"Initialize",
		"SetPreferences",
		"Path",
		"PathDown",
		"Refresh",
		"Close",
		"Periodic",
	} {
		s := fmt.Sprintf("function %s not implemented in script", fn)
		mod[fn] = func(L *lua.LState) int {
			state.Logger.Panic(s)
			return 0
		}
	}

	mod["Log"] = func(L *lua.LState) int {
		s := ""
		for i := 1; i <= L.GetTop(); i++ {
			s += L.Get(i).String() + " "
		}
		state.Println(s)
		return 0
	}

	mod["Now"] = func(L *lua.LState) int {
		L.Push(lua.LNumber(time.Now().UnixMicro()))
		return 1
	}

	panapi := state.RegisterModule("panapi", mod).(*lua.LTable)

	s := &LuaSelector{state, new_state(), panapi, time.Second}

	go func(s *LuaSelector) {
		old := time.Now()
		for {
			time.Sleep(s.d)
			s.Lock()
			seconds := time.Since(old).Seconds()
			s.CallByParam(
				lua.P{
					Protect: true,
					Fn:      s.mod.RawGetString("Periodic"),
					NRet:    0,
				},
				lua.LNumber(seconds),
			)
			old = time.Now()
			s.Unlock()
		}
	}(s)
	return s
}

func (s *LuaSelector) Initialize(prefs map[string]string, local, remote pan.UDPAddr, paths []*pan.Path) error {
	//s.Printf("Initialize(%s,%s,[%d]pan.Path)", local, remote, len(paths))
	s.Lock()
	defer s.Unlock()

	//assume that setpaths is called with all the currently valid options
	//meaning that anything we already know can be flushed
	s.state.clear_addr(remote)
	lpaths := s.set_paths(remote, paths)

	//call the "Initialize" function in the Lua script
	//with two arguments
	//and don't expect a return value

	err := s.CallByParam(
		lua.P{
			Protect: true,
			Fn:      s.mod.RawGetString("Initialize"),
			NRet:    0,
		},
		newLuaPreferences(prefs),
		lua.LString(local.String()),
		lua.LString(remote.String()),
		lua_table_slice_to_table(lpaths),
	)
	if err != nil {
		log.Println(err)
	}
	return err

}
func (s *LuaSelector) SetPreferences(prefs map[string]string, local, remote pan.UDPAddr) error {
	s.Lock()
	defer s.Unlock()

	return s.CallByParam(lua.P{
		Protect: true,
		Fn:      s.mod.RawGetString("SetPreferences"),
		NRet:    0},
		newLuaPreferences(prefs),
		lua.LString(local.String()),
		lua.LString(remote.String()),
	)
}

func (s *LuaSelector) Path(local, remote pan.UDPAddr) (*pan.Path, error) {
	s.Lock()
	defer s.Unlock()

	//call the "Path" function from the Lua script
	//expect 1 return value
	err := s.CallByParam(lua.P{
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
	return s.state.get_pan_path(lt), nil
}

func (s *LuaSelector) PathDown(local, remote pan.UDPAddr, fp pan.PathFingerprint, pi pan.PathInterface) error {
	//s.l.Println("PathDown()")
	s.Lock()
	defer s.Unlock()
	//s.Printf("PathDown called with fp %v and pi %v", fp, pi)
	return s.CallByParam(
		lua.P{
			Fn:      s.mod.RawGetString("PathDown"),
			NRet:    0,
			Protect: true,
		},
		lua.LString(local.String()),
		lua.LString(remote.String()),
		lua.LString(fp),
		newLuaPathInterface(pi),
	)

}

func (s *LuaSelector) Refresh(local, remote pan.UDPAddr, paths []*pan.Path) error {
	s.Println("Refresh()")
	s.Lock()
	defer s.Unlock()

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
	)
}

func (s *LuaSelector) Close(local, remote pan.UDPAddr) error {
	s.Lock()
	defer s.Unlock()

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
}
