package redo

import (
	"sync"
)

type JobState int

const (
	JobRunning JobState = iota
	JobStopping
)

type Recipet struct {
	done     chan struct{}
	pls_exit chan struct{}
	state    JobState
	*sync.Mutex
}

func newRecipet() *Recipet {
	return &Recipet{
		pls_exit: make(chan struct{}, 1),
		done:     make(chan struct{}, 1),
		state:    JobRunning,
		Mutex:    new(sync.Mutex),
	}
}

func (m *Recipet) Stop() bool {
	return m.stopWithRequest(true)
}

func (m *Recipet) stopWithRequest(with bool) bool {
	var op bool = false
	if m.state != JobRunning {
		return op
	}
	m.Lock()
	if m.state == JobRunning {
		m.state = JobStopping
		if with {
			m.pls_exit <- struct{}{}
		}
		op = true
	}
	m.Unlock()
	return op
}

func (m *Recipet) closeChannels() {
	close(m.pls_exit)
	close(m.done)
}

func (m *Recipet) requestStopChan() <-chan struct{} {
	return m.pls_exit
}

func (m *Recipet) WaitChan() <-chan struct{} {
	return m.done
}

func (m *Recipet) Wait() {
	<-m.WaitChan()
}
