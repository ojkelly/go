package fsm

import (
	"fmt"
	"sync"
)

// Context is used to store extra state that is considered when transitioning
// between states

// State
//
// Create an enum of State:
//
//  	const (
//  		StepZero fsm.State = iota
//  		StepOne
//  		StepTwo
//  		SetThree
//  	)
//
type State string

// States
type States map[State]StateNode

type StateNode struct {
	// TODO: nested states
	// States *States
	// Error is a special predefined event
	Error TransitionEventHandler
	// Success is a special predefined event
	Success TransitionEventHandler

	Events EventToTransition
}

// Machine
type Machine struct {
	// internals
	initWithNew  bool
	mtx          sync.RWMutex
	context      internalContext
	state        State
	events       eventMap
	states       States
	id           string
	errorHandler TransitionEventHandler
}

// New Machine
func New(
	id string,
	initialState State,
	context Context,
	events []Event,
	states States,
	errorHandler TransitionEventHandler,
) *Machine {
	// convert Events to a map for fast lookup
	eMap := eventMap{}
	for _, e := range events {
		eMap[e] = true
	}

	cMap := internalContext{}
	for c, meta := range context {
		cMap[c] = &contextMeta{
			key:   c,
			write: meta.Write,
			value: meta.Inital,
		}
	}

	return &Machine{
		initWithNew: true,
		mtx:         sync.RWMutex{},
		state:       initialState,
		events:      eMap,
		states:      states,
		context:     cMap,
		id:          id,
	}
}

func (m *Machine) Id() string {
	return m.id
}

// If the fsm.Machine wasn't made with New then we panic
// there's no way we can guarentee it will work
//
// This is a developer error.
func (m *Machine) checkIfCreatedCorrectly() {
	if !m.initWithNew {
		panic(fmt.Sprintf("[%s] fsm.Machine was not created with fsm.New()", m.id))
	}
}

// State returns the current state the Machine is in
func (m *Machine) State() State {
	// m.mtx.RLock()
	// defer m.mtx.RUnlock()

	return m.state
}

// GetNextStates returns the States if any that can be transistioned to
func (m *Machine) GetNextStates() []State {
	// m.mtx.RLock()
	// defer m.mtx.RUnlock()
	return []State{}
}
