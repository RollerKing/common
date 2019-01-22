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
	pls_exit    chan struct{}
	sigchan     chan os.Signal
	catchSignal bool
	state       JobState
	stop_type   StopType // signal or user request stop
	*sync.Mutex
}

func newRecipet() *Recipet {
	return &Recipet{
		pls_exit:    make(chan struct{}, 1),
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

func (m *Recipet) stopWithRequest(stop_type StopType) bool {
	var op = false
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

// WaitChan channel would get msg when job done
func (m *Recipet) WaitChan() <-chan struct{} {
	return m.done
}

// Wait wait job done
func (m *Recipet) Wait() StopType {
	<-m.WaitChan()
	return m.stop_type
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
	return cr.recipets[0].stop_type
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
