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
	"bytes"
	"encoding/gob"
	"testing"

	"github.com/netsec-ethz/scion-apps/pkg/pan"
)

func TestSelectorMsgEncoding(t *testing.T) {
	b := bytes.Buffer{}
	enc := gob.NewEncoder(&b)
	msg := SelectorMsg{
		Local:         new(pan.UDPAddr),
		Remote:        new(pan.UDPAddr),
		Fingerprint:   new(pan.PathFingerprint),
		PathInterface: new(pan.PathInterface),
		Paths:         []*Path{},
	}
	err := enc.Encode(msg)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", msg)
}
