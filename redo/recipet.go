package redo

import (
	"sync"
)

type StopType string

const (
	STOP_SYS  StopType = "SYS"
	STOP_USER          = "USER"
)

type JobState int

const (
	JobRunning JobState = iota
	JobStopping
)

type Recipet struct {
	done      chan struct{}
	pls_exit  chan struct{}
	state     JobState
	stop_type StopType // signal or user request stop
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
	return m.stopWithRequest(STOP_USER)
}

func (m *Recipet) stopWithRequest(stop_type StopType) bool {
	var op bool = false
	if m.state != JobRunning {
		return op
	}
	m.Lock()
	if m.state == JobRunning {
		m.state = JobStopping
		if stop_type == STOP_USER {
			m.pls_exit <- struct{}{}
		}
		m.stop_type = stop_type
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

func (m *Recipet) Wait() StopType {
	<-m.WaitChan()
	return m.stop_type
}
