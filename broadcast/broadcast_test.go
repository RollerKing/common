package broadcast

import (
	"testing"
)

func TestB(t *testing.T) {
	type Msg struct {
		ID int
	}
	b := New(Msg{})
	watchers := []chan Msg{make(chan Msg, 1), make(chan Msg, 1), make(chan Msg, 1)}
	for _, w := range watchers {
		b.Notify(w)
	}
	id := 522
	b.Send(Msg{ID: id})
	for _, w := range watchers {
		if v := <-w; v.ID != id {
			t.Fatal("error")
		}
	}
}
