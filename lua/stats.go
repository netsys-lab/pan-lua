// Copyright 2021,2022,2023 Thorben Krüger (thorben.krueger@ovgu.de)
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
	"log"
	"time"

	"github.com/aarzilli/golua/lua"
	"github.com/lucas-clemente/quic-go/logging"
	"github.com/netsec-ethz/scion-apps/pkg/pan"
	"github.com/netsys-lab/pan-lua/rpc"
)

var statsName string = "stats"

func new_lua_parameters(p *logging.TransportParameters) *Table {
	t := NewTable()
	if p != nil {
		t.String2int["InitialMaxStreamDataBidiLocal"] = int64(p.InitialMaxStreamDataBidiLocal)
		t.String2int["InitialMaxStreamDataBidiRemote"] = int64(p.InitialMaxStreamDataBidiRemote)
		t.String2int["InitialMaxStreamDataUni"] = int64(p.InitialMaxStreamDataUni)
		t.String2int["InitialMaxData"] = int64(p.InitialMaxData)
		t.String2int["MaxAckDelay"] = int64(p.MaxAckDelay)
		t.String2int["AckDelayExponent"] = int64(p.AckDelayExponent)
		t.String2bool["DisableActiveMigration"] = p.DisableActiveMigration
		t.String2int["MaxUDPPayloadSize"] = int64(p.MaxUDPPayloadSize)
		t.String2int["MaxUniStreamNum"] = int64(p.MaxUniStreamNum)
		t.String2int["MaxBidiStreamNum"] = int64(p.MaxBidiStreamNum)
		t.String2int["MaxIdleTimeout"] = int64(p.MaxIdleTimeout)
		a := p.PreferredAddress
		if a != nil {
			t.String2string["PreferredAddress"] = fmt.Sprintf(
				"IPv4: %s:%d, IPv6: %s:%d, ConnectionID: %s, Token: %x",
				a.IPv4, a.IPv4Port, a.IPv6, a.IPv6Port, a.ConnectionID, a.StatelessResetToken,
			)
		}
		t.String2string["OriginalDestinationConnectionID"] = p.OriginalDestinationConnectionID.String()
		t.String2string["InitialSourceConnectionID"] = p.InitialSourceConnectionID.String()
		if id := p.RetrySourceConnectionID; id != nil {
			t.String2string["RetrySourceConnectionID"] = id.String()
		}
		if token := p.StatelessResetToken; token != nil {
			t.String2string["StatelessResetToken"] = fmt.Sprintf("%x", token)
		}
		t.String2int["ActiveConnectionIDLimit"] = int64(p.ActiveConnectionIDLimit)
		t.String2int["MaxDatagramFrameSize"] = int64(p.MaxDatagramFrameSize)
	}
	return t
}

func new_lua_rtt_stats(stats *rpc.RTTStats) *Table {
	t := NewTable()
	if stats != nil {
		t.String2float["LatestRTT"] = float64(stats.LatestRTT.Seconds())
		t.String2float["MaxAckDelay"] = float64(stats.MaxAckDelay.Seconds())
		t.String2float["MeanDeviation"] = float64(stats.MeanDeviation.Seconds())
		t.String2float["MinRTT"] = float64(stats.MinRTT.Seconds())
		t.String2float["PTO"] = float64(stats.PTO.Seconds())
		t.String2float["SmoothedRTT"] = float64(stats.SmoothedRTT.Seconds())
	}
	return t
}

type Stats struct {
	*State
}

func NewStats(state *State) rpc.ServerConnectionTracer {
	state.Lock()
	defer state.Unlock()

	state.NewTable()

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
		s := fmt.Sprintf("function %s not implemented in script", fn)
		state.PushGoFunction(func(L *lua.State) int {
			state.GetGlobal(tableName)
			state.GetField(-1, "Log")
			state.PushString(s)
			state.Call(1, 0)
			state.Pop(1)
			return 0
		})
		state.SetField(-2, fn)
	}

	state.SetGlobal(statsName)

	return &Stats{state}
}

