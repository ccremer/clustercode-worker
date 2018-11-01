package util

import (
    "fmt"
)

func PanicWithMessage(formatString string, args ...interface{}) {
    panic(fmt.Sprintf(formatString, args))
}

func PanicOnErrorf(formatString string, err error, args ...interface{}) {
    if err != nil {
        panic(fmt.Sprintf(formatString, err, args))
    }
}

func PanicOnError(err error) {
    if err != nil {
        panic(err)
    }
}

