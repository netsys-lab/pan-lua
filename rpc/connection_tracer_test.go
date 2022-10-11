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

	"github.com/lucas-clemente/quic-go/logging"
	"github.com/netsec-ethz/scion-apps/pkg/pan"
)

func TestConnectionTracerMsgEncoding(t *testing.T) {
	b := bytes.Buffer{}
	enc := gob.NewEncoder(&b)
	msg := ConnectionTracerMsg{
		ID:             42,
		Local:          new(pan.UDPAddr),
		Remote:         new(pan.UDPAddr),
		SrcConnID:      new(logging.ConnectionID),
		DestConnID:     new(logging.ConnectionID),
		Chosen:         23,
		Versions:       []logging.VersionNumber{},
		ClientVersions: []logging.VersionNumber{},
		ServerVersions: []logging.VersionNumber{},
		ErrorMsg:       new(string),
		//Parameters:      &logging.TransportParameters{},
		ByteCount:       1337,
		Cwnd:            4711,
		CongestionState: 5,
		Packets:         9000,
		Header:          new(logging.Header),
		//ExtendedHeader:  &logging.ExtendedHeader{},
		Frames:          []logging.Frame{},
		AckFrame:        new(logging.AckFrame),
		PacketType:      13,
		DropReason:      99,
		LossReason:      100,
		EncryptionLevel: 128,
		PacketNumber:    1,
		TimerType:       88,
	}
	err := enc.Encode(msg)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+s", msg.String())
}
