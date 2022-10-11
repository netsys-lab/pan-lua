// Copyright 2021 Thorben Krüger (thorben.krueger@ovgu.de)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package lua

import (
	//"io/ioutil"
	"log"
	"os"
	"sync"

	lua "github.com/yuin/gopher-lua"
)

type State struct {
	*lua.LState
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

func (s *State) LoadScript(fname string) error {
	file, err := os.Open(fname)
	if err != nil {
		return err
	}
	if fn, err := s.Load(file, fname); err != nil {
		return err
	} else {
		s.Printf("loaded selector from file %s", fname)
		s.Push(fn)
		return s.PCall(0, lua.MultRet, nil)
	}
}
