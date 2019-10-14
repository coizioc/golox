package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"

	"golox/loxerror"
	"golox/vm"
)

func runFile(path string) {
	code, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Print(err)
		return
	}
	run(string(code))
}

func runPrompt() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		code, _ := reader.ReadString('\n')
		run(code)
		loxerror.HadError = false
	}
}

func run(source string) {
	vmachine := vm.New()
	vmachine.Interpret(source)
}

func main() {
	if len(os.Args) > 2 {
		fmt.Println("Usage: golox [script]")
		os.Exit(64)
	} else if len(os.Args) == 2 {
		runFile(os.Args[1])
	} else {
		runPrompt()
	}
}
