package fsm

import "fmt"

type ContextKey string

type Context map[ContextKey]ContextMeta

type ContextMeta struct {
	Write  bool
	Inital interface{}
}

type contextMeta struct {
	key   ContextKey
	write bool
	value interface{}
}

type internalContext map[ContextKey]*contextMeta

type ContextUpdateHandler func(
	m *Machine,
	current State,
	next State,
	event TransitionEvent,
) (
	key ContextKey,
	value interface{},
	err error,
)

func (m *Machine) Set(key ContextKey, value interface{}) {
	m.checkIfCreatedCorrectly()
	// m.mtx.Lock()
	// defer m.mtx.Unlock()

	if v := m.context[key]; v != nil {
		if !v.write {
			panic(fmt.Sprintf("[%s] fsm.Set tried to set value '%v' for protect ContextKey '%s'", m.id, value, key))
		}

		v.value = value
		m.context[key] = v
	}
}

func (m *Machine) Get(key ContextKey) interface{} {
	m.checkIfCreatedCorrectly()

	// m.mtx.RLock()
	// defer m.mtx.RUnlock()

	if v, ok := m.context[key]; ok && v != nil {
		return v.value
	}

	return nil
}
