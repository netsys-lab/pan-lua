# Lua Selector

* Follows pan.Selector interface

## Lua API

Lua scripts needs to implement the following functions:

```
-- gets called when a set of paths to addr is known
function panapi.Initialize(prefs, laddr, raddr, ps)

function panapi.SetPreferences(prefs, laddr, raddr)

-- gets called for every packet
-- implementation needs to be efficient
function panapi.Path(laddr, raddr)

-- gets called whenever a path disappears
function panapi.PathDown(laddr, raddr, fp, pi)

function panapi.Refresh(laddr, raddr, ps)

function panapi.Close(laddr, raddr)

function panapi.Periodic(seconds)
```

Lua scripts can call the following functions from the panapi module:
```
panapi.Log(...)
```

# Quic Tracer

QUIC connection properties are available in the following functions:

```Lua
function stats.UpdatedMetrics(laddr, raddr, rttStats, cwnd, bytesInFlight, packetsInFlight)

function stats.SentPacket(laddr, raddr, size)

function stats.TracerForConnection(id, p, odcid)

function stats.StartedConnection(laddr, raddr, srcid, dstid)

function stats.NegotiatedVersion(laddr, raddr)

function stats.ClosedConnection(laddr, raddr)

function stats.Close(laddr, raddr)

function stats.SentTransportParameters(laddr, raddr)

function stats.ReceivedTransportParameters(laddr, raddr)

function stats.RestoredTransportParameters(laddr, raddr)

function stats.ReceivedVersionNegotiationPacket(laddr, raddr)

function stats.ReceivedRetry(laddr, raddr)

function stats.ReceivedPacket(laddr, raddr)

function stats.BufferedPacket(laddr, raddr)

function stats.DroppedPacket(laddr, raddr)

function stats.AcknowledgedPacket(laddr, raddr)

function stats.LostPacket(laddr, raddr)

function stats.UpdatedCongestionState(laddr, raddr)

function stats.UpdatedPTOCount(laddr, raddr)

function stats.UpdatedKeyFromTLS(laddr, raddr)

function stats.UpdatedKey(laddr, raddr)

function stats.DroppedEncryptionLevel(laddr, raddr)

function stats.DroppedKey(laddr, raddr)

function stats.SetLossTimer(laddr, raddr)

function stats.LossTimerExpired(laddr, raddr)

function stats.LossTimerCanceled(laddr, raddr)

function stats.Debug(laddr, raddr)
```
