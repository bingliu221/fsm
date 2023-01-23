package fsm

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrStateExists    = errors.New("state exists")
	ErrStateNotExists = errors.New("state does not exist")
)

type Machine[State comparable] struct {
	mutex  sync.RWMutex
	states map[State]*Switch
}

func (m *Machine[State]) addState(s State, init bool) error {
	if _, ok := m.states[s]; ok {
		return ErrStateExists
	}

	m.states[s] = NewSwitch(init)

	return nil
}

func New[State comparable](init State, otherStates ...State) (*Machine[State], error) {
	m := &Machine[State]{
		states: make(map[State]*Switch),
	}

	err := m.addState(init, true)
	if err != nil {
		return nil, err
	}
	for _, s := range otherStates {
		err := m.addState(s, false)
		if err != nil {
			return nil, err
		}
	}

	return m, nil
}

func (m *Machine[State]) Release() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for _, sw := range m.states {
		sw.Release()
	}
	m.states = nil
}

func (m *Machine[State]) TransitTo(s State) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	sw, ok := m.states[s]
	if !ok {
		return ErrStateNotExists
	}
	sw.TrueOn()

	for _s, sw := range m.states {
		if _s != s {
			sw.TrueOff()
		}
	}

	return nil
}

func (m *Machine[State]) waitForState(ctx context.Context, s State, on bool) error {
	m.mutex.RLock()
	sw, ok := m.states[s]
	if !ok {
		m.mutex.RUnlock()
		return ErrStateNotExists
	}
	var c <-chan struct{}
	if on {
		c = sw.On()
	} else {
		c = sw.Off()
	}
	m.mutex.RUnlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c:
		return nil
	}
}

func (m *Machine[State]) WaitForState(ctx context.Context, s State) error {
	return m.waitForState(ctx, s, ON)
}

func (m *Machine[State]) WaitForStateToPass(ctx context.Context, s State) error {
	return m.waitForState(ctx, s, OFF)
}

func (m *Machine[State]) CurrentState() State {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for s, sw := range m.states {
		if sw.IsOn() {
			return s
		}
	}

	panic("machine without state")
}
