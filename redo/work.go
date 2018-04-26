package redo

import (
	"github.com/qjpcpu/log"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"
)

type Job func(*RedoCtx)

type JobState int

const (
	JobRunning JobState = iota
	JobStopping
	JobStopped
)

type Recipet struct {
	done     chan struct{}
	pls_exit chan struct{}
	state    JobState
	*sync.Mutex
}

type RedoCtx struct {
	delayBeforeNextLoop time.Duration
}

func newCtx(duration time.Duration) *RedoCtx {
	return &RedoCtx{delayBeforeNextLoop: duration}
}

func (ctx *RedoCtx) SetDelayBeforeNext(new_duration time.Duration) {
	ctx.delayBeforeNextLoop = new_duration
}

func (ctx *RedoCtx) StartNextRightNow() {
	ctx.SetDelayBeforeNext(time.Duration(0))
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

func (m *Recipet) Wait() {
	<-m.done
	if m.state == JobStopped {
		return
	}
	m.Lock()
	if m.state == JobStopping {
		m.state = JobStopped
		close(m.done)
	}
	m.Unlock()
}

func WrapFunc(work func()) Job {
	return func(ctx *RedoCtx) {
		work()
	}
}

func Perform(once Job, duration time.Duration) *Recipet {
	onceFunc := func(ctx *RedoCtx) {
		defer func() {
			if r := recover(); r != nil {
				buf := make([]byte, 1<<16)
				runtime.Stack(buf, false)
				log.Errorf("panic occur:%+v\nstacktrace:%s", r, string(buf))
			}
		}()
		once(ctx)
	}
	recipet := &Recipet{
		pls_exit: make(chan struct{}, 1),
		done:     make(chan struct{}, 1),
		state:    JobRunning,
		Mutex:    new(sync.Mutex),
	}
	go func(m *Recipet) {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, syscall.SIGABRT, syscall.SIGALRM, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
		for {
			ctx := newCtx(duration)
			onceFunc(ctx)

			select {
			case <-m.pls_exit:
				log.Debugf("user request exit")
				m.done <- struct{}{}
				close(m.pls_exit)
				return
			case s := <-sigchan:
				log.Debugf("get syscall signal %v", s)
				signal.Stop(sigchan)
				m.stopWithRequest(false)
				close(m.pls_exit)
				m.done <- struct{}{}
				return
			case <-time.After(ctx.delayBeforeNextLoop):
			}
		}
	}(recipet)
	return recipet
}
