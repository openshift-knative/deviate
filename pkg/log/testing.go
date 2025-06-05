package log

import "testing"

// TestingLogger implements a testing logger.
type TestingLogger struct {
	T testing.TB
}

func (t TestingLogger) Println(v ...interface{}) {
	t.T.Log(v...)
}

func (t TestingLogger) Printf(format string, v ...interface{}) {
	t.T.Logf(format, v...)
}

var _ Logger = TestingLogger{}
