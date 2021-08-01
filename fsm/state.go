package fsm

import "fmt"

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
type State int

// StateNames are optional, but useful for debugging and will print out in
// error messages
type StateNames map[State]string

// AddStateNames will add state names for debugging.
// This can only be called once, and does not affect the functioning of the
// machine.
func (m *Machine) AddStateNames(sn StateNames) {
	m.checkIfCreatedCorrectly()

	if m.hasSetStateNames {
		return
	}

	m.hasSetStateNames = true
	m.stateNames = sn
}

// GetNameForState will return the name set in StateNames for the given
// State, or it will return the State int as a string
func (m *Machine) GetNameForState(s State) string {
	m.checkIfCreatedCorrectly()

	if sn, ok := m.stateNames[s]; ok {
		return sn
	}
	return fmt.Sprintf("%d", s)
}

// States
type States map[State]StateNode

type StateNode struct {
	// TODO: nested states
	// States *States
	// Error is a special predefined event
	Error MachineErrorHandler

	// Success is a special predefined event
	Success TransitionEventHandler

	// Events this State can Transition to
	Events EventToTransition
}

// State returns the current state the Machine is in
func (m *Machine) State() State {
	m.checkIfCreatedCorrectly()
	m.stateChangeMtx.Lock()
	defer m.stateChangeMtx.Unlock()
	return m.state
}

// GetNextStates returns the States if any that can be transistioned to
func (m *Machine) GetNextStates() []State {
	m.checkIfCreatedCorrectly()

	return []State{}
}

// StateChange event sent via m.StateChangeChannel() after a State transition
// has completed.
// If IsLast is true, it means m.Stop() has been called, and your watcher
// can we stopped.
type StateChange struct {
	From   State
	To     State
	Cause  Event
	IsLast bool
}

// StateChangeChannel receives an StatesChange after the transition from one
// State to another has completed.
func (m *Machine) StateChangeChannel() <-chan StateChange {
	return m.stateChangeChannel
}
