package redo

import (
	"github.com/qjpcpu/log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

type Job func(*RedoCtx)

type Recipet struct {
	done     chan struct{}
	pls_exit chan struct{}
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

func (m *Recipet) Stop() {
	m.pls_exit <- struct{}{}
}

func (m *Recipet) Wait() {
	<-m.done
	close(m.done)
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
				m.done <- struct{}{}
				return
			case <-time.After(ctx.delayBeforeNextLoop):
			}
		}
	}(recipet)
	return recipet
}
