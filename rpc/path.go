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
package rpc

import (
	"time"

	"github.com/netsec-ethz/scion-apps/pkg/pan"
)

type Path struct {
	Source      pan.IA
	Destination pan.IA
	Metadata    *pan.PathMetadata
	Fingerprint pan.PathFingerprint
	//ForwardingPath pan.ForwardingPath
	Expiry time.Time
}

func (p *Path) PanPath() *pan.Path {
	return &pan.Path{
		Source:      p.Source,
		Destination: p.Destination,
		//Metadata:    p.Metadata.Copy(), // do we need this?
		Metadata:    p.Metadata,
		Fingerprint: p.Fingerprint,
		Expiry:      p.Expiry,
	}
}

func NewPathFrom(p *pan.Path) *Path {
	return &Path{
		Source:      p.Source,
		Destination: p.Destination,
		//Metadata:    p.Metadata.Copy(),
		Metadata:    p.Metadata,
		Fingerprint: p.Fingerprint,
		//ForwardingPath: p.ForwardingPath,
		Expiry: p.Expiry,
	}
}
