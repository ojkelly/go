# Control

## fsm - WIP

A Finite State Machine for go.

```golang
package main

import (
	"github.com/ojkelly/control/fsm"
)
```

## How it works

- `Events` cause a `Transition` from one `State` to another `State`
- `Guard` functions can prevent `Transitions`
- `fsm.Context` is available for data needed to inform `Guards`
- All `States`, `Transistions`, `Events`, and handlers are known
