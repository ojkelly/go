package fsm

import "fmt"

type Event int

// EventNames are optional, but useful for debugging and will print out in
// error messages
type EventNames map[Event]string

// AddEventNames will add Event names for debugging.
// This can only be called once, and does not affect the functioning of the
// machine.
func (m *Machine) AddEventNames(e EventNames) {
	if m.hasSetEventNames {
		return
	}

	m.hasSetEventNames = true
	m.eventNames = e
}

// GetNameForEvent will return the name set in EventNames for the given
// Event, or it will return the Event int as a string
func (m *Machine) GetNameForEvent(s Event) string {
	if sn, ok := m.eventNames[s]; ok {
		return sn
	}
	return fmt.Sprintf("%d", s)
}

type eventMap map[Event]bool
type EventToTransition map[Event]Transition

// SendEvent to the fsm.Machine to change States
func (m *Machine) SendEvent(e Event) {
	m.checkIfCreatedCorrectly()

	// validate event
	found := m.events[e]
	if !found {
		panic(
			fmt.Sprintf("[%s] fsm.Machine.Event() called with unregistered Event. All events must be registered in fsm.New()", m.id))
	}

	// get current state node
	currentState := m.state

	node := m.states[currentState]

	var transition Transition
	var foundEvent bool

	for ev, t := range node.Events {
		if e == ev {
			transition = t
			foundEvent = true
		}
	}

	if !foundEvent {
		return
	}

	if transition.Guard != nil {
		guardPass := transition.Guard(m, currentState, transition.State)

		if !guardPass {
			return
		}
	}

	// Transition hooks only run if the state changes to a different one
	if currentState != transition.State {
		if ev := node.Events[e]; ev.Exit != nil {
			ev.Exit(m, currentState, transition.State, TransitionEventEntry)
		}

		if transition.Entry != nil {
			transition.Entry(m, currentState, transition.State, TransitionEventEntry)
		}
	}

	if transition.UpdateContext != nil {
		m.handleContextUpdate(transition, currentState)
	}

	m.state = transition.State

	if m.stateChangeChannel != nil {
		m.stateChangeChannel <- StateChange{
			From:  currentState,
			To:    transition.State,
			Cause: e,
		}
	}
}

// Success is called by you when a state has completed successfully, and
// you want the FSM to transition automatically
func (m *Machine) Success() {
	m.checkIfCreatedCorrectly()

	currentState := m.state
	node := m.states[currentState]

	handler := node.Success
	if handler != nil {
		handler(m, currentState, currentState, TransitionEventSuccess)
		return
	}
}

// func (m *Machine) StateChangeChannel() *
