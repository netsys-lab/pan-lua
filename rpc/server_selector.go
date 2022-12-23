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
package rpc

import (
	"errors"
	"net"

	"github.com/netsec-ethz/scion-apps/pkg/pan"
	"github.com/netsys-lab/pan-lua/selector"
)

var (
	DefaultDaemonAddress = &net.UnixAddr{
		Name: "/tmp/scion-pan-rpc.sock",
		Net:  "unix",
	}
	ErrDeref = errors.New("Can not dereference Nil value")
)

type ServerSelector interface {
	Initialize(map[string]string, pan.UDPAddr, pan.UDPAddr, []*pan.Path) error
	SetPreferences(map[string]string, pan.UDPAddr, pan.UDPAddr) error
	Path(pan.UDPAddr, pan.UDPAddr) (*pan.Path, error)
	PathDown(pan.UDPAddr, pan.UDPAddr, pan.PathFingerprint, pan.PathInterface) error
	Refresh(pan.UDPAddr, pan.UDPAddr, []*pan.Path) error
	Close(pan.UDPAddr, pan.UDPAddr) error
}

type serverSelector struct {
	fn        func(pan.UDPAddr, pan.UDPAddr) selector.Selector
	selectors map[string]selector.Selector
}

func NewServerSelectorFunc(fn func(pan.UDPAddr, pan.UDPAddr) selector.Selector) ServerSelector {
	return serverSelector{fn, map[string]selector.Selector{}}
}

func (s *serverSelector) getSelector(local, remote pan.UDPAddr) selector.Selector {
	addr := local.String() + remote.String()
	selector, ok := s.selectors[addr]
	if !ok {
		selector = s.fn(local, remote)
		s.selectors[addr] = selector
	}
	return selector
}

func (s serverSelector) Initialize(prefs map[string]string, local, remote pan.UDPAddr, paths []*pan.Path) error {
	s.getSelector(local, remote).Initialize(local, remote, paths)
	return nil
}

func (s serverSelector) SetPreferences(prefs map[string]string, local, remote pan.UDPAddr) error {
	return s.getSelector(local, remote).SetPreferences(prefs)
}

func (s serverSelector) Path(local, remote pan.UDPAddr) (*pan.Path, error) {
	return s.getSelector(local, remote).Path(), nil
}

func (s serverSelector) PathDown(local, remote pan.UDPAddr, fp pan.PathFingerprint, pi pan.PathInterface) error {
	s.getSelector(local, remote).PathDown(fp, pi)
	return nil
}

func (s serverSelector) Refresh(local, remote pan.UDPAddr, paths []*pan.Path) error {
	s.getSelector(local, remote).Refresh(paths)
	return nil
}

func (s serverSelector) Close(local, remote pan.UDPAddr) error {
	err := s.getSelector(local, remote).Close()
	delete(s.selectors, local.String()+remote.String())
	return err
}
