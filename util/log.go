package util

import (
    "fmt"
    "github.com/aellwein/slf4go"
)

func PanicWithMessage(formatString string, args ...interface{}) {
    if args == nil {
        panic(formatString)
    } else {
        panic(fmt.Sprintf(formatString, args...))
    }
}

func PanicOnErrorf(formatString string, err error, args ...interface{}) {
    if err == nil {
        return
    }
    if args == nil {
        panic(fmt.Sprintf(formatString, err))
    } else {
        panic(fmt.Sprintf(formatString, err, args))
    }
}

func PanicOnError(err error) {
    if err != nil {
        panic(err)
    }
}

func StringToLogLevel(level string) slf4go.LogLevel {
    switch level {
    case "debug":
        return slf4go.LevelDebug
    case "info":
        return slf4go.LevelInfo
    case "warn":
    case "warning":
        return slf4go.LevelWarn
    case "error":
        return slf4go.LevelError
    }
    return slf4go.LevelInfo
}
