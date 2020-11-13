package tools

import (
	"github.com/aliyun/aliyun-ahas-go-sdk/logger"
	"runtime"
)

func PrintPanicStack() {
	if e := recover(); e != nil {
		buf := make([]byte, 1<<11)
		len := runtime.Stack(buf, true)
		logger.Warnf("Panic ==> %s", string(buf[:len]))
	}
}

func PrintPanicStackV2(flag string) {
	if e := recover(); e != nil {
		buf := make([]byte, 1<<11)
		len := runtime.Stack(buf, true)
		logger.Warnf("panic happens, %s:  %s\n", flag, string(buf[:len]))
	}
}
