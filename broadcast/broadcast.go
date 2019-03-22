package broadcast

import (
	"errors"
	"reflect"
)

type broadcast struct {
	c chan broadcast
	v interface{}
}

// Broadcaster notifier
type Broadcaster struct {
	// private fields:
	listenc   chan chan (chan broadcast)
	Sendc     chan<- interface{}
	valueType reflect.Type
}

// New create a new broadcaster object.
func New(meta interface{}) *Broadcaster {
	listenc := make(chan (chan (chan broadcast)))
	sendc := make(chan interface{})
	go func() {
		currc := make(chan broadcast, 1)
		for {
			select {
			case v, ok := <-sendc:
				if !ok {
					return
				}
				c := make(chan broadcast, 1)
				b := broadcast{c: c, v: v}
				currc <- b
				currc = c
			case r := <-listenc:
				r <- currc
			}
		}
	}()
	return &Broadcaster{
		listenc:   listenc,
		Sendc:     sendc,
		valueType: reflect.TypeOf(meta),
	}
}

// Notify start listening to the broadcasts.
func (b *Broadcaster) Notify(typpedChan interface{}) error {
	tp := reflect.TypeOf(typpedChan)
	outC := reflect.ValueOf(typpedChan)
	if tp.Kind() != reflect.Chan {
		return errors.New("input parameter should be channel")
	}
	if tp.ChanDir() == reflect.RecvDir {
		return errors.New("channel should be writable")
	}
	if tp.Elem() != b.valueType {
		return errors.New("bad channel value type")
	}
	c := make(chan chan broadcast)
	b.listenc <- c
	rc := <-c
	go func() {
		defer func() {
			_ = recover()
		}()
		for {
			b := <-rc
			v := b.v
			rc <- b
			rc = b.c
			outC.Send(reflect.ValueOf(v))
		}
	}()
	return nil
}

// Send broadcast a value to all listeners.
func (b *Broadcaster) Send(v interface{}) { b.Sendc <- v }

// Stop broadcast
func (b *Broadcaster) Stop() {
	close(b.Sendc)
}
