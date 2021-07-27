package fsm

type MachineError string

const (
	MachineErrorUnknown         MachineError = "MachineErrorUnknown"
	MachineErrorExternal        MachineError = "MachineErrorExternal" // fsm.Error()
	MachineErrorTransitionEvent MachineError = "MachineErrorTransitionEvent"
	MachineErrorContextUpdate   MachineError = "MachineErrorContextUpdate"
)

type MachineErrorHandler func(m *Machine, current State, next State, machineError MachineError)

func (m *Machine) handleError(e error, machineError MachineError) {
	currentState := m.state
	node := m.states[currentState]

	handler := node.Error
	if handler != nil {
		handler(m, currentState, currentState, machineError)
		return
	}

	m.errorHandler(m, currentState, currentState, machineError)
}

// Error is called by you when a state encounters an error
// the FSM will check and see if there is an error handler
// otherwise it will bubble the error up to the top level
// error handler
func (m *Machine) Error(e error) {
	m.checkIfCreatedCorrectly()

	m.handleError(e, MachineErrorExternal)
}
