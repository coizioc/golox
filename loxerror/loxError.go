package loxerror

import (
	"fmt"
	"os"
	"runtime/debug"
)

var HadError = false

func Error(line int, message string) {
	Report(line, "", message)
}

func Report(line int, where, message string) {
	_, _ = fmt.Fprintf(os.Stderr, "[line %d] Error%s: %s\n", line, where, message)
	HadError = true
	debug.PrintStack()
	os.Exit(2)
}
