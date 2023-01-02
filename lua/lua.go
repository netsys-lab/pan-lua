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
package lua

import (
	//"io/ioutil"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/aarzilli/golua/lua"
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
	*lua.State
	sync.Mutex
	*log.Logger
}

func NewState() *State {
	L := lua.NewState()
	L.OpenLibs()
	//l := log.New(ioutil.Discard, "lua ", log.Ltime)
	l := log.Default()
	l.SetFlags(log.Ltime | log.Lshortfile)
	l.SetPrefix("lua ")
	return &State{L, sync.Mutex{}, l}
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
	if s.LoadFile(fname) != 0 {
		return errors.New(s.ToString(-1))
	}
	return s.Call(0, 0)
}
