package pubsub

import (
	"fmt"
	"testing"
	"time"
)

func TestIsParent(t *testing.T) {
	var tests = []struct {
		tp, ts string
		want   bool
	}{
		{"a", "b", false},
		{"b", "a", false},
		{"a", "a", true},
		{"a.c", "a", false},
		{"a.c", "c", false},
		{"a.c", "a.c", true},
		{"a", "a.c", true},
	}

	for _, tt := range tests {

		testname := fmt.Sprintf("%s,%s", tt.tp, tt.ts)
		t.Run(testname, func(t *testing.T) {
			ans := isParent(tt.tp, tt.ts)
			if ans != tt.want {
				t.Errorf("got %t, want %t", ans, tt.want)
			}
		})
	}
}

func TestPubSubString(t *testing.T) {
	var tests = []struct {
		tp, ts, msg string
		want        string
	}{
		{"a", "a", "test", "test"},
		{"b", "a", "test", "TIMEOUT"},
		{"a", "a.c", "test", "test"},
		{"a.c", "a.c", "test", "test"},
	}

	for _, tt := range tests {

		testname := fmt.Sprintf("%s,%s,%s", tt.tp, tt.ts, tt.msg)
		t.Run(testname, func(t *testing.T) {
			ans := psSubInteraction(tt.tp, tt.ts, tt.msg)
			if ans != tt.want {
				t.Errorf("got %s, want %s", ans, tt.want)
			}
		})
	}
}

func TestPubSubInt(t *testing.T) {
	var tests = []struct {
		tp, ts string
		msg    int
		want   string
	}{
		{"a", "a", 42, "42"},
		{"b", "a", 42, "TIMEOUT"},
		{"a", "a.c", 42, "42"},
		{"a.c", "a.c", 42, "42"},
	}

	for _, tt := range tests {

		testname := fmt.Sprintf("%s,%s,%d", tt.tp, tt.ts, tt.msg)
		t.Run(testname, func(t *testing.T) {
			ans := psSubInteraction(tt.tp, tt.ts, tt.msg)
			if ans != tt.want {
				t.Errorf("got %s, want %s", ans, tt.want)
			}
		})
	}
}

func TestPubUnsub(t *testing.T) {
	var tests = []struct {
		tp, ts, msg string
		unsub       bool
		want        string
	}{
		{"a", "a", "test", false, "testtest"},
		{"a", "a", "test", true, "test"},
	}

	for _, tt := range tests {

		testname := fmt.Sprintf("%s,%s,%s,%v", tt.tp, tt.ts, tt.msg, tt.unsub)
		t.Run(testname, func(t *testing.T) {
			ans := psUnsubInteraction(tt.tp, tt.ts, tt.msg, tt.unsub)
			if ans != tt.want {
				t.Errorf("got %s, want %s", ans, tt.want)
			}
		})
	}
}

func psSubInteraction(tp string, ts string, msg interface{}) string {
	var ps PubSub
	clk := make(chan bool)
	ch := make(chan interface{})
	timeout := make(chan bool, 1)

	go func() {
		time.Sleep(25 * time.Millisecond)
		timeout <- true
	}()

	go func() {
		psch := ps.Subscribe(ts)
		clk <- true
		ch <- <-psch
	}()

	<-clk
	ps.Publish(tp, msg)

	select {
	case a := <-ch:
		return fmt.Sprintf("%v", a)
	case <-timeout:
		return "TIMEOUT"
	}
}

func psUnsubInteraction(tp string, ts string, msg string, unsub bool) string {
	var ps PubSub
	clk := make(chan bool)
	ch := make(chan interface{})
	timeout := make(chan bool, 1)
	var ans string

	go func() {
		time.Sleep(25 * time.Millisecond)
		timeout <- true
	}()

	go func() {
		psch := ps.Subscribe(ts)
		clk <- true
		ch <- <-psch
		if unsub {
			ps.Unsubscribe(psch)
		}
		clk <- true
		ch <- <-psch
	}()

	<-clk
	ps.Publish(tp, msg)
	select {
	case a := <-ch:
		ans += fmt.Sprintf("%v", a)
	case <-timeout:
		ans += "TIMEOUT"
	}

	<-clk
	ps.Publish(tp, msg)

	select {
	case a := <-ch:
		if !(a == nil) {
			ans += fmt.Sprintf("%v", a)
		}
	case <-timeout:
		ans += "TIMEOUT"
	}
	return ans
}

func Example() {
	var ps PubSub

	go func() {
		for {
			ps.Publish("topic1", "test")
			time.Sleep(1 * time.Second)
		}
	}()

	go func() {
		for {
			ps.Publish("topic2", "example")
			time.Sleep(2 * time.Second)
		}
	}()

	ch := ps.Subscribe("topic1.subtopic")
	for i := 0; i < 10; i++ {
		msg := <-ch
		fmt.Println("topic1.subtopic", msg)
	}
	ps.Unsubscribe(ch)
}
