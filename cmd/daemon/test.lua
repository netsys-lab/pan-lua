print("HELLO FROM TEST SCRIPT")

-- global variable to track "time"
tick = 0

-- global table to keep function call statistics
calls = {
   cur = {},
   old = {},
}

paths = {}

-- gets called every second or so
function panapi.Periodic(seconds)
   tick = tick + 1
   panapi.Log(string.rep("=", 10), "Tick", string.rep("=", 10))
end


-- gets called when a set of paths to addr is known
function panapi.Initialize(prefs, laddr, raddr, ps)
   panapi.Log("New connection [" .. laddr, "|", raddr .. "]")
   panapi.Log("Paths:", tprint(ps))
   paths[laddr .. "-" .. raddr] = ps
   panapi.SetPreferences(prefs, laddr, raddr)
end

function panapi.SetPreferences(prefs, laddr, raddr)
   panapi.Log("Update Preferences [" .. laddr, "|", raddr .. "] Profile:", tprint(prefs))
end

-- gets called for every packet
-- implementation needs to be efficient
function panapi.Path(laddr, raddr)
   return paths[laddr .. "-" .. raddr][1]
end

-- gets called whenever a path disappears(?)
function panapi.PathDown(laddr, raddr, fp, pi)
   panapi.Log("PathDown called with", laddr, raddr, fp, pi)
   local ps = paths[laddr .. "-" .. raddr]
   for i, p in ipairs(ps) do
      if p.Fingerprint == fp then
         panapi.Log("Found path to remove")
         table.remove(ps, i)
         break
      end
   end
end

function panapi.Refresh(laddr, raddr, ps)
   panapi.Log("Refresh", raddr, ps)
   panapi.Initialize(nil, laddr, raddr, ps)
--   shutdown = tick + 10
end


function panapi.Close(laddr, raddr)
   panapi.Log("Close", laddr, raddr)
end


function stats.UpdatedMetrics(laddr, raddr, rttStats, cwnd, bytesInFlight, packetsInFlight)
   calls.cur.UpdatedMetrics = (calls.cur.UpdatedMetrics or 0) + 1
   panapi.Log("UpdatedMetrics")
end

function stats.SentPacket(laddr, raddr, size)
   calls.cur.SentPacket = (calls.cur.SentPacket or 0) + 1
   panapi.Log("SentPacket")
end

function stats.TracerForConnection(id, p, odcid)
   calls.cur.TracerForConnection = (calls.cur.TracerForConnection or 0) + 1
   panapi.Log("TracerForConnection")
   --panapi.Log("id:", id, "perspective", p, "odcid", odcid)
end
function stats.StartedConnection(laddr, raddr, srcid, dstid)
   calls.cur.StartedConnection = (calls.cur.StartedConnection or 0) + 1
   panapi.Log("StartedConnection")
   
end
function stats.NegotiatedVersion(laddr, raddr)
   calls.cur.NegotiatedVersion = (calls.cur.NegotiatedVersion or 0) + 1
   panapi.Log("NegotiatedVersion")

end
function stats.ClosedConnection(laddr, raddr)
   calls.cur.ClosedConnection = (calls.cur.ClosedConnection or 0) + 1
   panapi.Log("ClosedConnection")

end
function stats.Close(laddr, raddr)
   calls.cur.Close = (calls.cur.Close or 0) + 1
   panapi.Log("Close")

end
function stats.SentTransportParameters(laddr, raddr)
   calls.cur.SentTransportParameters = (calls.cur.SentTransportParameters or 0) + 1
   panapi.Log("SentTransportParameters")

end
function stats.ReceivedTransportParameters(laddr, raddr)
   calls.cur.ReceivedTransportParameters = (calls.cur.ReceivedTransportParameters or 0) + 1
   panapi.Log("ReceivedTransportParameters")

end
function stats.RestoredTransportParameters(laddr, raddr)
   calls.cur.RestoredTransportParameters = (calls.cur.RestoredTransportParameters or 0) + 1
   panapi.Log("RestoredTransportParameters")

end
function stats.ReceivedVersionNegotiationPacket(laddr, raddr)
   calls.cur.ReceivedVersionNegotiationPacket = (calls.cur.ReceivedVersionNegotiationPacket or 0) + 1
   panapi.Log("ReceivedVersionNegotiationPacket")

end
function stats.ReceivedRetry(laddr, raddr)
   calls.cur.ReceivedRetry = (calls.cur.ReceivedRetry or 0) + 1
   panapi.Log("ReceivedRetry")

