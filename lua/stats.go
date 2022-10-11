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
package lua

import (
	"fmt"
	"time"

	"github.com/lucas-clemente/quic-go/logging"
	"github.com/netsec-ethz/scion-apps/pkg/pan"
	"github.com/netsys-lab/pan-lua/rpc"
	lua "github.com/yuin/gopher-lua"
)

func strhlpr(s fmt.Stringer) lua.LString {
	return lua.LString(fmt.Sprintf("%s", s))
}

func new_lua_parameters(p *logging.TransportParameters) *lua.LTable {
	t := new(lua.LTable)
	if p != nil {
		t.RawSetString("InitialMaxStreamDataBidiLocal", lua.LNumber(p.InitialMaxStreamDataBidiLocal))
		t.RawSetString("InitialMaxStreamDataBidiRemote", lua.LNumber(p.InitialMaxStreamDataBidiRemote))
		t.RawSetString("InitialMaxStreamDataUni", lua.LNumber(p.InitialMaxStreamDataUni))
		t.RawSetString("InitialMaxData", lua.LNumber(p.InitialMaxData))
		t.RawSetString("MaxAckDelay", lua.LNumber(p.MaxAckDelay))
		t.RawSetString("AckDelayExponent", lua.LNumber(p.AckDelayExponent))
		t.RawSetString("DisableActiveMigration", lua.LBool(p.DisableActiveMigration))
		t.RawSetString("MaxUDPPayloadSize", lua.LNumber(p.MaxUDPPayloadSize))
		t.RawSetString("MaxUniStreamNum", lua.LNumber(p.MaxUniStreamNum))
		t.RawSetString("MaxBidiStreamNum", lua.LNumber(p.MaxBidiStreamNum))
		t.RawSetString("MaxIdleTimeout", lua.LNumber(p.MaxIdleTimeout))
		a := p.PreferredAddress
		if a != nil {
			t.RawSetString("PreferredAddress", lua.LString(
				fmt.Sprintf(
					"IPv4: %s:%d, IPv6: %s:%d, ConnectionID: %s, Token: %x",
					a.IPv4, a.IPv4Port, a.IPv6, a.IPv6Port, a.ConnectionID, a.StatelessResetToken,
				),
			))
		}
		t.RawSetString("OriginalDestinationConnectionID", strhlpr(p.OriginalDestinationConnectionID))
		t.RawSetString("InitialSourceConnectionID", strhlpr(p.InitialSourceConnectionID))
		t.RawSetString("RetrySourceConnectionID", strhlpr(p.RetrySourceConnectionID))
		t.RawSetString("StatelessResetToken", lua.LString(fmt.Sprintf("%x", p.StatelessResetToken)))
		t.RawSetString("ActiveConnectionIDLimit", lua.LNumber(p.ActiveConnectionIDLimit))
		t.RawSetString("MaxDatagramFrameSize", lua.LNumber(p.MaxDatagramFrameSize))
		return t
	}
	return nil
}

func new_lua_rtt_stats(stats *rpc.RTTStats) *lua.LTable {
	t := new(lua.LTable)
	if stats != nil {
		t.RawSetString("LatestRTT", lua.LNumber(stats.LatestRTT.Seconds()))
		t.RawSetString("MaxAckDelay", lua.LNumber(stats.MaxAckDelay.Seconds()))
		t.RawSetString("MeanDeviation", lua.LNumber(stats.MeanDeviation.Seconds()))
		t.RawSetString("MinRTT", lua.LNumber(stats.MinRTT.Seconds()))
		t.RawSetString("PTO", lua.LNumber(stats.PTO.Seconds()))
		t.RawSetString("SmoothedRTT", lua.LNumber(stats.SmoothedRTT.Seconds()))
		return t
	}
	return nil
}

type Stats struct {
	*State
	mod *lua.LTable
}

func NewStats(state *State) rpc.ServerConnectionTracer {
	state.Lock()
	defer state.Unlock()
	mod := map[string]lua.LGFunction{}
	for _, fn := range []string{
		"TracerForConnection",
		"StartedConnection",
		"NegotiatedVersion",
		"ClosedConnection",
		"SentTransportParameters",
		"ReceivedTransportParameters",
		"RestoredTransportParameters",
		"SentPacket",
		"ReceivedVersionNegotiationPacket",
		"ReceivedRetry",
		"ReceivedPacket",
		"BufferedPacket",
		"DroppedPacket",
		"UpdatedMetrics",
		"AcknowledgedPacket",
		"LostPacket",
		"UpdatedCongestionState",
		"UpdatedPTOCount",
		"UpdatedKeyFromTLS",
		"UpdatedKey",
		"DroppedEncryptionLevel",
		"DroppedKey",
		"SetLossTimer",
		"LossTimerExpired",
		"LossTimerCanceled",
		"Debug",
	} {
		//s := fmt.Sprintf("function %s not implemented in script", fn)
		mod[fn] = func(L *lua.LState) int {
			//state.Logger.Println(s)
			return 0
		}
	}

	stats := state.RegisterModule("stats", mod).(*lua.LTable)
	return &Stats{state, stats}
}

