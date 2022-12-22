// Copyright 2021 Thorben Krüger (thorben.krueger@ovgu.de)
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
	"log"
	"sync"

	"github.com/aarzilli/golua/lua"
)

type State struct {
	*lua.State
	sync.Mutex
	*log.Logger
}

func NewState() *State {
	L := lua.NewState()
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

func (s *State) LoadScript(fname string) error {
	/*file, err := os.Open(fname)
	if err != nil {
		return err
	}*/
	s.OpenLibs()
	if s.LoadFile(fname) != 0 {
		return errors.New(s.ToString(-1))
	}
	return s.Call(0, 0)
}
