package fsm

import (
	"fmt"
	"sync"
)

// Machine
type Machine struct {
	// internals
	initWithNew        bool
	context            internalContext
	events             eventMap
	state              State
	states             States
	id                 string
	errorHandler       MachineErrorHandler
	stateChangeChannel chan StateChange

	stateChangeMtx   sync.Mutex
	contextChangeMtx sync.Mutex

	lockPublicSet sync.Mutex

	// Debug / Optional
	hasSetStateNames      bool
	stateNames            StateNames
	hasSetEventNames      bool
	eventNames            EventNames
	hasSetContextKeyNames bool
	contextKeyNames       ContextKeyNames
}

// New Machine
//
// stateChangeChannelSize should be set to the number of channels you will be
// reacting to events from. Most of the time this is safe to set to 1.
func New(
	id string,
	stateChangeChannelSize int,
	initialState State,
	context Context,
	events []Event,
	states States,
	errorHandler MachineErrorHandler,
) *Machine {
	// convert Events to a map for fast lookup
	eMap := eventMap{}
	for _, e := range events {
		eMap[e] = true
	}

	cMap := internalContext{}
	for c, meta := range context {
		cMap[c] = &contextMeta{
			key:       c,
			protected: meta.Protected,
			value:     meta.Inital,
		}
	}

	return &Machine{
		initWithNew:        true,
		events:             eMap,
		state:              initialState,
		states:             states,
		context:            cMap,
		id:                 id,
		stateChangeChannel: make(chan StateChange, stateChangeChannelSize),
		lockPublicSet:      sync.Mutex{},
		// Debug
		hasSetStateNames:      false,
		stateNames:            StateNames{},
		hasSetEventNames:      false,
		eventNames:            EventNames{},
		hasSetContextKeyNames: false,
		contextKeyNames:       ContextKeyNames{},
	}
}

// Id returns the id string of this machine
func (m *Machine) Id() string {
	m.checkIfCreatedCorrectly()
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

// Stop the machine, and signalt to all watching for changes, that it's done
func (m *Machine) Stop() {
	m.stateChangeChannel <- StateChange{
		From:   m.state,
		To:     m.state,
		IsLast: true,
	}
}
