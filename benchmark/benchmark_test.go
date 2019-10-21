package benchmark

import (
	"golox/vm"
	"testing"
)

func RunTest(t *testing.T, source string, result interface{}) {
	vmachine := vm.New()
	vmachine.Interpret(source)
	if vmachine.Out != result {
		t.Errorf("Incorrect result for source '%s'. Expected: %v. Got: %v.", source, result, vmachine.Out)
	}
}

func TestBenchmark(t *testing.T) {
	tests := []struct {
		source string
		result interface{}
	}{
		{`fun fib(n) { if(n < 2) return n; return fib(n - 1) + fib(n - 2);}
				 print fib(35);`, 9227465.0},
	}

	for _, test := range tests {
		RunTest(t, test.source, test.result)
	}
}
