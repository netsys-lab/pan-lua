print("HELLO FROM PROFILES SCRIPT")

-- global variable to track "time"
tick = 0

-- global table to keep function call statistics
calls = {
   cur = {},
   old = {},
}

-- shutdown
shutdown = 0


-- map remote address to path fingerprints
raddr2fps = {}

-- map path fingerprint to path
fp2path = {}

-- map path fingerprint to tick ("time") when path was last used
fp2last = {}

-- map path fingerprint to observed RTT
fp2rtt = {}

-- map path fingerprint to observed bandwidth
fp2bw = {}

-- map path fingerprint to used ticks
fp2ticks = {}

-- map path fingerprint to path ID (number)
fp2id = {}

-- map path fingerprint to number of local addresses
fp2laddrs = {}

-- map local address to preferences table
laddr2prefs = {}

-- map local address to path fingerprint
laddr2fp = {}

-- map local address to path switch microsecond
laddr2switchtime = {}

-- map local address to transferred bytes on the current path
laddr2bytes_on_path = {}

-- gets called every second or so
function panapi.Periodic(seconds)
   tick = tick + 1
   local fingerprints = {}
   for fp, _ in pairs(fp2path) do
      table.insert(fingerprints, fp)
   end
   table.sort(fingerprints)
   panapi.Log(string.rep("=", 10), "Tick", string.rep("=", 10))
   for i, fp in ipairs(fingerprints) do
      panapi.Log("Path", "[", fp, "]:", string.format("%.4f", (fp2bw[fp] or 0) / (fp2laddrs[fp] or 1)), string.rep("#", fp2laddrs[fp]))
   end

   if shutdown ~= 0 then
      panapi.Log("shutting down in", shutdown - tick, "seconds")
      if tick > shutdown then
         panapi.Log("shutting down to inspect correct path expiry behavior")
         panapi.Log(tprint(fp2bw))
         os.exit(0)
      end
   end
end


-- pick path with oldest tick (i.e., longest unused path)
function nextOldestPath(raddr)
   local fingerprints = {}
   for _, fp in ipairs(raddr2fps[raddr]) do
      table.insert(fingerprints, fp)
   end
   table.sort(
      fingerprints,
      function(fp_a, fp_b)
         return (fp2last[fp_a] or 0) < (fp2last[fp_b] or 0)
      end
   )
--   panapi.Log("Chosen tick:", fp2last[fingerprints[1]])
   return fp2path[fingerprints[1]]
end


function nextScavengePath(raddr)
   local fingerprints = {}
   for _, fp in ipairs(raddr2fps[raddr]) do
      table.insert(fingerprints, fp)
   end
   table.sort(
      fingerprints,
      function(fp_a, fp_b)
         return (fp2laddrs[fp_a] or 0) < (fp2laddrs[fp_b] or 0)
      end
   )
--   panapi.Log("Chosen tick:", fp2last[fingerprints[1]])
   return fp2path[fingerprints[1]]
end

-- pick path with worst RTT
function nextWorstRTTPath(raddr)
   local fingerprints = {}
   for _, fp in ipairs(raddr2fps[raddr]) do
      table.insert(fingerprints, fp)
   end
    table.sort(
      fingerprints,
      function(fp_a, fp_b)
         return (fp2rtt[fp_a] or 1000) > (fp2rtt[fp_b] or 1000)
      end
    )
    return fp2path[fingerprints[1]]
end

function nextBestRTTPath(raddr)
   local fingerprints = {}
   for _, fp in ipairs(raddr2fps[raddr]) do
      table.insert(fingerprints, fp)
   end
    table.sort(
      fingerprints,
      function(fp_a, fp_b)
         return (fp2rtt[fp_a] or 1000) < (fp2rtt[fp_b] or 0)
      end
    )
   return fp2path[fingerprints[1]]
end

function nextBestBWPath(laddr, raddr)
   local fingerprints = {}
   for _, fp in ipairs(raddr2fps[raddr]) do
      table.insert(fingerprints, fp)
   end
   table.sort(
      fingerprints,
      function(fp_a, fp_b)
         return (fp2bw[fp_a] or 1000) / (fp2laddrs[fp_a] or 1) > (fp2bw[fp_b] or 0) / (fp2laddrs[fp_b] or 1)
      end
    )
   return fp2path[fingerprints[1]]
end

