package joint

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"sync/atomic"
)

var Debug bool

const (
	closeI = 0
	readI  = 1
	writeI = 2
)

// Joint connect two channel
type Joint struct {
	list          *linkedList
	readC, writeC reflect.Value
	breakC        chan struct{}
	broken        int32
}

// Pipe two channel
func Pipe(readC interface{}, writeC interface{}) (*Joint, error) {
	if readC == nil || writeC == nil {
		return nil, errors.New("data channel should not be nil")
	}
	rv, wv, err := checkChan(readC, writeC)
	if err != nil {
		return nil, err
	}
	j := &Joint{
		readC:  rv,
		writeC: wv,
		breakC: make(chan struct{}, 1),
		list:   newList(),
	}
	go j.transport()
	return j, nil
}

// Breakoff conjuction
func (j *Joint) Breakoff() {
	if atomic.CompareAndSwapInt32(&j.broken, 0, 1) {
		close(j.breakC)
	}
}

/*
 * private methods
 */

func (j *Joint) transport() {
	cases := []reflect.SelectCase{
		{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(j.breakC),
		},
		{
			Dir:  reflect.SelectRecv,
			Chan: j.readC,
		},
	}
	cases2 := append(cases, reflect.SelectCase{
		Dir:  reflect.SelectSend,
		Chan: j.writeC,
	})
	var queueSize uint64
	var lastE, lastD interface{}
	if Debug {
		defer func() {
			log.Println("[joint] exited.")
		}()
	}
	var rClosed bool
	for {
		if queueSize == 0 {
			if rClosed {
				return
			}
			// list is empty
			_, recv, ok := reflect.Select(cases)
			if !ok {
				return
			}
			queueSize++
			cases2[writeI].Send = recv
			if Debug {
				lastE = recv.Interface()
				lastD = recv.Interface()
				log.Printf("[joint] Enqueue %v", lastE)
			}
		} else {
			chosen, recv, ok := reflect.Select(cases2)
			if chosen == writeI {
				// write ok
				queueSize--
				if Debug {
					log.Printf("[joint] Dequeue %v", lastD)
				}
				if queueSize > 0 {
					cases2[writeI].Send, _ = j.list.pop()
					if Debug {
						lastD = cases2[writeI].Send.Interface()
					}
				}
			} else {
				if !ok {
					if chosen == closeI {
						return
					} else {
						dumpC := reflect.ValueOf(make(chan struct{}, 1))
						cases[readI].Chan = dumpC
						cases2[readI].Chan = dumpC
						rClosed = true
						continue
					}
				}
				if chosen == readI {
					// read ok
					j.list.push(recv)
					queueSize++
					if Debug {
						lastE = recv.Interface()
						log.Printf("[joint] Enqueue %v", lastE)
					}
				}
			}

		}
	}
}

func checkChan(r interface{}, w interface{}) (rv reflect.Value, wv reflect.Value, err error) {
	rtp := reflect.TypeOf(r)
	wtp := reflect.TypeOf(w)
	if rtp.Kind() != reflect.Chan {
		err = errors.New("argument should be channel")
		return
	}
	if wtp.Kind() != reflect.Chan {
		err = errors.New("argument should be channel")
		return
	}
	if rtp.ChanDir() == reflect.SendDir {
		err = errors.New("read channel should be readable")
		return
	}
	if wtp.ChanDir() == reflect.RecvDir {
		err = errors.New("write channel should be writable")
		return
	}
	if rkind := rtp.Elem().Kind(); rkind != wtp.Elem().Kind() {
		err = fmt.Errorf("write channel element should be %v", rkind)
		return
	}
	if retp := rtp.Elem(); retp != wtp.Elem() {
		err = fmt.Errorf("write channel element should be %v", retp)
		return
	}
	return reflect.ValueOf(r), reflect.ValueOf(w), nil
}
