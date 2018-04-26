package redo

import (
	"github.com/qjpcpu/log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Job func() error
type Recipet struct {
	done     chan struct{}
	pls_exit chan struct{}
}

func (m *Recipet) Stop() {
	m.pls_exit <- struct{}{}
}

func (m *Recipet) Wait() {
	<-m.done
	close(m.done)
}

func Perform(once Job, duration time.Duration) *Recipet {
	onceFunc := func() {
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("panic occur:%+v", r)
			}
		}()
		if err := once(); err != nil {
			log.Debugf("error occur:%v", err)
		}
	}
	recipet := &Recipet{
		pls_exit: make(chan struct{}, 1),
		done:     make(chan struct{}, 1),
	}
	go func(m *Recipet) {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, syscall.SIGABRT, syscall.SIGALRM, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
		for {
			onceFunc()

			select {
			case <-time.After(duration):
			case <-m.pls_exit:
				log.Info("user request exit")
				m.done <- struct{}{}
				close(m.pls_exit)
				return
			case s := <-sigchan:
				log.Infof("get syscall signal %v", s)
				signal.Stop(sigchan)
				m.done <- struct{}{}
				return
			}
		}
	}(recipet)
	return recipet
}