function nextRandomPath(raddr)
   local fingerprints = {}
   for _, fp in ipairs(raddr2fps[raddr]) do
      table.insert(fingerprints, fp)
   end
   return fp2path[fingerprints[math.random(#fingerprints)]]
end


-- gets called when a set of paths to addr is known
function panapi.Initialize(prefs, laddr, raddr, ps)
   panapi.Log("New connection [" .. laddr, "|", raddr .. "]")
   raddr2fps[raddr] = raddr2fps[raddr] or {}
   laddr2switchtime[laddr] = panapi.Now()
   table.sort(ps, function(a, b) return a.Fingerprint < b.Fingerprint end)
   for i, path in ipairs(ps) do
      local fp = path.Fingerprint
      fp2path[fp] = path
--      fp2last[fp] = tick
      fp2id[fp] = i
      fp2laddrs[fp] = fp2laddrs[fp] or 0
      panapi.Log("Path", i, fp)
      local exists = false
      for i, fp2 in ipairs(raddr2fps[raddr]) do
         if fp == fp2 then
            exists = true
            break
         end
      end
      if exists == false then
         table.insert(raddr2fps[raddr], fp)
      end
   end
   panapi.SetPreferences(prefs, laddr, raddr)
end

function panapi.SetPreferences(prefs, laddr, raddr)
   panapi.Log("Update Preferences [" .. laddr, "|", raddr .. "] Profile:", tprint(prefs))
   if prefs.ConnCapacityProfile == "CapacitySeeking" then

   end
   if prefs ~= nil then
      laddr2prefs[laddr] = laddr2prefs[laddr] or {}
      laddr2prefs[laddr] = prefs
   end
end

-- gets called for every packet
-- implementation needs to be efficient
function panapi.Path(laddr, raddr)
   if #raddr2fps[raddr] > 0 then
      local oldfp = laddr2fp[laddr]
      local p = fp2path[oldfp]
      local profile = laddr2prefs[laddr]["ConnCapacityProfile"]
      local now = panapi.Now()
      if p == nil then
         p = nextScavengePath(raddr)
      elseif now - laddr2switchtime[laddr] > 1000000 then
         if math.random(100) == 1 then
            profile = "Exploration"
            p = nextOldestPath(raddr)
         elseif profile == "LowLatencyInteractive" or profile == "LowLatencyNonInteractive" then
            p = nextBestRTTPath(raddr)
         elseif profile == "CapacitySeeking" then
            p = nextBestBWPath(laddr, raddr)
         elseif profile == "Scavenger" then
            -- set to scavenger by default
            profile = "Scavenger"
            p = nextScavengePath(raddr)
         else
            profile = "Random"
            p = nextRandomPath(raddr)
         end
      end
      if p and p.Fingerprint ~= oldfp then
	 fp2last[p.Fingerprint] = tick
         --panapi.Log("Changed path [" .. laddr, "|", raddr .. "]:", profile, "from Path", fp2id[oldfp], "to Path", fp2id[p.Fingerprint])
      -- keep track of chosen path via local address
         laddr2fp[laddr] = p.Fingerprint
         if oldfp then
            fp2laddrs[oldfp] = fp2laddrs[oldfp] - 1
            fp2bw[oldfp] = ((fp2bw[oldfp] or 0 ) * 9 + (laddr2bytes_on_path[laddr] or 0) / (now - laddr2switchtime[laddr])) / 10
         end
         laddr2bytes_on_path[laddr] = 0
         laddr2switchtime[laddr] = now
	 fp2laddrs[p.Fingerprint] = (fp2laddrs[p.Fingerprint] or 0) + 1
      end
      if p == nil then
         panapi.Log("path gone, trying to recover")
         return panapi.Path(laddr, raddr)
      end
      return p
   end
end

-- gets called whenever a path disappears(?)
function panapi.PathDown(laddr, raddr, fp, pi)
   panapi.Log("PathDown called with", laddr, raddr, fp, pi)
   fp2path[fp] = nil
   fp2rtt[fp] = nil
   fp2last[fp] = nil
   fp2bw[fp] = nil
   fp2laddrs[fp] = 0

   for i,fp2 in ipairs(raddr2fps[raddr]) do
      if fp == fp2 then
         table.remove(raddr2fps[raddr], i)
      end
   end
   for laddr,fp2 in pairs(laddr2fp) do
      if fp == fp2 then
         laddr2fp[laddr] = nil
      end
   end
       
--   shutdown = tick + 10
end

function panapi.Refresh(laddr, raddr, ps)
   panapi.Log("Refresh", raddr, ps)
   panapi.Initialize(nil, laddr, raddr, ps)
--   shutdown = tick + 10
end


function panapi.Close(laddr, raddr)
   panapi.Log("Close", laddr, raddr)
   --panapi.Log(tprint(fp2bw))
   --panapi.Log(tprint(fp2laddrs))
   local fp = laddr2fp[laddr]
   laddr2fp[laddr] = nil
   fp2laddrs[fp] = (fp2laddrs[fp] or 1) - 1
   laddr2prefs[laddr] = nil
   --shutdown = tick + 10
end


function stats.UpdatedMetrics(laddr, raddr, rttStats, cwnd, bytesInFlight, packetsInFlight)
   calls.cur.UpdatedMetrics = (calls.cur.UpdatedMetrics or 0) + 1
   local fp = laddr2fp[laddr]
   if fp == nil then
       return 
   end
--   fp2last[fp] = tick
   fp2rtt[fp] = rttStats.LatestRTT
--   panapi.Log("\n", tprint(rttStats, 1))
end

function stats.SentPacket(laddr, raddr, size)
   calls.cur.SentPacket = (calls.cur.SentPacket or 0) + 1
   laddr2bytes_on_path[laddr] = (laddr2bytes_on_path[laddr] or 0) + size
end


function diffcallstats()
   local d = {}
   for k,v in pairs(calls.cur) do
      local old = calls.old[k]
      if v ~= old then
         d[k] = v - (old or 0)
      end
   end
   calls.old = copy(calls.cur)
   panapi.Log("function calls since last tick:\n" .. tprint(d, 2))
end

function abscallstats()
   local d = {}
   for k,v in pairs(calls.cur) do
      if v ~= 0 then
         d[k] = v
      end
   end
   panapi.Log("function calls:\n" .. tprint(d, 2))
end

function stats.TracerForConnection(id, p, odcid)
   calls.cur.TracerForConnection = (calls.cur.TracerForConnection or 0) + 1
   --panapi.Log("id:", id, "perspective", p, "odcid", odcid)
end
function stats.StartedConnection(laddr, raddr, srcid, dstid)
   calls.cur.StartedConnection = (calls.cur.StartedConnection or 0) + 1
   
end
function stats.NegotiatedVersion(laddr, raddr)
   calls.cur.NegotiatedVersion = (calls.cur.NegotiatedVersion or 0) + 1

end
function stats.ClosedConnection(laddr, raddr)
   calls.cur.ClosedConnection = (calls.cur.ClosedConnection or 0) + 1

end
function stats.Close(laddr, raddr)
   calls.cur.Close = (calls.cur.Close or 0) + 1

end
function stats.SentTransportParameters(laddr, raddr)
   calls.cur.SentTransportParameters = (calls.cur.SentTransportParameters or 0) + 1

end
function stats.ReceivedTransportParameters(laddr, raddr)
   calls.cur.ReceivedTransportParameters = (calls.cur.ReceivedTransportParameters or 0) + 1

end
function stats.RestoredTransportParameters(laddr, raddr)
   calls.cur.RestoredTransportParameters = (calls.cur.RestoredTransportParameters or 0) + 1

end
function stats.ReceivedVersionNegotiationPacket(laddr, raddr)
   calls.cur.ReceivedVersionNegotiationPacket = (calls.cur.ReceivedVersionNegotiationPacket or 0) + 1

end
function stats.ReceivedRetry(laddr, raddr)
   calls.cur.ReceivedRetry = (calls.cur.ReceivedRetry or 0) + 1

end
function stats.ReceivedPacket(laddr, raddr)
   calls.cur.ReceivedPacket = (calls.cur.ReceivedPacket or 0) + 1

end
function stats.BufferedPacket(laddr, raddr)
   calls.cur.BufferedPacket = (calls.cur.BufferedPacket or 0) + 1

end
function stats.DroppedPacket(laddr, raddr)
   calls.cur.DroppedPacket = (calls.cur.DroppedPacket or 0) + 1

end
function stats.AcknowledgedPacket(laddr, raddr)
   calls.cur.AcknowledgedPacket = (calls.cur.AcknowledgedPacket or 0) + 1

end
function stats.LostPacket(laddr, raddr)
   calls.cur.LostPacket = (calls.cur.LostPacket or 0) + 1

end
function stats.UpdatedCongestionState(laddr, raddr)
   calls.cur.UpdatedCongestionState = (calls.cur.UpdatedCongestionState or 0) + 1

end
function stats.UpdatedPTOCount(laddr, raddr)
   calls.cur.UpdatedPTOCount = (calls.cur.UpdatedPTOCount or 0) + 1

end
function stats.UpdatedKeyFromTLS(laddr, raddr)
   calls.cur.UpdatedKeyFromTLS = (calls.cur.UpdatedKeyFromTLS or 0) + 1

end
function stats.UpdatedKey(laddr, raddr)
   calls.cur.UpdatedKey = (calls.cur.UpdatedKey or 0) + 1

end
function stats.DroppedEncryptionLevel(laddr, raddr)
   calls.cur.DroppedEncryptionLevel = (calls.cur.DroppedEncryptionLevel or 0) + 1

end
function stats.DroppedKey(laddr, raddr)
   calls.cur.DroppedKey = (calls.cur.DroppedKey or 0) + 1

end
function stats.SetLossTimer(laddr, raddr)
   calls.cur.SetLossTimer = (calls.cur.SetLossTimer or 0) + 1

end
function stats.LossTimerExpired(laddr, raddr)
   calls.cur.LossTimerExpired = (calls.cur.LossTimerExpired or 0) + 1

end
function stats.LossTimerCanceled(laddr, raddr)
   calls.cur.LossTimerCanceled = (calls.cur.LossTimerCanceled or 0) + 1

end
function stats.Debug(laddr, raddr)
   calls.cur.Debug = (calls.cur.Debug or 0) + 1
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

