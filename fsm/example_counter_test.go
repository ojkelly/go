package fsm_test

import (
	"fmt"

	"ojkelly.dev/fsm"
)

func Example_counter() {

	// States a machine can be in ------------------------------------------------
	const (
		// the first value (your zero-value) should be the default
		Inactive fsm.State = iota
		Active
	)

	stateNames := fsm.StateNames{
		Inactive: "Inactive",
		Active:   "Active",
	}

	// Events that can change state ----------------------------------------------
	const (
		Activate fsm.Event = iota
		Deactivate
		Increment
		Decrement
	)

	eventNames := fsm.EventNames{
		Activate:   "Activate",
		Deactivate: "Deactivate",
		Increment:  "Increment",
		Decrement:  "Decrement",
	}

	// ContextKeys for storing extra state ---------------------------------------
	const (
		KeyCounter fsm.ContextKey = iota
		KeyIsReady
	)

	contextKeyNames := fsm.ContextKeyNames{
		KeyCounter: "Counter",
		KeyIsReady: "IsReady",
	}

	// Event Handlers ------------------------------------------------------------
	errorHandler := func(m *fsm.Machine, current fsm.State, next fsm.State, machineError fsm.MachineError) {
		fmt.Println("Error: Left", m.GetNameForState(current), "entered", m.GetNameForState(next), machineError)
		m.SendEvent(Deactivate)
	}

	successHandler := func(m *fsm.Machine, current fsm.State, next fsm.State, event fsm.TransitionEvent) {
		fmt.Println("Success: Left", m.GetNameForState(current), "entered", m.GetNameForState(next))
	}

	logEvent := func(m *fsm.Machine, current fsm.State, next fsm.State, event fsm.TransitionEvent) {
		fmt.Println("Left", m.GetNameForState(current), "entered", m.GetNameForState(next), event)
	}

	guardActive := func(m *fsm.Machine, current fsm.State, next fsm.State) bool {
		if v := m.GetContext(KeyIsReady); v != nil {
			ready := v.(bool)
			return ready
		}
		return false
	}

	// Machine Creator -----------------------------------------------------------
	machine := fsm.New(
		// machine ID
		"counterExample",
		// initial state
		Inactive,

		// Context Keys
		fsm.Context{
			KeyIsReady: fsm.ContextMeta{
				Protected: true, // can we use .Set()
				Inital:    false,
			},
			KeyCounter: fsm.ContextMeta{
				Protected: false, // this can only be changed by events
				Inital:    0,
			},
		},

		// Possible events
		[]fsm.Event{Activate, Deactivate, Increment, Decrement},

		// State Map
		fsm.States{
			// Inactive state
			Inactive: fsm.StateNode{
				// Events that Inactive will transition on
				Events: fsm.EventToTransition{
					// On Activate event tranisition to Active
					Activate: fsm.Transition{
						State: Active,
						Guard: guardActive,
						Entry: logEvent,
						Exit: func(m *fsm.Machine, current fsm.State, next fsm.State, event fsm.TransitionEvent) {
							m.SetContext(KeyIsReady, false)
						},
					},
				},
			},

			// Active State
			Active: fsm.StateNode{
				// Set our generic error handler to print out errors
				Error: errorHandler,
				// Set a success handler to log out the success
				Success: successHandler,
				/// Register all the Events that have an affect on this State
				Events: fsm.EventToTransition{
					// for the Increment Event increase the counter, but transition
					// back to the current state
					Increment: fsm.Transition{
						State: Active, // Set the current state, because we want to stay
						// Set our ContextUpdate handler, this allows us to modify any
						// value in Context, in particular the ones marked with `write: false`
						UpdateContext: func(
							m *fsm.Machine,
							current fsm.State,
							next fsm.State,
							event fsm.TransitionEvent,
						) (
							update fsm.ContextUpdate,
							err error,
						) {
							update = fsm.ContextUpdate{}
							if v := m.GetContext(KeyCounter); v != nil {
								update[KeyCounter] = v.(int) + 1
							} else {
								err = fmt.Errorf("Unable to update KeyCounter")
							}
							return
						},
					},
					Deactivate: fsm.Transition{
						State: Inactive,
					},
				},
			},
		},
		// Machine level handlers
		errorHandler,
	)

	// This is optional, but useful if you want to enhance your logging, or
	// you have a large number of states
	machine.AddStateNames(stateNames)
	machine.AddEventNames(eventNames)
	machine.AddContextKeyNames(contextKeyNames)

	// Check the initial state of the machine
	fmt.Println("Initial state is", machine.State())

	// Again, but this time with our debug name - usefull for debugging and logs
	fmt.Println("StateName for the current state is", machine.GetNameForState(machine.State()))

	// Try to increment - nothing should happen as we're Inactive at the moment
	machine.SendEvent(Increment)

	// Let's check the counter. It shouldn't increase, as we can only do that
	// when the State is Active
	fmt.Println("Before incrementing Counter it's", machine.GetContext(KeyCounter).(int))

	// Check if the ContextKey KeyIsReady
	fmt.Println("Before setting the KeyIsReady it's", machine.GetContext(KeyIsReady).(bool))

	// Set the ContextKey KeyIsReady to true
	machine.SetContext(KeyIsReady, true)

	// Confirm that our ContextKey is set
	fmt.Println("Context KeyIsReady is", machine.GetContext(KeyIsReady).(bool))

	// Send an Activate Event, which will transition to the Active State if
	// KeyIsReady is true
	machine.SendEvent(Activate)

	// Let's see what State we're in now
	fmt.Println("After Activate Event machine State is", machine.GetNameForState(machine.State()))

	// Increment 3 times
	machine.SendEvent(Increment)
	machine.SendEvent(Increment)
	machine.SendEvent(Increment)

	// Check the status of the KeyCounter value
	fmt.Println("Final counter is at", machine.GetContext(KeyCounter).(int))

	// Now we can send the Deactivate Event as we're finished
	machine.SendEvent(Deactivate)
	fmt.Println("Final State is", machine.GetNameForState(machine.State()))

	// Output:
	// Initial state is 0
	// StateName for the current state is Inactive
	// Before incrementing Counter it's 0
	// Before setting the KeyIsReady it's false
	// Context KeyIsReady is true
	// Left Inactive entered Active Entry
	// After Activate Event machine State is Active
	// Final counter is at 3
	// Final State is Inactive
}