func (s *Stats) TracerForConnection(tracer_id uint64, p logging.Perspective, odcid logging.ConnectionID) error {
	//s.Printf("TracerForConnection")
	s.Lock()
	defer s.Unlock()
	return s.CallByParam(
		lua.P{
			Fn:      s.mod.RawGetString("TracerForConnection"),
			NRet:    0,
			Protect: true,
		},
		lua.LNumber(tracer_id),
		lua.LNumber(p),
		strhlpr(odcid),
	)
}
func (s *Stats) StartedConnection(local, remote *pan.UDPAddr, srcConnID, destConnID logging.ConnectionID) error {
	//s.Printf("StartedConnection")
	s.Lock()
	defer s.Unlock()
	return s.CallByParam(
		lua.P{
			Fn:      s.mod.RawGetString("StartedConnection"),
			NRet:    0,
			Protect: true,
		},
		strhlpr(local),
		strhlpr(remote),
		strhlpr(srcConnID),
		strhlpr(destConnID),
	)

}
func (s *Stats) NegotiatedVersion(local, remote *pan.UDPAddr, chosen logging.VersionNumber, clientVersions, serverVersions []logging.VersionNumber) error {
	//s.Printf("NegotiatedVersion")
	s.Lock()
	defer s.Unlock()
	var (
		c_vs = lua.LTable{}
		s_vs = lua.LTable{}
	)
	for _, v := range clientVersions {
		c_vs.Append(strhlpr(v))
	}
	for _, v := range serverVersions {
		s_vs.Append(strhlpr(v))
	}

	return s.CallByParam(
		lua.P{
			Fn:      s.mod.RawGetString("NegotiatedVersion"),
			NRet:    0,
			Protect: true,
		},
		strhlpr(local), strhlpr(remote),
		strhlpr(chosen),
		&c_vs,
		&s_vs,
	)

}
func (s *Stats) ClosedConnection(local, remote *pan.UDPAddr, err error) error {
	//s.Printf("ClosedConnection")
	s.Lock()
	defer s.Unlock()
	return s.CallByParam(
		lua.P{
			Fn:      s.mod.RawGetString("ClosedConnection"),
			NRet:    0,
			Protect: true,
		},
		strhlpr(local), strhlpr(remote),
		lua.LString(err.Error()),
	)

}
func (s *Stats) SentTransportParameters(local, remote *pan.UDPAddr, parameters *logging.TransportParameters) error {
	//s.Printf("SentTransportParameters")
	s.Lock()
	defer s.Unlock()
	return s.CallByParam(
		lua.P{
			Fn:      s.mod.RawGetString("SentTransportParameters"),
			NRet:    0,
			Protect: true,
		},
		strhlpr(local), strhlpr(remote),
		new_lua_parameters(parameters),
	)

}
func (s *Stats) ReceivedTransportParameters(local, remote *pan.UDPAddr, parameters *logging.TransportParameters) error {
	//s.Printf("ReceivedTransportParameters")
	s.Lock()
	defer s.Unlock()
	return s.CallByParam(
		lua.P{
			Fn:      s.mod.RawGetString("ReceivedTransportParameters"),
			NRet:    0,
			Protect: true,
		},
		strhlpr(local), strhlpr(remote),
		new_lua_parameters(parameters),
	)

}
func (s *Stats) RestoredTransportParameters(local, remote *pan.UDPAddr, parameters *logging.TransportParameters) error {
	//s.Printf("RestoredTransportParameters")
	s.Lock()
	defer s.Unlock()
	return s.CallByParam(
		lua.P{
			Fn:      s.mod.RawGetString("RestoredTransportParameters"),
			NRet:    0,
			Protect: true,
		},
		strhlpr(local), strhlpr(remote),
		new_lua_parameters(parameters),
	)

}
func (s *Stats) SentPacket(local, remote *pan.UDPAddr, hdr *logging.ExtendedHeader, size logging.ByteCount, ack *logging.AckFrame, frames []logging.Frame) error {
	//s.Printf("SentPacket: only stub implementation")
	s.Lock()
	defer s.Unlock()
	return s.CallByParam(
		lua.P{
			Fn:      s.mod.RawGetString("SentPacket"),
			NRet:    0,
			Protect: true,
		},
		strhlpr(local), strhlpr(remote), lua.LNumber(size),
	)

}
func (s *Stats) ReceivedVersionNegotiationPacket(local, remote *pan.UDPAddr, hdr *logging.Header, versions []logging.VersionNumber) error {
	//s.Printf("ReceivedVersionNegotiationPacket: only stub implementation")
	s.Lock()
	defer s.Unlock()
	var vs = new(lua.LTable)
	for _, v := range versions {
		vs.Append(strhlpr(v))
	}

	return s.CallByParam(
		lua.P{
			Fn:      s.mod.RawGetString("ReceivedVersionNegotiationPacket"),
			NRet:    0,
			Protect: true,
		},
		strhlpr(local), strhlpr(remote),
		vs,
	)

}
func (s *Stats) ReceivedRetry(local, remote *pan.UDPAddr, hdr *logging.Header) error {
	//s.Printf("ReceivedRetry: only stub implementation")
	s.Lock()
	defer s.Unlock()
	return s.CallByParam(
		lua.P{
			Fn:      s.mod.RawGetString("ReceivedRetry"),
			NRet:    0,
			Protect: true,
		},
		strhlpr(local), strhlpr(remote),
	)

}
func (s *Stats) ReceivedPacket(local, remote *pan.UDPAddr, hdr *logging.ExtendedHeader, size logging.ByteCount, frames []logging.Frame) error {
	//s.Printf("ReceivedPacket: only stub implementation")
	s.Lock()
	defer s.Unlock()
	return s.CallByParam(
		lua.P{
			Fn:      s.mod.RawGetString("ReceivedPacket"),
			NRet:    0,
			Protect: true,
		},
		strhlpr(local), strhlpr(remote),
	)

}
func (s *Stats) BufferedPacket(local, remote *pan.UDPAddr, ptype logging.PacketType) error {
	//s.Printf("BufferedPacket")
	s.Lock()
	defer s.Unlock()
	return s.CallByParam(
		lua.P{
			Fn:      s.mod.RawGetString("BufferedPacket"),
			NRet:    0,
			Protect: true,
		},
		strhlpr(local), strhlpr(remote),
		lua.LNumber(ptype),
	)

}
func (s *Stats) DroppedPacket(local, remote *pan.UDPAddr, ptype logging.PacketType, size logging.ByteCount, reason logging.PacketDropReason) error {
	//s.Printf("DroppedPacket")
	s.Lock()
	defer s.Unlock()
	return s.CallByParam(
		lua.P{
			Fn:      s.mod.RawGetString("DroppedPacket"),
			NRet:    0,
			Protect: true,
		},
		strhlpr(local), strhlpr(remote),
		lua.LNumber(ptype),
		lua.LNumber(size),
		lua.LNumber(reason),
	)

}
func (s *Stats) UpdatedMetrics(local, remote *pan.UDPAddr, rttStats *rpc.RTTStats, cwnd, bytesInFlight logging.ByteCount, packetsInFlight int) error {
	//s.Printf("UpdatedMetrics")
	s.Lock()
	defer s.Unlock()
	return s.CallByParam(
		lua.P{
			Fn:      s.mod.RawGetString("UpdatedMetrics"),
			NRet:    0,
			Protect: true,
		},
		strhlpr(local), strhlpr(remote),
		new_lua_rtt_stats(rttStats),
		lua.LNumber(cwnd),
		lua.LNumber(bytesInFlight),
		lua.LNumber(packetsInFlight),
	)

}
func (s *Stats) AcknowledgedPacket(local, remote *pan.UDPAddr, level logging.EncryptionLevel, num logging.PacketNumber) error {
	//s.Printf("AcknowledgedPacket")
	s.Lock()
	defer s.Unlock()
	return s.CallByParam(
		lua.P{
			Fn:      s.mod.RawGetString("AcknowledgedPacket"),
			NRet:    0,
			Protect: true,
		},
		strhlpr(local), strhlpr(remote),
		strhlpr(level),
		lua.LNumber(num),
	)

}
func (s *Stats) LostPacket(local, remote *pan.UDPAddr, level logging.EncryptionLevel, num logging.PacketNumber, reason logging.PacketLossReason) error {
	//s.Printf("LostPacket")
	s.Lock()
	defer s.Unlock()
	return s.CallByParam(
		lua.P{
			Fn:      s.mod.RawGetString("LostPacket"),
			NRet:    0,
			Protect: true,
		},
		strhlpr(local), strhlpr(remote),
		strhlpr(level),
		lua.LNumber(num),
		lua.LNumber(reason),
	)

}
func (s *Stats) UpdatedCongestionState(local, remote *pan.UDPAddr, state logging.CongestionState) error {
	//s.Printf("UpdatedCongestionState")
	s.Lock()
	defer s.Unlock()
	return s.CallByParam(
		lua.P{
			Fn:      s.mod.RawGetString("UpdatedCongestionState"),
			NRet:    0,
			Protect: true,
		},
		strhlpr(local), strhlpr(remote),
		lua.LNumber(state),
	)

}
func (s *Stats) UpdatedPTOCount(local, remote *pan.UDPAddr, value uint32) error {
	//s.Printf("UpdatedPTOCount")
	s.Lock()
	defer s.Unlock()
	return s.CallByParam(
		lua.P{
			Fn:      s.mod.RawGetString("UpdatedPTOCount"),
			NRet:    0,
			Protect: true,
		},
		strhlpr(local), strhlpr(remote),
		lua.LNumber(value),
	)

}
func (s *Stats) UpdatedKeyFromTLS(local, remote *pan.UDPAddr, level logging.EncryptionLevel, p logging.Perspective) error {
	//s.Printf("UpdatedKeyFromTLS")
	s.Lock()
	defer s.Unlock()
	return s.CallByParam(
		lua.P{
			Fn:      s.mod.RawGetString("UpdatedKeyFromTLS"),
			NRet:    0,
			Protect: true,
		},
		strhlpr(local), strhlpr(remote),
		strhlpr(level),
		lua.LNumber(p),
	)

}
func (s *Stats) UpdatedKey(local, remote *pan.UDPAddr, generation logging.KeyPhase, rmte bool) error {
	//s.Printf("UpdatedKey")
	s.Lock()
	defer s.Unlock()
	return s.CallByParam(
		lua.P{
			Fn:      s.mod.RawGetString("UpdatedKey"),
			NRet:    0,
			Protect: true,
		},
		strhlpr(local), strhlpr(remote),
		lua.LNumber(generation),
		lua.LBool(rmte),
	)

}
func (s *Stats) DroppedEncryptionLevel(local, remote *pan.UDPAddr, level logging.EncryptionLevel) error {
	//s.Printf("DroppedEncryptionLevel")
	s.Lock()
	defer s.Unlock()
	return s.CallByParam(
		lua.P{
			Fn:      s.mod.RawGetString("DroppedEncryptionLevel"),
			NRet:    0,
			Protect: true,
		},
		strhlpr(local), strhlpr(remote),
		strhlpr(level),
	)

}
func (s *Stats) DroppedKey(local, remote *pan.UDPAddr, generation logging.KeyPhase) error {
	//s.Printf("DroppedKey")
	s.Lock()
	defer s.Unlock()
	return s.CallByParam(
		lua.P{
			Fn:      s.mod.RawGetString("DroppedKey"),
			NRet:    0,
			Protect: true,
		},
		strhlpr(local), strhlpr(remote),
		lua.LNumber(generation),
	)

}
func (s *Stats) SetLossTimer(local, remote *pan.UDPAddr, ttype logging.TimerType, level logging.EncryptionLevel, t time.Time) error {
	//s.Printf("SetLossTimer")
	s.Lock()
	defer s.Unlock()
	return s.CallByParam(
		lua.P{
			Fn:      s.mod.RawGetString("SetLossTimer"),
			NRet:    0,
			Protect: true,
		},
		strhlpr(local), strhlpr(remote),
		lua.LNumber(ttype),
		strhlpr(level),
		strhlpr(t),
	)

}
func (s *Stats) LossTimerExpired(local, remote *pan.UDPAddr, ttype logging.TimerType, level logging.EncryptionLevel) error {
	//s.Printf("LossTimerExpired")
	s.Lock()
	defer s.Unlock()
	return s.CallByParam(
		lua.P{
			Fn:      s.mod.RawGetString("LossTimerExpired"),
			NRet:    0,
			Protect: true,
		},
		strhlpr(local), strhlpr(remote),
		lua.LNumber(ttype),
		strhlpr(level),
	)

}
func (s *Stats) LossTimerCanceled(local, remote *pan.UDPAddr) error {
	//s.Printf("LossTimerCanceled")
	s.Lock()
	defer s.Unlock()
	return s.CallByParam(
		lua.P{
			Fn:      s.mod.RawGetString("LossTimerCanceled"),
			NRet:    0,
			Protect: true,
		},
		strhlpr(local), strhlpr(remote),
	)

}
func (s *Stats) Close(local, remote *pan.UDPAddr) error {
	//s.Printf("Close")
	s.Lock()
	defer s.Unlock()
	return s.CallByParam(
		lua.P{
			Fn:      s.mod.RawGetString("Close"),
			NRet:    0,
			Protect: true,
		},
		strhlpr(local), strhlpr(remote),
	)

}
func (s *Stats) Debug(local, remote *pan.UDPAddr, name, msg string) error {
	//s.Printf("Debug")
	s.Lock()
	defer s.Unlock()
	return s.CallByParam(
		lua.P{
			Fn:      s.mod.RawGetString("Debug"),
			NRet:    0,
			Protect: true,
		},
		strhlpr(local), strhlpr(remote),
		lua.LString(name),
		lua.LString(msg),
	)

}
