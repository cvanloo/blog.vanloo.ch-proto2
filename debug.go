package main

import (
	"fmt"
	"runtime"
)

func assert(v bool, msg string) {
	if !v {
		pc, f, l, _ := runtime.Caller(1)
		panic(fmt.Sprintf("assertion in %s[%s:%d] failed: %s", runtime.FuncForPC(pc).Name(), f, l, msg))
	}
}

func must[T any](t T, err error) T {
	if err != nil {
		pc, f, l, _ := runtime.Caller(1)
		panic(fmt.Sprintf("must in %s[%s:%d] failed: %v", runtime.FuncForPC(pc).Name(), f, l, err))
	}
	return t
}

func panicIf(err error) {
	if err != nil {
		pc, f, l, _ := runtime.Caller(1)
		panic(fmt.Sprintf("panicIf in %s[%s:%d] failed: %v", runtime.FuncForPC(pc).Name(), f, l, err))
	}
}

func unreachable() {
	pc, f, l, _ := runtime.Caller(1)
	panic(fmt.Sprintf("unreachable in %s[%s:%d] failed", runtime.FuncForPC(pc).Name(), f, l))
}
