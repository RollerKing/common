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
	recipet := newRecipet()
	go func(m *Recipet) {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, syscall.SIGABRT, syscall.SIGALRM, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
		pls_exit := m.requestStopChan()
		for {
			ctx := newCtx(duration)
			onceFunc(ctx)

			select {
			case <-pls_exit:
				log.Debugf("user request exit")
				m.closeChannels()
				return
			case s := <-sigchan:
				log.Debugf("get syscall signal %v", s)
				signal.Stop(sigchan)
				m.stopWithRequest(false)
				m.closeChannels()
				return
			case <-time.After(ctx.delayBeforeNextLoop):
			}
		}
	}(recipet)
	return recipet
}