func (s *Stats) TracerForConnection(tracer_id uint64, p logging.Perspective, odcid logging.ConnectionID) error {
	//s.Printf("TracerForConnection")
	s.Lock()
	defer s.Unlock()
	s.GetGlobal(statsName)
	s.GetField(-1, "TracerForConnection")
	s.Remove(-2)
	s.PushInteger(int64(tracer_id))
	s.PushInteger(int64(p))
	s.pushAsLuaString(odcid)
	err := s.Call(3, 0)
	if err != nil {
		log.Println(err)
	}
	return err
}

func (s *Stats) StartedConnection(local, remote *pan.UDPAddr, srcConnID, destConnID logging.ConnectionID) error {
	//s.Printf("StartedConnection")
	s.Lock()
	defer s.Unlock()

	s.GetGlobal(statsName)
	s.GetField(-1, "StartedConnection")
	s.Remove(-2)
	s.pushAsLuaString(local)
	s.pushAsLuaString(remote)
	s.pushAsLuaString(srcConnID)
	s.pushAsLuaString(destConnID)
	err := s.Call(4, 0)
	if err != nil {
		log.Println(err)
	}

	return err
}

func (s *Stats) NegotiatedVersion(local, remote *pan.UDPAddr, chosen logging.VersionNumber, clientVersions, serverVersions []logging.VersionNumber) error {
	//s.Printf("NegotiatedVersion")

	c_vs := NewTable()
	s_vs := NewTable()

	for i, v := range clientVersions {
		c_vs.Int2int[int64(i+1)] = int64(v)
	}
	for i, v := range serverVersions {
		s_vs.Int2int[int64(i+1)] = int64(v)
	}

	s.Lock()
	defer s.Unlock()
	s.GetGlobal(statsName)
	s.GetField(-1, "NegotiatedVersion")
	s.Remove(-2)
	s.pushAsLuaString(local)
	s.pushAsLuaString(remote)
	s.PushInteger(int64(chosen))
	s.PushTable(c_vs)
	s.PushTable(s_vs)
	err := s.Call(5, 0)
	if err != nil {
		log.Println(err)
	}

	return err
}

func (s *Stats) ClosedConnection(local, remote *pan.UDPAddr, err error) error {
	//s.Printf("ClosedConnection")
	s.Lock()
	defer s.Unlock()
	s.GetGlobal(statsName)
	s.GetField(-1, "ClosedConnection")
	s.Remove(-2)
	s.pushAsLuaString(local)
	s.pushAsLuaString(remote)
	s.PushString(err.Error())
	err2 := s.Call(3, 0)
	if err2 != nil {
		log.Println(err2)
	}

	return err2
}

func (s *Stats) SentTransportParameters(local, remote *pan.UDPAddr, parameters *logging.TransportParameters) error {
	//s.Printf("SentTransportParameters")
	s.Lock()
	defer s.Unlock()
	s.GetGlobal(statsName)
	s.GetField(-1, "SentTransportParameters")
	s.Remove(-2)
	s.pushAsLuaString(local)
	s.pushAsLuaString(remote)
	s.PushTable(new_lua_parameters(parameters))
	err := s.Call(3, 0)
	if err != nil {
		log.Println(err)
	}

	return err
}

func (s *Stats) ReceivedTransportParameters(local, remote *pan.UDPAddr, parameters *logging.TransportParameters) error {
	//s.Printf("ReceivedTransportParameters")
	s.Lock()
	defer s.Unlock()
	s.GetGlobal(statsName)
	s.GetField(-1, "ReceivedTransportParameters")
	s.Remove(-2)
	s.pushAsLuaString(local)
	s.pushAsLuaString(remote)
	s.PushTable(new_lua_parameters(parameters))
	err := s.Call(3, 0)
	if err != nil {
		log.Println(err)
	}

	return err
}

func (s *Stats) RestoredTransportParameters(local, remote *pan.UDPAddr, parameters *logging.TransportParameters) error {
	//s.Printf("RestoredTransportParameters")
	s.Lock()
	defer s.Unlock()
	s.GetGlobal(statsName)
	s.GetField(-1, "RestoredTransportParameters")
	s.Remove(-2)
	s.pushAsLuaString(local)
	s.pushAsLuaString(remote)
	s.PushTable(new_lua_parameters(parameters))
	err := s.Call(3, 0)
	if err != nil {
		log.Println(err)
	}

	return err
}

