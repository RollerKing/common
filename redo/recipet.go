package redo

import (
	"os"
	"sync"
)

// StopType stop type
type StopType string

const (
	// STOP_SYS sys stop
	STOP_SYS StopType = "SYS"
	// STOP_USER user stop
	STOP_USER = "USER"
)

// JobState job state
type JobState int

const (
	// JobRunning job is running
	JobRunning JobState = iota
	// JobStopping job is stopped
	JobStopping
)

// Recipet handler of job
type Recipet struct {
	done        chan struct{}
	plsExit     chan struct{}
	wakeup      chan struct{}
	sigchan     chan os.Signal
	catchSignal bool
	state       JobState
	stopType    StopType // signal or user request stop
	*sync.Mutex
}

func newRecipet() *Recipet {
	return &Recipet{
		plsExit:     make(chan struct{}, 1),
		wakeup:      make(chan struct{}, 1),
		done:        make(chan struct{}, 1),
		state:       JobRunning,
		sigchan:     make(chan os.Signal, 1),
		catchSignal: false,
		Mutex:       new(sync.Mutex),
	}
}

// Stop stop job,return means ok/already stopped
func (m *Recipet) Stop() bool {
	return m.stopWithRequest(STOP_USER)
}

// Wakeup wake up right now
func (m *Recipet) Wakeup() bool {
	if m.state != JobRunning {
		return false
	}
	select {
	case m.wakeup <- struct{}{}:
		return true
	default:
		return false
	}
}

func (m *Recipet) stopWithRequest(stopType StopType) bool {
	var op = false
	if m.state != JobRunning {
		return op
	}
	m.Lock()
	if m.state == JobRunning {
		m.state = JobStopping
		if stopType == STOP_USER {
			m.plsExit <- struct{}{}
		}
		m.stopType = stopType
		op = true
	}
	m.Unlock()
	return op
}

func (m *Recipet) closeChannels() {
	close(m.plsExit)
	close(m.done)
	close(m.wakeup)
}

func (m *Recipet) requestStopChan() <-chan struct{} {
	return m.plsExit
}

func (m *Recipet) wakeupChan() <-chan struct{} {
	return m.wakeup
}

// WaitChan channel would get msg when job done
func (m *Recipet) WaitChan() <-chan struct{} {
	return m.done
}

// Wait wait job done
func (m *Recipet) Wait() StopType {
	<-m.WaitChan()
	return m.stopType
}

// Concat combine serveral recipets into single one
func (m *Recipet) Concat(others ...*Recipet) *CombiRecipt {
	rs := append([]*Recipet{m}, others...)
	return NewCombiRecipt(rs...)
}

// CombiRecipt merged from serveral recipt
type CombiRecipt struct {
	recipets []*Recipet
	alldone  chan struct{}
	once     *sync.Once
}

// NewCombiRecipt create combirecipt
func NewCombiRecipt(list ...*Recipet) *CombiRecipt {
	var unsafeRecipets []*Recipet
	for i, r := range list {
		if !r.catchSignal {
			unsafeRecipets = append(unsafeRecipets, list[i])
		}
	}
	if len(unsafeRecipets) > 0 && len(unsafeRecipets) < len(list) {
		for i := range unsafeRecipets {
			unsafeRecipets[i].catchSignal = true
			batchCatchSignals(unsafeRecipets[i].sigchan)
		}
	}
	return &CombiRecipt{
		recipets: list,
		alldone:  make(chan struct{}, 1),
		once:     new(sync.Once),
	}
}

// Stop stop all job
func (cr *CombiRecipt) Stop() bool {
	var ok bool
	for _, r := range cr.recipets {
		ok = r.Stop()
	}
	return ok
}

// Wait wait all job done
func (cr *CombiRecipt) Wait() StopType {
	<-cr.WaitChan()
	return cr.recipets[0].stopType
}

// WaitChan get msg when all job done
func (cr *CombiRecipt) WaitChan() <-chan struct{} {
	cr.once.Do(func() {
		go func() {
			for _, r := range cr.recipets {
				r.Wait()
			}
			close(cr.alldone)
		}()
	})
	return cr.alldone
}
