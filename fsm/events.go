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
// blocks until the state transition has completed or failed
func (m *Machine) SendEvent(e Event) bool {
	m.checkIfCreatedCorrectly()

	m.stateChangeMtx.Lock()
	defer m.stateChangeMtx.Unlock()
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
		if m.errorHandler != nil {
			m.errorHandler(m, currentState, currentState, MachineErrorEventNotFoundForState)
		}
		return false
	}

	if transition.Guard != nil {
		guardPass := transition.Guard(m, currentState, transition.State)

		if !guardPass {
			if m.errorHandler != nil {
				m.errorHandler(m, currentState, currentState, MachineErrorGuardFail)
			}
			return false
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
		m.handleUpdateContext(transition, currentState)
	}

	m.state = transition.State

	m.stateChangeChannel <- StateChange{
		From:  currentState,
		To:    transition.State,
		Cause: e,
	}
	return true
}