func (s *Stats) SentPacket(local, remote *pan.UDPAddr, hdr *logging.ExtendedHeader, size logging.ByteCount, ack *logging.AckFrame, frames []logging.Frame) error {
	//s.Printf("SentPacket: only stub implementation")
	s.Lock()
	defer s.Unlock()
	s.GetGlobal(statsName)
	s.GetField(-1, "SentPacket")
	s.Remove(-2)
	s.pushAsLuaString(local)
	s.pushAsLuaString(remote)
	s.PushInteger(int64(size))
	err := s.Call(3, 0)
	if err != nil {
		log.Println(err)
	}

	return err
}

func (s *Stats) ReceivedVersionNegotiationPacket(local, remote *pan.UDPAddr, hdr *logging.Header, versions []logging.VersionNumber) error {
	//s.Printf("ReceivedVersionNegotiationPacket: only stub implementation")

	vs := NewTable()
	for i, v := range versions {
		vs.Int2int[int64(i+1)] = int64(v)
	}

	s.Lock()
	defer s.Unlock()
	s.GetGlobal(statsName)
	s.GetField(-1, "ReceivedVersionNegotiationPacket")
	s.Remove(-2)
	s.pushAsLuaString(local)
	s.pushAsLuaString(remote)
	s.PushTable(vs)
	err := s.Call(3, 0)
	if err != nil {
		log.Println(err)
	}

	return err
}

func (s *Stats) ReceivedRetry(local, remote *pan.UDPAddr, hdr *logging.Header) error {
	//s.Printf("ReceivedRetry: only stub implementation")
	s.Lock()
	defer s.Unlock()
	s.GetGlobal(statsName)
	s.GetField(-1, "ReceivedRetry")
	s.Remove(-2)
	s.pushAsLuaString(local)
	s.pushAsLuaString(remote)
	err := s.Call(2, 0)
	if err != nil {
		log.Println(err)
	}

	return err
}

func (s *Stats) ReceivedPacket(local, remote *pan.UDPAddr, hdr *logging.ExtendedHeader, size logging.ByteCount, frames []logging.Frame) error {
	//s.Printf("ReceivedPacket: only stub implementation")
	s.Lock()
	defer s.Unlock()
	s.GetGlobal(statsName)
	s.GetField(-1, "ReceivedPacket")
	s.Remove(-2)
	s.pushAsLuaString(local)
	s.pushAsLuaString(remote)
	s.PushInteger(int64(size))
	err := s.Call(3, 0)
	if err != nil {
		log.Println(err)
	}

	return err
}

func (s *Stats) BufferedPacket(local, remote *pan.UDPAddr, ptype logging.PacketType) error {
	//s.Printf("BufferedPacket")
	s.Lock()
	defer s.Unlock()
	s.GetGlobal(statsName)
	s.GetField(-1, "BufferedPacket")
	s.Remove(-2)
	s.pushAsLuaString(local)
	s.pushAsLuaString(remote)
	s.PushInteger(int64(ptype))
	err := s.Call(3, 0)
	if err != nil {
		log.Println(err)
	}

	return err
}

func (s *Stats) DroppedPacket(local, remote *pan.UDPAddr, ptype logging.PacketType, size logging.ByteCount, reason logging.PacketDropReason) error {
	//s.Printf("DroppedPacket")
	s.Lock()
	defer s.Unlock()
	s.GetGlobal(statsName)
	s.GetField(-1, "DroppedPacket")
	s.Remove(-2)
	s.pushAsLuaString(local)
	s.pushAsLuaString(remote)
	s.PushInteger(int64(ptype))
	s.PushInteger(int64(size))
	s.PushInteger(int64(reason))
	err := s.Call(5, 0)
	if err != nil {
		log.Println(err)
	}

	return err
}

