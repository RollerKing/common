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
	stopRedo            bool
}

func newCtx(duration time.Duration) *RedoCtx {
	return &RedoCtx{
		delayBeforeNextLoop: duration,
		stopRedo:            false,
	}
}

func (ctx *RedoCtx) SetDelayBeforeNext(new_duration time.Duration) {
	ctx.delayBeforeNextLoop = new_duration
}

func (ctx *RedoCtx) StartNextRightNow() {
	ctx.SetDelayBeforeNext(time.Duration(0))
}

func (ctx *RedoCtx) StopRedo() {
	ctx.stopRedo = true
}

func WrapFunc(work func()) Job {
	return func(ctx *RedoCtx) {
		work()
	}
}

// perform job without gracefull exit
func Perform(once Job, duration time.Duration) *Recipet {
	return performWork(once, duration, false)
}

// perform job with gracefull exit
func PerformSafe(once Job, duration time.Duration) *Recipet {
	return performWork(once, duration, true)
}

func performWork(once Job, duration time.Duration, catchSignal bool) *Recipet {
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
		if catchSignal {
			signal.Notify(sigchan, syscall.SIGABRT, syscall.SIGALRM, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
		}
		pls_exit := m.requestStopChan()
		for {
			ctx := newCtx(duration)
			onceFunc(ctx)
			if ctx.stopRedo {
				m.Stop()
			}

			select {
			case <-pls_exit:
				log.Debugf("user request exit")
				m.closeChannels()
				return
			case s := <-sigchan:
				log.Debugf("get syscall signal %v", s)
				signal.Stop(sigchan)
				m.stopWithRequest(STOP_SYS)
				m.closeChannels()
				return
			case <-time.After(ctx.delayBeforeNextLoop):
			}
		}
	}(recipet)
	return recipet
}
