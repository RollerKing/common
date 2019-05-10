package joint

import (
	"errors"
	"fmt"
	"log"
	"math"
	"reflect"
	"sync/atomic"
	"time"
)

// Debug would print enquue/dequeue information
var Debug bool

const (
	reloadI = iota
	timeoutI
	closeI
	readI
	writeI
)

// PipeFilter filter data
type PipeFilter func(interface{}) bool

// Joint connect two channel
type Joint struct {
	list            *linkedList
	readC, writeC   reflect.Value
	breakC, reloadC chan struct{}
	broken          int32
	maxIn           uint64
	queueSize       uint64
	filter          PipeFilter
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
		readC:   rv,
		writeC:  wv,
		breakC:  make(chan struct{}, 1),
		reloadC: make(chan struct{}, 1),
		list:    newList(),
		maxIn:   math.MaxUint64 - 1,
	}
	go j.transport()
	return j, nil
}

// SetFilter of pipe
func (j *Joint) SetFilter(f PipeFilter) {
	j.filter = f
}

// SetCap set max pipe buffer size, can be ajust in runtime
func (j *Joint) SetCap(l uint64) error {
	chCap := uint64(j.readC.Cap() + j.writeC.Cap())
	min := chCap + 1
	if l < min {
		if Debug {
			log.Println("[joint] extend buffer size to", min)
		}
		l = min
	}
	max := uint64(math.MaxUint64 - 1)
	if l > max {
		return fmt.Errorf("[joint] length should not greater than %v", max)
	}
	maxIn := atomic.LoadUint64(&j.maxIn)
	if maxIn != l-chCap && atomic.LoadInt32(&j.broken) == 0 && atomic.CompareAndSwapUint64(&j.maxIn, maxIn, l-chCap) {
		j.reloadC <- struct{}{}
	}
	return nil
}

// Len return buffer length
func (j *Joint) Len() uint64 {
	return j.queueSize
}

// Cap return pipe cap
func (j *Joint) Cap() uint64 {
	return j.maxIn
}

// Breakoff halt conjuction, drop remain data in pipe
func (j *Joint) Breakoff() {
	if atomic.CompareAndSwapInt32(&j.broken, 0, 1) {
		close(j.breakC)
		close(j.reloadC)
	}
}

// DoneC return finished channel
func (j *Joint) DoneC() <-chan struct{} {
	return j.breakC
}

/*
 * private methods
 */

func (j *Joint) transport() {
	// add timer to prevent fatal error: all goroutines are asleep - deadlock!
	term := time.Hour * 1
	timer := time.NewTimer(term)
	cases := []reflect.SelectCase{
		{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(j.reloadC), // reload channel
		},
		{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(timer.C),
		},
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
	dummyC := reflect.ValueOf(make(chan struct{}, 1))
	var lastE, lastD interface{}
	defer func() {
		j.Breakoff()
		if Debug {
			log.Println("[joint] Exited.")
		}
	}()
	var rClosed bool
	for {
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		timer.Reset(term)
		if j.queueSize == 0 {
			if rClosed {
				return
			}
			// list is empty
			chosen, recv, ok := reflect.Select(cases)
			if !ok {
				return
			}
			if chosen == timeoutI || chosen == reloadI {
				continue
			}
			dataVal := recv.Interface()
			// drop data by filter
			if j.filter != nil && !j.filter(dataVal) {
				continue
			}
			j.queueSize++
			cases2[writeI].Send = recv
			if Debug {
				lastE = recv.Interface()
				lastD = recv.Interface()
				log.Printf("[joint] Enqueue %v", lastE)
			}
		} else {
			var chosen int
			var recv reflect.Value
			var ok bool
			if buff := atomic.LoadUint64(&j.maxIn); j.queueSize >= buff {
				cases2[readI].Chan = dummyC
				chosen, recv, ok = reflect.Select(cases2)
				// restore readC
				cases2[readI].Chan = j.readC
			} else {
				chosen, recv, ok = reflect.Select(cases2)
			}
			if chosen == timeoutI || chosen == reloadI {
				continue
			}
			if chosen == writeI {
				// write ok
				j.queueSize--
				if Debug {
					log.Printf("[joint] Dequeue %v", lastD)
				}
				for j.queueSize > 0 {
					val, _ := j.list.pop()
					if j.filter != nil && !j.filter(val.Interface()) {
						j.queueSize--
					} else {
						cases2[writeI].Send = val
						if Debug {
							lastD = val.Interface()
						}
						break
					}
				}
			} else {
				if !ok {
					if chosen == closeI {
						return
					} else {
						if Debug {
							log.Println("[joint] Input channel closed.")
						}
						cases[readI].Chan = dummyC
						cases2[readI].Chan = dummyC
						rClosed = true
						continue
					}
				}
				if chosen == readI {
					// read ok
					j.list.push(recv)
					j.queueSize++
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
