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

