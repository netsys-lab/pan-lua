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
)

type SelectorMsg struct {
	Local         *pan.UDPAddr
	Remote        *pan.UDPAddr
	Fingerprint   *pan.PathFingerprint
	PathInterface *pan.PathInterface
	Preferences   map[string]string
	Paths         []*Path
}

// SelectorServer is the RPC-facing server part (the one with the rigit function signatures)
type SelectorServer struct {
	selector ServerSelector
}

/*func NewSelectorServer(selector ServerSelector) (*rpc.Server, error) {
	err := rpc.Register(&SelectorServer{selector})
	if err != nil {
		return nil, err
	}
	return rpc.DefaultServer, nil
}

func (s *SelectorServer) RegisterAt(r *rpc.Server) error {
	return r.Register(s)
        }*/

func NewSelectorServer(selector ServerSelector) *SelectorServer {
	return &SelectorServer{selector}
}

func (s *SelectorServer) Initialize(args, resp *SelectorMsg) error {
	log.Println("Initialize invoked")
	paths := make([]*pan.Path, len(args.Paths))
	for i, p := range args.Paths {
		paths[i] = p.PanPath()
		//log.Printf("%s", paths[i].Source)
	}
	if args.Local == nil || args.Remote == nil {
		return ErrDeref
	}
	return s.selector.Initialize(args.Preferences, *args.Local, *args.Remote, paths)
}

func (s *SelectorServer) SetPreferences(args, resp *SelectorMsg) error {
	if args.Local == nil || args.Remote == nil {
		return ErrDeref
	}

	return s.selector.SetPreferences(args.Preferences, *args.Local, *args.Remote)
}

func (s *SelectorServer) Path(args, resp *SelectorMsg) error {
	if args.Remote == nil {
		return ErrDeref
	}
	p, err := s.selector.Path(*args.Local, *args.Remote)
	if p != nil {
		resp.Fingerprint = &p.Fingerprint
	}
	return err
}

func (s *SelectorServer) PathDown(args, resp *SelectorMsg) error {
	//log.Println("PathDown called")
	if args.Remote == nil || args.Fingerprint == nil || args.PathInterface == nil {
		return ErrDeref
	}
	return s.selector.PathDown(*args.Local, *args.Remote, *args.Fingerprint, *args.PathInterface)
}

func (s *SelectorServer) Refresh(args, resp *SelectorMsg) error {
	//log.Println("Refresh invoked")
	paths := make([]*pan.Path, len(args.Paths))
	for i, p := range args.Paths {
		paths[i] = p.PanPath()
		//log.Printf("%s", paths[i].Source)
	}
	if args.Remote == nil {
		return ErrDeref
	}
	return s.selector.Refresh(*args.Local, *args.Remote, paths)
}

func (s *SelectorServer) Close(args, resp *SelectorMsg) error {
	//log.Println("Close called")
	if args.Remote == nil {
		return ErrDeref
	}
	return s.selector.Close(*args.Local, *args.Remote)
}