func (s *Stats) UpdatedMetrics(local, remote *pan.UDPAddr, rttStats *rpc.RTTStats, cwnd, bytesInFlight logging.ByteCount, packetsInFlight int) error {
	//s.Printf("UpdatedMetrics")

	s.Lock()
	defer s.Unlock()
	s.GetGlobal(statsName)
	s.GetField(-1, "UpdatedMetrics")
	s.Remove(-2)
	s.pushAsLuaString(local)
	s.pushAsLuaString(remote)
	s.PushTable(new_lua_rtt_stats(rttStats))
	s.PushInteger(int64(cwnd))
	s.PushInteger(int64(bytesInFlight))
	s.PushInteger(int64(packetsInFlight))
	err := s.Call(6, 0)
	if err != nil {
		log.Println(err)
	}

	return err
}

func (s *Stats) AcknowledgedPacket(local, remote *pan.UDPAddr, level logging.EncryptionLevel, num logging.PacketNumber) error {
	//s.Printf("AcknowledgedPacket")
	s.Lock()
	defer s.Unlock()
	s.GetGlobal(statsName)
	s.GetField(-1, "AcknowledgedPacket")
	s.Remove(-2)
	s.pushAsLuaString(local)
	s.pushAsLuaString(remote)
	s.PushInteger(int64(level))
	s.PushInteger(int64(num))
	err := s.Call(4, 0)
	if err != nil {
		log.Println(err)
	}

	return err
}

func (s *Stats) LostPacket(local, remote *pan.UDPAddr, level logging.EncryptionLevel, num logging.PacketNumber, reason logging.PacketLossReason) error {
	//s.Printf("LostPacket")
	s.Lock()
	defer s.Unlock()
	s.GetGlobal(statsName)
	s.GetField(-1, "LostPacket")
	s.Remove(-2)
	s.pushAsLuaString(local)
	s.pushAsLuaString(remote)
	s.PushInteger(int64(level))
	s.PushInteger(int64(num))
	s.PushInteger(int64(reason))
	err := s.Call(5, 0)
	if err != nil {
		log.Println(err)
	}
	return err
}

func (s *Stats) UpdatedCongestionState(local, remote *pan.UDPAddr, state logging.CongestionState) error {
	//s.Printf("UpdatedCongestionState")
	s.Lock()
	defer s.Unlock()
	s.GetGlobal(statsName)
	s.GetField(-1, "UpdatedCongestionState")
	s.Remove(-2)
	s.pushAsLuaString(local)
	s.pushAsLuaString(remote)
	s.PushInteger(int64(state))
	err := s.Call(3, 0)
	if err != nil {
		log.Println(err)
	}
	return err
}

func (s *Stats) UpdatedPTOCount(local, remote *pan.UDPAddr, value uint32) error {
	//s.Printf("UpdatedPTOCount")
	s.Lock()
	defer s.Unlock()
	s.GetGlobal(statsName)
	s.GetField(-1, "UpdatedPTOCount")
	s.Remove(-2)
	s.pushAsLuaString(local)
	s.pushAsLuaString(remote)
	s.PushInteger(int64(value))
	err := s.Call(3, 0)
	if err != nil {
		log.Println(err)
	}
	return err
}

func (s *Stats) UpdatedKeyFromTLS(local, remote *pan.UDPAddr, level logging.EncryptionLevel, p logging.Perspective) error {
	//s.Printf("UpdatedKeyFromTLS")
	s.Lock()
	defer s.Unlock()
	s.GetGlobal(statsName)
	s.GetField(-1, "UpdatedKeyFromTLS")
	s.Remove(-2)
	s.pushAsLuaString(local)
	s.pushAsLuaString(remote)
	s.PushInteger(int64(level))
	s.PushInteger(int64(p))
	err := s.Call(4, 0)
	if err != nil {
		log.Println(err)
	}
	return err
}

func (s *Stats) UpdatedKey(local, remote *pan.UDPAddr, generation logging.KeyPhase, rmte bool) error {
	//s.Printf("UpdatedKey")
	s.Lock()
	defer s.Unlock()
	s.GetGlobal(statsName)
	s.GetField(-1, "UpdatedKey")
	s.Remove(-2)
	s.pushAsLuaString(local)
	s.pushAsLuaString(remote)
	s.PushInteger(int64(generation))
	s.PushBoolean(rmte)
	err := s.Call(4, 0)
	if err != nil {
		log.Println(err)
	}
	return err
}

