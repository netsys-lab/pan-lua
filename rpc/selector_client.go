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
package rpc

import (
	"log"

	"github.com/netsec-ethz/scion-apps/pkg/pan"
	"github.com/netsys-lab/pan-lua/selector"
)

type SelectorClient struct {
	connectionPreferences map[string]string
	client                *Client
	paths                 map[pan.PathFingerprint]*pan.Path
	last                  *pan.Path
	local                 *pan.UDPAddr
	remote                *pan.UDPAddr
	l                     *log.Logger
}

func NewSelectorClient(client *Client) selector.Selector {
	client.l.Printf("RPC connection etablished")
	return &SelectorClient{map[string]string{}, client, map[pan.PathFingerprint]*pan.Path{}, nil, nil, nil, client.l}
}

// LastPath implements selector.Selector
func (s *SelectorClient) LastPath() *pan.Path {
	return s.last
}

func (s *SelectorClient) Initialize(local, remote pan.UDPAddr, paths []*pan.Path) {
	s.l.Println("Initialize called")
	s.remote = &remote
	s.local = &local
	ps := make([]*Path, len(paths))
	for i, p := range paths {
		s.paths[p.Fingerprint] = p
		ps[i] = NewPathFrom(p)
	}
	err := s.client.Call("SelectorServer.Initialize", &SelectorMsg{
		Local:  s.local,
		Remote: s.remote,
		Paths:  ps,
	}, &SelectorMsg{})
	if err != nil {
		s.l.Fatalln(err)
	}
	s.l.Printf("Initialize returned")
}

func (s *SelectorClient) SetPreferences(prefs map[string]string) error {
	s.l.Println("SetPreferences called")
	s.connectionPreferences = prefs
	if s.local != nil && s.remote != nil {
		return s.client.Call("SelectorServer.SetPreferences", &SelectorMsg{
			Local:       s.local,
			Remote:      s.remote,
			Preferences: s.connectionPreferences,
		}, &SelectorMsg{})
	} else {
		//we don't know the connection-identifying local and remote addresses yet
		//so we wait until "Initialize" gets called naturally
		s.l.Println("local and remote addresses not yet known, doing nothing for now")
		return nil
	}
}
func (s *SelectorClient) Path() *pan.Path {
	//s.l.Println("Path called")
	msg := SelectorMsg{}
	err := s.client.Call("SelectorServer.Path", &SelectorMsg{
		Local:  s.local,
		Remote: s.remote,
	}, &msg)
	if err != nil {
		s.l.Fatalln(err)
	}
	if msg.Fingerprint != nil {
		s.last = s.paths[*msg.Fingerprint]
		return s.last
	}
	return nil
}

func (s *SelectorClient) PathDown(fp pan.PathFingerprint, pi pan.PathInterface) {
	s.l.Println("PathDown called")
	s.paths[fp] = nil // remove from local table
	err := s.client.Call("SelectorServer.PathDown", &SelectorMsg{
		Local:         s.local,
		Remote:        s.remote,
		Fingerprint:   &fp,
		PathInterface: &pi,
	}, &SelectorMsg{})
	if err != nil {
		s.l.Fatalln(err)
	}

}

func (s *SelectorClient) Refresh(paths []*pan.Path) {
	s.l.Println("Refresh called")
	ps := make([]*Path, len(paths))
	for i, p := range paths {
		s.paths[p.Fingerprint] = p
		ps[i] = NewPathFrom(p)
	}
	err := s.client.Call("SelectorServer.Refresh", &SelectorMsg{
		Local:  s.local,
		Remote: s.remote,
		Paths:  ps,
	}, &SelectorMsg{})
	if err != nil {
		s.l.Fatalln(err)
	}
	s.l.Printf("Refresh returned")
}

func (s *SelectorClient) Close() error {
	s.l.Println("Close called")
	err := s.client.Call("SelectorServer.Close", &SelectorMsg{Local: s.local, Remote: s.remote}, &SelectorMsg{})
	if err != nil {
		s.l.Println(err)
		s.l.Println(s.client.client.Close())
		return err
	}
	return s.client.client.Close()
}
