/* fsm

A Finite State Machine with support for States, Transitions, Events and Handlers.

Inpsired by the likes of XState and similar FSM's this one isn't as extensive.
Instead it's focussed on being idiomatic go.

How it works

At its core an FSM is a way to manage the process from moving from one State to
another.

In "ojkelly.dev/fsm" you define your States as a const of fsm.State for example:

	const (
		StepZero fsm.State = iota
		StepOne
		StepTwo
		SetThree
	)

After that we define the Events that could be sent to change to a different State,
with consts of fsm.Event.

	const (
		Activate fsm.Event = iota
		Deactivate
		Increment
		Decrement
	)

Optionally we can also store some key, value pairs in the FSM itself. This is
useful for Guard functions that can prevent a State transition if conditions
are not met.

	const (
		KeyCounter fsm.ContextKey = iota
		KeyIsReady
	)

After that, we call fsm.New() and pass in what we defined above. Look at the
Counter example below, for the full code snippet.

A key bit to look at for the moment is how we define fsm.States. In the example
below we can see the state Inactive being defined, as responding only to the
event `Activate`. When the fsm.Machine receives that event, if it's in the
`Inactive` State it will attempt to transition to `Active`.

Before it can, the guard function will be called. If that returns `true`, then
the `Exit` function defined below will run. Followed by the `Entry` function
on `Active` if it exists. After that, the State will transition to `Active`.

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
					Exit: func(
						m *fsm.Machine,
						current fsm.State,
						next fsm.State,
						event fsm.TransitionEvent,
					) {
						m.Set(KeyIsReady, false)
					},
				},
			},
		},
		// ... Active State defintion not shown, see full example
	}

With the new fsm.Machine you can optionally add some maps to convert the State
const's into a string. This is helpful for debugging, but not required.

	machine.AddStateNames(stateNames)
	machine.AddEventNames(eventNames)
	machine.AddContextKeyNames(contextKeyNames)

Now we can start using it!

We can send events like this:

	machine.Event(Increment)

And update context values like this:

	machine.Set(KeyIsReady, true)


*/
package fsm // import "ojkelly.dev/fsm"
