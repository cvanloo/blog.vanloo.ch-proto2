package debug

import (
	"fmt"
	"runtime"
)

func Assert(v bool, msg string) {
	if !v {
		pc, f, l, _ := runtime.Caller(1)
		panic(fmt.Sprintf("assertion in %s[%s:%d] failed: %s", runtime.FuncForPC(pc).Name(), f, l, msg))
	}
}

func Must[T any](t T, err error) T {
	if err != nil {
		pc, f, l, _ := runtime.Caller(1)
		panic(fmt.Sprintf("must in %s[%s:%d] failed: %v", runtime.FuncForPC(pc).Name(), f, l, err))
	}
	return t
}

func PanicIf(err error) {
	if err != nil {
		pc, f, l, _ := runtime.Caller(1)
		panic(fmt.Sprintf("panicIf in %s[%s:%d] failed: %v", runtime.FuncForPC(pc).Name(), f, l, err))
	}
}

func Unreachable() {
	pc, f, l, _ := runtime.Caller(1)
	panic(fmt.Sprintf("unreachable in %s[%s:%d] failed", runtime.FuncForPC(pc).Name(), f, l))
}
