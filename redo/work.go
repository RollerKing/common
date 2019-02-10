package redo

import (
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

// Job executable work
type Job func(*RedoCtx)

// RedoCtx redo context,control interval, stop now etc.
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

// SetDelayBeforeNext set interval before next loop, only once affective
func (ctx *RedoCtx) SetDelayBeforeNext(new_duration time.Duration) {
	ctx.delayBeforeNextLoop = new_duration
}

// StartNextRightNow start next loop right now
func (ctx *RedoCtx) StartNextRightNow() {
	ctx.SetDelayBeforeNext(time.Duration(0))
}

// StopRedo stop whole redo after this job done
func (ctx *RedoCtx) StopRedo() {
	ctx.stopRedo = true
}

// WrapFunc helper function convert normal func(){} to Job
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
				log.Printf("panic occur:%+v\nstacktrace:%s\n", r, string(buf))
			}
		}()
		once(ctx)
	}
	recipet := newRecipet()
	recipet.catchSignal = catchSignal
	go func(m *Recipet) {
		if catchSignal {
			batchCatchSignals(m.sigchan)
		}
		for {
			ctx := newCtx(duration)
			onceFunc(ctx)
			if ctx.stopRedo {
				m.Stop()
			}

			select {
			case <-m.requestStopChan():
				m.closeChannels()
				return
			case <-m.sigchan:
				signal.Stop(m.sigchan)
				m.stopWithRequest(STOP_SYS)
				m.closeChannels()
				return
			case <-time.After(ctx.delayBeforeNextLoop):
			case <-m.wakeupChan():
			}
		}
	}(recipet)
	return recipet
}

func batchCatchSignals(sigchan chan os.Signal) {
	signal.Notify(sigchan, syscall.SIGABRT, syscall.SIGALRM, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
}
