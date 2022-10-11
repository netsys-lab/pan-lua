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
	"context"
	"errors"
	"log"

	"github.com/lucas-clemente/quic-go"
	"github.com/lucas-clemente/quic-go/logging"
)

type DebugConnectionTracerServer struct {
	l        *log.Logger
	tracer   logging.Tracer
	ctracers map[uint64]logging.ConnectionTracer
}

func NewDebugConnectionTracerServer(tracer logging.Tracer, l *log.Logger) *DebugConnectionTracerServer {
	return &DebugConnectionTracerServer{l, tracer, map[uint64]logging.ConnectionTracer{}}
}

func (c *DebugConnectionTracerServer) NewTracerForConnection(args *ConnectionTracerMsg, resp *NilMsg) error {
	if args.OdcID == nil {
		return ErrDeref
	}
	tracing_id := args.TracingID
	c.ctracers[tracing_id] = c.tracer.TracerForConnection(context.WithValue(context.Background(), quic.SessionTracingKey, tracing_id), args.Perspective, *args.OdcID)

	return nil
}

func (c *DebugConnectionTracerServer) StartedConnection(args *ConnectionTracerMsg, resp *NilMsg) error {
	c.l.Println("StartedConnection called")
	if args.Local == nil || args.Remote == nil || args.SrcConnID == nil || args.DestConnID == nil {
		return ErrDeref
	}
	tracing_id := args.TracingID
	c.ctracers[tracing_id].StartedConnection(args.Local, args.Remote, *args.SrcConnID, *args.DestConnID)
	return nil
}
func (c *DebugConnectionTracerServer) NegotiatedVersion(args *ConnectionTracerMsg, resp *NilMsg) error {
	c.l.Println("NegotiatedVersion called")
	tracing_id := args.TracingID
	c.ctracers[tracing_id].NegotiatedVersion(args.Chosen, args.ClientVersions, args.ServerVersions)
	return nil
}
func (c *DebugConnectionTracerServer) ClosedConnection(args *ConnectionTracerMsg, resp *NilMsg) error {
	c.l.Println("ClosedConnection called")
	tracing_id := args.TracingID

	if args.ErrorMsg == nil {
		c.ctracers[tracing_id].ClosedConnection(nil)
	} else {
		c.ctracers[tracing_id].ClosedConnection(errors.New(*args.ErrorMsg))
	}
	return nil
}
func (c *DebugConnectionTracerServer) SentTransportParameters(args *ConnectionTracerMsg, resp *NilMsg) error {
	c.l.Println("SentTransportParameters called")
	tracing_id := args.TracingID
	c.ctracers[tracing_id].SentTransportParameters(args.Parameters)
	return nil
}
func (c *DebugConnectionTracerServer) ReceivedTransportParameters(args *ConnectionTracerMsg, resp *NilMsg) error {
	c.l.Println("ReceivedTransportParameters called")
	tracing_id := args.TracingID
	c.ctracers[tracing_id].ReceivedTransportParameters(args.Parameters)
	return nil
}
func (c *DebugConnectionTracerServer) RestoredTransportParameters(args *ConnectionTracerMsg, resp *NilMsg) error {
	c.l.Println("RestoredTransportParameters called")
	tracing_id := args.TracingID
	c.ctracers[tracing_id].RestoredTransportParameters(args.Parameters)
	return nil
}
func (c *DebugConnectionTracerServer) SentPacket(args *ConnectionTracerMsg, resp *NilMsg) error {
	c.l.Println("SentPacket called")
	tracing_id := args.TracingID
	c.ctracers[tracing_id].SentPacket(args.ExtendedHeader, args.ByteCount, args.AckFrame, args.Frames)
	return nil
}
func (c *DebugConnectionTracerServer) ReceivedVersionNegotiationPacket(args *ConnectionTracerMsg, resp *NilMsg) error {
	c.l.Println("ReceivedVersionNegotiationPacket called")
	tracing_id := args.TracingID
	c.ctracers[tracing_id].ReceivedVersionNegotiationPacket(args.Header, args.Versions)
	return nil
}
func (c *DebugConnectionTracerServer) ReceivedRetry(args *ConnectionTracerMsg, resp *NilMsg) error {
	c.l.Println("ReceivedRetry called")
	tracing_id := args.TracingID
	c.ctracers[tracing_id].ReceivedRetry(args.Header)
	return nil
}
func (c *DebugConnectionTracerServer) ReceivedPacket(args *ConnectionTracerMsg, resp *NilMsg) error {
	c.l.Println("ReceivedPacket called")
	tracing_id := args.TracingID
	c.ctracers[tracing_id].ReceivedPacket(args.ExtendedHeader, args.ByteCount, args.Frames)
	return nil
}
func (c *DebugConnectionTracerServer) BufferedPacket(args *ConnectionTracerMsg, resp *NilMsg) error {
	c.l.Println("BufferedPacket called")
	tracing_id := args.TracingID
	c.ctracers[tracing_id].BufferedPacket(args.PacketType)
	return nil
}
func (c *DebugConnectionTracerServer) DroppedPacket(args *ConnectionTracerMsg, resp *NilMsg) error {
	c.l.Println("DroppedPacket called")
	tracing_id := args.TracingID
	c.ctracers[tracing_id].DroppedPacket(args.PacketType, args.ByteCount, args.DropReason)
	return nil
}
func (c *DebugConnectionTracerServer) UpdatedMetrics(args *ConnectionTracerMsg, resp *NilMsg) error {
	c.l.Println("UpdatedMetrics called")
	tracing_id := args.TracingID
	c.ctracers[tracing_id].UpdatedMetrics(&logging.RTTStats{}, args.Cwnd, args.ByteCount, args.Packets)
	return nil
}
func (c *DebugConnectionTracerServer) AcknowledgedPacket(args *ConnectionTracerMsg, resp *NilMsg) error {
	c.l.Println("AcknowledgedPacket called")
	tracing_id := args.TracingID
	c.ctracers[tracing_id].AcknowledgedPacket(args.EncryptionLevel, args.PacketNumber)
	return nil
}
func (c *DebugConnectionTracerServer) LostPacket(args *ConnectionTracerMsg, resp *NilMsg) error {
	c.l.Println("LostPacket called")
	tracing_id := args.TracingID
	c.ctracers[tracing_id].LostPacket(args.EncryptionLevel, args.PacketNumber, args.LossReason)
	return nil
}
func (c *DebugConnectionTracerServer) UpdatedCongestionState(args *ConnectionTracerMsg, resp *NilMsg) error {
	c.l.Printf("UpdatedCongestionState called")
	tracing_id := args.TracingID
	c.ctracers[tracing_id].UpdatedCongestionState(args.CongestionState)
	return nil
}
func (c *DebugConnectionTracerServer) UpdatedPTOCount(args *ConnectionTracerMsg, resp *NilMsg) error {
	c.l.Println("UpdatedPTOCount called")
	tracing_id := args.TracingID
	c.ctracers[tracing_id].UpdatedPTOCount(args.PTOCount)
	return nil
}
func (c *DebugConnectionTracerServer) UpdatedKeyFromTLS(args *ConnectionTracerMsg, resp *NilMsg) error {
	c.l.Println("UpdatedKeyFromTLS called")
	tracing_id := args.TracingID
	c.ctracers[tracing_id].UpdatedKeyFromTLS(args.EncryptionLevel, args.Perspective)
	return nil
}
func (c *DebugConnectionTracerServer) UpdatedKey(args *ConnectionTracerMsg, resp *NilMsg) error {
	c.l.Println("UpdatedKey called")
	tracing_id := args.TracingID
	c.ctracers[tracing_id].UpdatedKey(args.Generation, args.Bool)
	return nil
}
func (c *DebugConnectionTracerServer) DroppedEncryptionLevel(args *ConnectionTracerMsg, resp *NilMsg) error {
	c.l.Println("DroppedEncryptionLevel called")
	tracing_id := args.TracingID
	c.ctracers[tracing_id].DroppedEncryptionLevel(args.EncryptionLevel)
	return nil
}
func (c *DebugConnectionTracerServer) DroppedKey(args *ConnectionTracerMsg, resp *NilMsg) error {
	c.l.Println("DroppedKey called")
	tracing_id := args.TracingID
	c.ctracers[tracing_id].DroppedKey(args.Generation)
	return nil
}
func (c *DebugConnectionTracerServer) SetLossTimer(args *ConnectionTracerMsg, resp *NilMsg) error {
	c.l.Println("SetLossTimer called")
	if args.Time == nil {
		return ErrDeref
	}
	tracing_id := args.TracingID
	c.ctracers[tracing_id].SetLossTimer(args.TimerType, args.EncryptionLevel, *args.Time)
	return nil
}
func (c *DebugConnectionTracerServer) LossTimerExpired(args *ConnectionTracerMsg, resp *NilMsg) error {
	c.l.Println("LossTimerExpired called")
	tracing_id := args.TracingID
	c.ctracers[tracing_id].LossTimerExpired(args.TimerType, args.EncryptionLevel)
	return nil
}
func (c *DebugConnectionTracerServer) LossTimerCanceled(args *ConnectionTracerMsg, resp *NilMsg) error {
	c.l.Println("LossTimerCanceled called")
	tracing_id := args.TracingID
	c.ctracers[tracing_id].LossTimerCanceled()
	return nil
}
func (c *DebugConnectionTracerServer) Close(args *ConnectionTracerMsg, resp *NilMsg) error {
	c.l.Println("Close called")
	tracing_id := args.TracingID
	c.ctracers[tracing_id].Close()
	return nil
}
func (c *DebugConnectionTracerServer) Debug(args *ConnectionTracerMsg, resp *NilMsg) error {
	c.l.Println("Debug called")
	if args.Key == nil || args.Value == nil {
		return ErrDeref
	}
	tracing_id := args.TracingID
	c.ctracers[tracing_id].Debug(*args.Key, *args.Value)
	return nil
}
