package fsm

type TransitionEvent string

const (
	TransitionEventEntry   TransitionEvent = "Entry"
	TransitionEventExit    TransitionEvent = "Exit"
	TransitionEventSuccess TransitionEvent = "Success"
)

type TransitionEventHandler func(m *Machine, current State, next State, event TransitionEvent)
type Guard func(m *Machine, current State, next State) bool

type Transition struct {
	// State to transition to
	State State
	// Guard hook can prevent this transition
	Guard Guard
	// Entry called when transitioning to this State
	Entry TransitionEventHandler
	// Exit called when leaving this State for another
	Exit TransitionEventHandler

	// ContextUpdate allows to update protected context values in response
	// to an Event
	ContextUpdate ContextUpdateHandler
}
