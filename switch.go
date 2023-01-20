package fsm

import (
	"context"
	"sync"
)

const (
	ON  = true
	OFF = false
)

type Switch struct {
	mutex   sync.RWMutex
	context map[bool]context.Context
	cancel  map[bool]context.CancelFunc
}

func NewSwitch(on bool) *Switch {
	onCtx, onCancel := context.WithCancel(context.Background())
	offCtx, offCancel := context.WithCancel(context.Background())

	s := &Switch{
		context: make(map[bool]context.Context),
		cancel:  make(map[bool]context.CancelFunc),
	}

	s.context[ON] = onCtx
	s.context[OFF] = offCtx
	s.cancel[ON] = onCancel
	s.cancel[OFF] = offCancel

	s.cancel[!on]()

	return s
}

func (s *Switch) Release() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, cancel := range s.cancel {
		cancel()
	}
	s.context = nil
	s.cancel = nil
}

func (s *Switch) isOn() bool {
	return s.context[ON].Err() == nil
}

func (s *Switch) trun(on bool) {
	if s.isOn() == on {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel[!on]()
	s.context[on] = ctx
	s.cancel[on] = cancel
}

func (s *Switch) IsOn() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.isOn()
}

func (s *Switch) TrueOn() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.trun(ON)
}

func (s *Switch) TrueOff() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.trun(OFF)
}

func (s *Switch) On() <-chan struct{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.context[OFF].Done()
}

func (s *Switch) Off() <-chan struct{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.context[ON].Done()
}
