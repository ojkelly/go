# Go libraries

This repository is a collection of go libraries I've built and maintain.

## fsm - a WIP Finite State machine

A Finite State Machine for go.

```golang
package main

import (
	"ojkelly.dev/fsm"
)
```

### How it works

- `Events` cause a `Transition` from one `State` to another `State`
- `Guard` functions can prevent `Transitions`
- `fsm.Context` is available for data needed to inform `Guards`
- All `States`, `Transistions`, `Events`, and handlers are known
