package fsm

import "fmt"

type Event string
type eventMap map[Event]bool
type EventToTransition map[Event]Transition

func (m *Machine) Event(e Event) {
	m.checkIfCreatedCorrectly()
	// m.mtx.RLock()
	// defer m.mtx.RUnlock()

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

	// Transition hooks
	// m.mtx.RUnlock()
	if ev := node.Events[e]; ev.Exit != nil {
		ev.Exit(m, currentState, transition.State, TransitionEventEntry)
	}

	if transition.Entry != nil {
		transition.Entry(m, currentState, transition.State, TransitionEventEntry)
	}

	if transition.Update != nil {
		key, value, err := transition.Update(m, currentState, transition.State, TransitionEventEntry)

		if err == nil && value != nil {
			if v := m.context[key]; v != nil {
				m.context[key] = &contextMeta{
					key:   key,
					write: v.write,
					value: value,
				}
			}
		}
	}
	// m.mtx.RLock()

	m.state = transition.State
}

// Error is called by you when a state encounters an error
// the FSM will check and see if there is an error handler
// otherwise it will bubble the error up to the top level
// error handler
func (m *Machine) Error() {
	currentState := m.state
	node := m.states[currentState]

	handler := node.Error
	if handler != nil {
		handler(m, currentState, currentState, TransitionEventError)
		return
	}
	m.errorHandler(m, currentState, currentState, TransitionEventError)
}

// Success is called by you when a state has completed successfully, and
// you want the FSM to transition automatically
func (m *Machine) Success() {
	currentState := m.state
	node := m.states[currentState]

	handler := node.Success
	if handler != nil {
		handler(m, currentState, currentState, TransitionEventError)
		return
	}

}
