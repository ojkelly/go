package fsm

import "fmt"

type ContextKey int

// ContextKeyNames are optional, but useful for debugging and will print out in
// error messages
type ContextKeyNames map[ContextKey]string

// AddEventNames will add Event names for debugging.
// This can only be called once, and does not affect the functioning of the
// machine.
func (m *Machine) AddContextKeyNames(c ContextKeyNames) {
	m.checkIfCreatedCorrectly()

	if m.hasSetContextKeyNames {
		return
	}

	m.hasSetContextKeyNames = true
	m.contextKeyNames = c
}

// GetNameForContextKey will return the name set in ContextKeyNames for the given
// ContextKey, or it will return the ContextKey int as a string
func (m *Machine) GetNameForContextKey(s ContextKey) string {
	if sn, ok := m.contextKeyNames[s]; ok {
		return sn
	}
	return fmt.Sprintf("%d", s)
}

// Context is used to store extra state that is considered when transitioning
// between states
type Context map[ContextKey]ContextMeta

type ContextMeta struct {
	Protected bool
	Inital    interface{}
}

type contextMeta struct {
	key       ContextKey
	protected bool
	value     interface{}
}

type internalContext map[ContextKey]*contextMeta

// SetContext for a given key with value
func (m *Machine) SetContext(key ContextKey, value interface{}) {
	m.checkIfCreatedCorrectly()

	if v := m.context[key]; v != nil {
		if v.protected {
			panic(
				fmt.Sprintf(
					"[%s] fsm.Set tried to set value '%v' for protected ContextKey '%s'",
					m.id,
					value,
					m.GetNameForContextKey(key),
				),
			)
		}

		v.value = value
		m.context[key] = v
	}
}

// GetContext the value for a given ContextKey
// it's up to you to know what the type is and cast it
//
// 	isReady = machine.GetContext(KeyIsReady).(bool)
func (m *Machine) GetContext(key ContextKey) interface{} {
	m.checkIfCreatedCorrectly()
	v, ok := m.context[key]

	if !ok {
		if m.errorHandler != nil {
			m.errorHandler(m, m.state, m.state, MachineErrorGuardFail)
		}
		return nil
	}

	if v != nil {
		return v.value
	}

	return nil
}

// UpdateContext returned from UpdateContextHandler is a map of which
// key, value pairs in Context to update
type UpdateContext map[ContextKey]interface{}

type UpdateContextHandler func(
	m *Machine,
	current State,
	next State,
	event TransitionEvent,
) (
	update UpdateContext,
	err error,
)

func (m *Machine) handleUpdateContext(t Transition, currentState State) {
	update, err := t.UpdateContext(m, currentState, t.State, TransitionEventEntry)

	if err != nil {
		m.handleError(err, MachineErrorUpdateContext)
	}

	for key, value := range update {
		if value != nil {
			v := m.context[key]
			if v != nil {
				m.context[key] = &contextMeta{
					key:       key,
					protected: v.protected,
					value:     value,
				}
			} else {
				panic(fmt.Sprintf("[%s] You tried to update an unregistered ContextKey. Register it first in fsm.New()", m.id))
			}
		}
	}
}