end
function stats.ReceivedPacket(laddr, raddr)
   calls.cur.ReceivedPacket = (calls.cur.ReceivedPacket or 0) + 1
   panapi.Log("ReceivedPacket")

end
function stats.BufferedPacket(laddr, raddr)
   calls.cur.BufferedPacket = (calls.cur.BufferedPacket or 0) + 1
   panapi.Log("BufferedPacket")

end
function stats.DroppedPacket(laddr, raddr)
   calls.cur.DroppedPacket = (calls.cur.DroppedPacket or 0) + 1
   panapi.Log("DroppedPacket")

end
function stats.AcknowledgedPacket(laddr, raddr)
   calls.cur.AcknowledgedPacket = (calls.cur.AcknowledgedPacket or 0) + 1
   panapi.Log("AcknowledgedPacket")

end
function stats.LostPacket(laddr, raddr)
   calls.cur.LostPacket = (calls.cur.LostPacket or 0) + 1
   panapi.Log("LostPacket")

end
function stats.UpdatedCongestionState(laddr, raddr)
   calls.cur.UpdatedCongestionState = (calls.cur.UpdatedCongestionState or 0) + 1
   panapi.Log("UpdatedCongestionState")

end
function stats.UpdatedPTOCount(laddr, raddr)
   calls.cur.UpdatedPTOCount = (calls.cur.UpdatedPTOCount or 0) + 1
   panapi.Log("UpdatedPTOCount")

end
function stats.UpdatedKeyFromTLS(laddr, raddr)
   calls.cur.UpdatedKeyFromTLS = (calls.cur.UpdatedKeyFromTLS or 0) + 1
   panapi.Log("UpdatedKeyFromTLS")

end
function stats.UpdatedKey(laddr, raddr)
   calls.cur.UpdatedKey = (calls.cur.UpdatedKey or 0) + 1
   panapi.Log("UpdatedKey")

end
function stats.DroppedEncryptionLevel(laddr, raddr)
   calls.cur.DroppedEncryptionLevel = (calls.cur.DroppedEncryptionLevel or 0) + 1
   panapi.Log("DroppedEncryptionLevel")

end
function stats.DroppedKey(laddr, raddr)
   calls.cur.DroppedKey = (calls.cur.DroppedKey or 0) + 1
   panapi.Log("DroppedKey")

end
function stats.SetLossTimer(laddr, raddr)
   calls.cur.SetLossTimer = (calls.cur.SetLossTimer or 0) + 1
   panapi.Log("SetLossTimer")

end
function stats.LossTimerExpired(laddr, raddr)
   calls.cur.LossTimerExpired = (calls.cur.LossTimerExpired or 0) + 1
   panapi.Log("LossTimerExpired")

end
function stats.LossTimerCanceled(laddr, raddr)
   calls.cur.LossTimerCanceled = (calls.cur.LossTimerCanceled or 0) + 1
   panapi.Log("LossTimerCanceled")

end
function stats.Debug(laddr, raddr)
   calls.cur.Debug = (calls.cur.Debug or 0) + 1
   panapi.Log("Debug")
end

-- HELPER FUNCTIONS ---
-- 
-- Print contents of `tbl`, with indentation.
-- `indent` sets the initial level of indentation.
function tprint (tbl, indent)
   if not indent then indent = 0 end
   if type(tbl) == "table" then
      local s = ""
      for k, v in pairs(tbl) do
         formatting = string.rep("  ", indent) .. tprint(k, indent) .. ": "
         if type(v) == "table" then
            --print(formatting)
            s = s ..  formatting .. "\n" .. tprint(v, indent+1)
         else
            s = s .. formatting .. tprint(v) .. "\n"
         end
      end
      return s
   else
      return tostring(tbl)
   end
end


-- recursively perform a deep copy of a table
function copy(thing)
   if type(thing) == "table" then
      local r = {}
      for k,v in pairs(thing) do
         r[k] = copy(v)
      end
      return r
   else
      return thing
   end
end

-- return a new thing containing everything about thing1 that is different from thing2
function diff(thing1, thing2)
   if type(thing1) == "table" then
      local thing = {}
      if type(thing2) == "table" then
         local thing = {}
         for k,v in pairs(thing1) do
            thing[k] = diff(thing1[k], thing2[k])
         end
         return thing
      else
         return copy(thing1)
      end
   else
      if thing1 == thing2 then
         return nil
      else
         return thing1
      end
   end
end

