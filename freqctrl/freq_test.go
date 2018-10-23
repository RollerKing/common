package freqctrl

import (
	"github.com/alicebob/miniredis"
	"github.com/qjpcpu/common/redisutil"
	"testing"
)

func TestFreq(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	var thr, win_sec int64 = 5, 1
	freq, err := New(
		redisutil.CreatePool(s.Addr(), "", ""),
		thr,
		win_sec,
	)
	if err != nil {
		t.Fatal(err)
	}
	if freq.Check("user1", "rule1") > 1 {
		t.Fatal("user1 should not excceed")
	}
	if freq.Check("user2", "rule1") > 1 {
		t.Fatal("user2 should not excceed")
	}
	for i := 0; i < 6; i++ {
		level := freq.Tick("user1", "rule1")
		t.Logf("user1 level:%v", level)
	}
	if freq.Check("user1", "rule1") <= 1 {
		t.Fatal("user1 should  excceed")
	}
	if freq.Check("user1", "another-rule") > 1 {
		t.Fatal("user1 another-rule should not  excceed")
	}
	if freq.Check("user2", "rule1") > 1 {
		t.Fatal("user2 should not excceed")
	}
}

func TestSet(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	set := NewFreqCtrlSet(s.Addr(), "", "", "test-ns")
	name := "10 times in 1 second"
	set.SetCtrl(name, 10, 1)
	if set.GetCtrl(name).Check("user1", "rule1") > 1 {
		t.Fatal("should not overflow")
	}
	for i := 0; i < 11; i++ {
		set.GetCtrl(name).Tick("user1", "rule1")
	}
	if set.GetCtrl(name).Check("user1", "rule1") <= 1 {
		t.Fatal("should  overflow")
	}
}
