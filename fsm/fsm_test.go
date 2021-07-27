package fsm_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"ojkelly.dev/fsm"
)

func Test_Counter(t *testing.T) {

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
		m.Event(Deactivate)
	}

	successHandler := func(m *fsm.Machine, current fsm.State, next fsm.State, event fsm.TransitionEvent) {
		fmt.Println("Success: Left", m.GetNameForState(current), "entered", m.GetNameForState(next))
	}

	logEvent := func(m *fsm.Machine, current fsm.State, next fsm.State, event fsm.TransitionEvent) {
		fmt.Println("Left", m.GetNameForState(current), "entered", m.GetNameForState(next), event)
	}

	guardActive := func(m *fsm.Machine, current fsm.State, next fsm.State) bool {
		if v := m.Get(KeyIsReady); v != nil {
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
							m.Set(KeyIsReady, false)
						},
					},
				},
			},

			// Active state
			Active: fsm.StateNode{
				Error:   errorHandler,
				Success: successHandler,
				Events: fsm.EventToTransition{
					Increment: fsm.Transition{
						State: Active,
						ContextUpdate: func(
							m *fsm.Machine,
							current fsm.State,
							next fsm.State,
							event fsm.TransitionEvent,
						) (
							update fsm.ContextUpdate,
							err error,
						) {
							update = fsm.ContextUpdate{}
							if v := m.Get(KeyCounter); v != nil {
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

	assert.Equal(t, machine.State(), Inactive, "initial state should be Inactive")

	assert.Equal(
		t,
		machine.GetNameForState(machine.State()),
		stateNames[Inactive],
		"our state names were set correctly",
	)

	// Try to increment - nothing should happen as we're Inactive at the moment
	machine.Event(Increment)
	counter := machine.Get(KeyCounter).(int)
	assert.Equal(t, counter, 0, "our counter should still be 0")

	isReady := machine.Get(KeyIsReady).(bool)
	assert.Equal(t, isReady, false, "The machine shouldn't be ready yet")

	machine.Set(KeyIsReady, true)
	isReady = machine.Get(KeyIsReady).(bool)

	assert.Equal(t, isReady, true, "context isReady should be true now")

	machine.Event(Activate)
	assert.Equal(t, machine.State(), Active, "machine state should be Active")

	// Increment 3 times
	machine.Event(Increment)
	machine.Event(Increment)
	machine.Event(Increment)

	counter = machine.Get(KeyCounter).(int)
	assert.Equal(t, counter, 3, "our counter should now be 3")

	machine.Event(Deactivate)
	assert.Equal(t, machine.State(), Inactive, "machine state should now be Inactive")

}

func Test_TCPMachine(t *testing.T) {

	// All the possible states for this FSM
	// const (
	// 	NoConnection          fsm.State = "NoConnection"
	// 	ConnectionEstablished fsm.State = "ConnectionEstablished"
	// 	// A mock TCP handshake
	// 	SendACK    fsm.State = "SendACK"
	// 	RecieveSYN fsm.State = "RecieveSYN"
	// 	SendSYNACK fsm.State = "SendSYNACK"
	// )

	// const (
	// 	RemoteIp fsm.ContextKey = "RemoteIp"
	// )

	// machine := fsm.New(NoConnection,
	// 	fsm.NewContextKeys(
	// 		RemoteIp,
	// 	))

	// machine.Set("remoteIp", "0.0.0.0")

	// fmt.Printf("Current State: %v \n", machine.State())

	// fmt.Printf("end of test %v\n", machine.Get("remoteIp").(string))

	// t.Fail()
}
