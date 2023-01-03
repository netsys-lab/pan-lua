// Copyright 2022,2023 Thorben Krüger (thorben.krueger@ovgu.de)
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

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"unsafe"
)

type Table struct {
	Int2int    map[int64]int64
	Int2bool   map[int64]bool
	Int2string map[int64]string
	Int2float  map[int64]float64
	Int2table  map[int64]*Table

	String2int    map[string]int64
	String2bool   map[string]bool
	String2string map[string]string
	String2float  map[string]float64
	String2table  map[string]*Table
}

func NewTable() *Table {
	return &Table{
		map[int64]int64{},
		map[int64]bool{},
		map[int64]string{},
		map[int64]float64{},
		map[int64]*Table{},
		map[string]int64{},
		map[string]bool{},
		map[string]string{},
		map[string]float64{},
		map[string]*Table{},
	}
}

type State struct {
	L       *C.lua_State
	strptrs map[string]*C.char
	sync.Mutex
	*log.Logger
}

func NewState() *State {
	L := C.luaL_newstate()
	C.luaL_openlibs(L)
	//l := log.New(ioutil.Discard, "lua ", log.Ltime)
	l := log.Default()
	l.SetFlags(log.Ltime | log.Lshortfile)
	l.SetPrefix("luajit ")
	return &State{L, map[string]*C.char{}, sync.Mutex{}, l}
}

func (s *State) strptr(str string) *C.char {
	ptr, ok := s.strptrs[str]
	if !ok {
		ptr = C.CString(str)
		s.strptrs[str] = ptr
	}
	return ptr
}

func (s *State) Call(nargs, nresults int) error {
	C.lua_call(s.L, C.int(nargs), C.int(nresults))
	return nil
}

func (s *State) GetField(index int, name string) {
	C.lua_getfield(s.L, C.int(index), s.strptr(name))
}

func (s *State) GetGlobal(name string) {
	s.GetField(C.LUA_GLOBALSINDEX, name)
}

func (s *State) NewTable() {
	C.lua_createtable(s.L, 0, 0)
}

func (s *State) Pop(n int) {
	// FIXME
	//C.lua_pop(s.L, C.int(n));
	C.lua_settop(s.L, C.int(-n-1))
}

func (s *State) PushBoolean(b bool) {
	var cb int
	if b {
		cb = 1
	} else {
		cb = 0
	}
	C.lua_pushboolean(s.L, C.int(cb))
}

func (s *State) PushInteger(n int64) {
	C.lua_pushinteger(s.L, C.lua_Integer(n))
}

func (s *State) PushNumber(f float64) {
	C.lua_pushnumber(s.L, C.lua_Number(f))
}

func (s *State) PushString(str string) {
	C.lua_pushlstring(s.L, s.strptr(str), C.size_t(len(str)))
}

func (s *State) Remove(index int) {
	C.lua_remove(s.L, C.int(index))
}

func (s *State) SetField(index int, key string) {
	C.lua_setfield(s.L, C.int(index), s.strptr(key))
}

func (s *State) SetGlobal(name string) {
	C.lua_setfield(s.L, C.LUA_GLOBALSINDEX, s.strptr(name))
}

func (s *State) SetTable(index int) {
	C.lua_settable(s.L, C.int(index))
}

func (s *State) ToString(index int) string {
	var size C.size_t
	r := C.lua_tolstring(s.L, C.int(index), &size)
	return C.GoStringN(r, C.int(size))
}

func (s *State) Free() {
	for _, ptr := range s.strptrs {
		C.free(unsafe.Pointer(ptr))
	}
}

func (s *State) push_string_map_string_as_table(m map[string]string) {
	s.NewTable()
	for k, v := range m {
		s.PushString(k)
		s.PushString(v)
		s.SetTable(-3)
	}
}

func (s *State) pushAsLuaString(item fmt.Stringer) {
	s.PushString(fmt.Sprintf("%s", item))
}

func (s *State) PushTable(t *Table) {
	s.NewTable()
	for i, j := range t.Int2int {
		s.PushInteger(i)
		s.PushInteger(j)
		s.SetTable(-3)
	}
	for i, b := range t.Int2bool {
		s.PushInteger(i)
		s.PushBoolean(b)
		s.SetTable(-3)
	}
	for i, str := range t.Int2string {
		s.PushInteger(i)
		s.PushString(str)
		s.SetTable(-3)
	}
	for i, f := range t.Int2float {
		s.PushInteger(i)
		s.PushNumber(f)
		s.SetTable(-3)
	}
	for i, t2 := range t.Int2table {
		s.PushInteger(i)
		s.PushTable(t2)
		s.SetTable(-3)
	}

	for str, i := range t.String2int {
		s.PushInteger(i)
		s.SetField(-2, str)
	}

	for str, b := range t.String2bool {
		s.PushBoolean(b)
		s.SetField(-2, str)
	}

	for str, str2 := range t.String2string {
		s.PushString(str2)
		s.SetField(-2, str)
	}

	for str, f := range t.String2float {
		s.PushNumber(f)
		s.SetField(-2, str)
	}

	for str, t2 := range t.String2table {
		s.PushTable(t2)
		s.SetField(-2, str)
	}
}

func (s *State) LoadScript(fname string) error {
	/*file, err := os.Open(fname)
	if err != nil {
		return err
	}*/
	fn := C.CString(fname)
	defer C.free(unsafe.Pointer(fn))
	if C.luaL_loadfile(s.L, fn) != 0 {
		return errors.New(s.ToString(-1))
	}
	return s.Call(0, 0)
}