func (s *Stats) DroppedEncryptionLevel(local, remote *pan.UDPAddr, level logging.EncryptionLevel) error {
	//s.Printf("DroppedEncryptionLevel")
	s.Lock()
	defer s.Unlock()
	s.GetGlobal(statsName)
	s.GetField(-1, "DroppedEncryptionLevel")
	s.Remove(-2)
	s.pushAsLuaString(local)
	s.pushAsLuaString(remote)
	s.PushInteger(int64(level))
	err := s.Call(3, 0)
	if err != nil {
		log.Println(err)
	}
	return err
}

func (s *Stats) DroppedKey(local, remote *pan.UDPAddr, generation logging.KeyPhase) error {
	//s.Printf("DroppedKey")
	s.Lock()
	defer s.Unlock()
	s.GetGlobal(statsName)
	s.GetField(-1, "DroppedKey")
	s.Remove(-2)
	s.pushAsLuaString(local)
	s.pushAsLuaString(remote)
	s.PushInteger(int64(generation))
	err := s.Call(3, 0)
	if err != nil {
		log.Println(err)
	}
	return err
}

func (s *Stats) SetLossTimer(local, remote *pan.UDPAddr, ttype logging.TimerType, level logging.EncryptionLevel, t time.Time) error {
	//s.Printf("SetLossTimer")
	s.Lock()
	defer s.Unlock()
	s.GetGlobal(statsName)
	s.GetField(-1, "SetLossTimer")
	s.Remove(-2)
	s.pushAsLuaString(local)
	s.pushAsLuaString(remote)
	s.PushInteger(int64(ttype))
	s.PushInteger(int64(level))
	s.pushAsLuaString(t)
	err := s.Call(5, 0)
	if err != nil {
		log.Println(err)
	}
	return err
}

func (s *Stats) LossTimerExpired(local, remote *pan.UDPAddr, ttype logging.TimerType, level logging.EncryptionLevel) error {
	//s.Printf("LossTimerExpired")
	s.Lock()
	defer s.Unlock()
	s.GetGlobal(statsName)
	s.GetField(-1, "LossTimerExpired")
	s.Remove(-2)
	s.pushAsLuaString(local)
	s.pushAsLuaString(remote)
	s.PushInteger(int64(ttype))
	s.PushInteger(int64(level))
	err := s.Call(4, 0)
	if err != nil {
		log.Println(err)
	}
	return err
}

func (s *Stats) LossTimerCanceled(local, remote *pan.UDPAddr) error {
	//s.Printf("LossTimerCanceled")
	s.Lock()
	defer s.Unlock()
	s.GetGlobal(statsName)
	s.GetField(-1, "LossTimerCanceled")
	s.Remove(-2)
	s.pushAsLuaString(local)
	s.pushAsLuaString(remote)
	err := s.Call(2, 0)
	if err != nil {
		log.Println(err)
	}
	return err
}

func (s *Stats) Close(local, remote *pan.UDPAddr) error {
	//s.Printf("Close")
	s.Lock()
	defer s.Unlock()
	s.GetGlobal(statsName)
	s.GetField(-1, "Close")
	s.Remove(-2)
	s.pushAsLuaString(local)
	s.pushAsLuaString(remote)
	err := s.Call(2, 0)
	if err != nil {
		log.Println(err)
	}
	return err
}

func (s *Stats) Debug(local, remote *pan.UDPAddr, name, msg string) error {
	//s.Printf("Debug")
	s.Lock()
	defer s.Unlock()
	s.GetGlobal(statsName)
	s.GetField(-1, "Debug")
	s.Remove(-2)
	s.pushAsLuaString(local)
	s.pushAsLuaString(remote)
	s.PushString(name)
	s.PushString(msg)
	err := s.Call(4, 0)
	if err != nil {
		log.Println(err)
	}
	return err
}
