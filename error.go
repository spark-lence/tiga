package tiga

import (
	"fmt"
	"runtime"
	"strings"
)

type ErrorWarp struct {
	File  string `json:"file"`
	Stack string `json:"stack"`
	Func  string `json:"func"`
	Err   error  `json:"err"`
}

func NewErrorWarp(err error, stack string) ErrorWarp {
	DebugStack := stack
	for _, v := range strings.Split(DebugStack, "\n") {
		DebugStack += v
	}

	// err := fmt.Errorf("%s", msg)
	// 取上一帧栈
	pc, file, lineNo, _ := runtime.Caller(1)
	f := runtime.FuncForPC(pc)
	warp := ErrorWarp{
		File:  fmt.Sprintf("%s:%d", file, lineNo),
		Stack: DebugStack,
		Func:  f.Name(),
		Err:   err,
	}
	return warp
}
func (e ErrorWarp) Error() string {
	return fmt.Sprintf("%s:%s:%s:%s", e.File, e.Func, e.Err.Error(), e.Stack)
}
func CheckMySQLDuplicateError(err error) bool {
	if err != nil && strings.Contains(err.Error(), "Duplicate entry") {
		return true
	}
	return false
}


type Errors struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	ErrMsg string `json:"err_msg"`
	

}