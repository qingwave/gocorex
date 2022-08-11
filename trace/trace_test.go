package trace

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/go-logr/logr/funcr"
)

var (
	logMsg     = ""
	testLogger = funcr.New(func(prefix, args string) {
		logMsg = fmt.Sprintln(prefix, args)
	}, funcr.Options{})
)

func TestTrace(t *testing.T) {
	trace := New("test", testLogger, Field{Key: "test", Value: "val"})

	trace.Step("step1")
	trace.Step("step2")

	trace.Log()

	if len(trace.traceItems) != 2 {
		t.Errorf("expected items %d, but got %d", 2, len(trace.traceItems))
	}

	if trace.traceItems[0].(traceStep).msg != "step1" {
		t.Errorf("expected msg %v", "step1")
	}

	if trace.TotalTime() == 0 {
		t.Errorf("got zero total time")
	}

	if !strings.Contains(logMsg, "step1") || !strings.Contains(logMsg, "step2") {
		t.Errorf("invaild log msg")
	}

	// log long not output
	{
		logMsg = ""
		trace.LogIfLong(1 * time.Second)
		if logMsg != "" {
			t.Errorf("LongIfLong should not output")
		}
	}

	// log long
	{
		time.Sleep(10 * time.Millisecond)
		trace.Step("long running step")
		logMsg = ""
		trace.LogIfLong(10 * time.Millisecond)

		if logMsg == "" {
			t.Errorf("LongIfLong should output")
		}
	}
}

func TestNestTrace(t *testing.T) {
	trace := New("test", testLogger)

	new := trace.Nest("nest1")

	logMsg = ""
	trace.Log()

	if len(trace.traceItems) != 1 {
		t.Errorf("expected items length %d, but got %d", 1, len(trace.traceItems))
	}

	if reflect.DeepEqual(trace.traceItems[0], *new) {
		t.Errorf("expected item %+v, but got %+v", *new, trace.traceItems[0])
	}

	if logMsg == "" {
		t.Errorf("log msg should output")
	}
}
