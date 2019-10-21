package tests

import (
	"golox/vm"
	"testing"
)

func RunFunctionTest(t *testing.T, source string, result interface{}) {
	vmachine := vm.New()
	vmachine.Interpret(source)
	if vmachine.Out != result {
		t.Errorf("Incorrect result for source '%s'. Expected: %v. Got: %v.", source, result, vmachine.Out)
	}
}

// TODO fix output for <fn X>. Tests otherwise return correct output.
func TestFunction(t *testing.T) {
	tests := []struct {
		source string
		result interface{}
	}{
		{`fun foo() {} print foo;`, "<fn foo>"},
		{`print clock;`, "<native fn>"},
		{`fun hello() { print "Hello"; } hello();`, "Hello"},
		{`fun fib(n) { if(n < 2) return n; return fib(n - 1) + fib(n - 2);}
				 print fib(8);`, 21.0},
		{`fun isEven(n) { if (n == 0) return true; return isOdd(n - 1); }
				 fun isOdd(n) { return isEven(n - 1); } print isEven(4);`, true},
		{`fun isEven(n) { if (n == 0) return true; return isOdd(n - 1); }
				 fun isOdd(n) { return isEven(n - 1); } print isOdd(3);`, true},
		{`fun add(a, b) { return a + b; } fun mul(a, b) { return a * b; }
			     print add(mul(2, 4), 5);`, 13.0},
	}

	for _, test := range tests {
		RunFunctionTest(t, test.source, test.result)
	}
}
