package fsm_test

import (
	"fmt"
	"testing"

	"github.com/ojkelly/control/fsm"
)

func Test_Counter(t *testing.T) {

	// States a machine can be in
	const (
		Inactive fsm.State = "Inactive"
		Active   fsm.State = "Active"
	)

	// Events that can change state
	const (
		Activate   fsm.Event = "Activate"
		Deactivate fsm.Event = "Deactivate"
		Increment  fsm.Event = "Increment"
		Decrement  fsm.Event = "Decrement"
	)

	// ContextKeys for storing extra state
	const (
		KeyCounter fsm.ContextKey = "Counter"
		KeyIsReady fsm.ContextKey = "IsReady"
	)

	errorHandler := func(m *fsm.Machine, current fsm.State, next fsm.State, event fsm.TransitionEvent) {
		fmt.Println("Error: Left", current, "entered", next, event)

		m.Event(Deactivate)
	}
	successHandler := func(m *fsm.Machine, current fsm.State, next fsm.State, event fsm.TransitionEvent) {
		fmt.Println("Success: Left", current, "entered", next)
	}
	logEvent := func(m *fsm.Machine, current fsm.State, next fsm.State, event fsm.TransitionEvent) {
		fmt.Println("Left", current, "entered", next, event)
	}
	guardActive := func(m *fsm.Machine, current fsm.State, next fsm.State) bool {
		if v := m.Get(KeyIsReady); v != nil {
			fmt.Println("guardActive", v)

			ready := v.(bool)
			fmt.Println("guardActive", current, "to", next, "is ready", ready)

			return ready
		}
		return false
	}

	machine := fsm.New(
		// machine ID
		"counterExample",
		// initial state
		Inactive,
		fsm.Context{
			KeyIsReady: fsm.ContextMeta{
				Write:  true, // can we use .Set()
				Inital: false,
			},
			KeyCounter: fsm.ContextMeta{
				Write:  false, // this can only be changed by events
				Inital: 0,
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
						Update: func(
							m *fsm.Machine,
							current fsm.State,
							next fsm.State,
							event fsm.TransitionEvent,
						) (
							key fsm.ContextKey,
							value interface{},
							err error,
						) {
							key = KeyCounter
							var count int
							if v := m.Get(KeyCounter); v != nil {

								count = v.(int) + 1
								fmt.Println("update", count)
								value = count
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
		errorHandler,
	)

	fmt.Printf("Current State: %v \n", machine.State())

	machine.Event(Increment)
	fmt.Printf("State after first Increment: %v \n", machine.State())

	machine.Set(KeyIsReady, true)
	machine.Event(Activate)
	fmt.Printf("State after first Activate: %v \n", machine.State())

	machine.Event(Increment)
	machine.Event(Increment)
	machine.Event(Increment)

	machine.Event(Deactivate)
	s := machine.State()
	fmt.Println(fmt.Errorf("state %v", s))
	if s != Inactive {
		fmt.Println("State should equal Inactive, got ", s)
		t.Fail()
	}
	if v := machine.Get(KeyCounter).(int); v != 3 {
		fmt.Println("Counter should equal 3, got ", v)
		t.Fail()
	}

	fmt.Printf("end of test %v\n", machine.Get(KeyCounter).(int))
}

func Test_TCPMachine(t *testing.T) {
	t.Skip()
	// // All the possible states for this FSM
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
