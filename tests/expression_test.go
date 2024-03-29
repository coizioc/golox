package tests

import (
	"golox/vm"
	"testing"
)

func RunExpressionTest(t *testing.T, source string, result interface{}) {
	vmachine := vm.New()
	source = "print " + source + ";"
	vmachine.Interpret(source)
	if vmachine.Out != result {
		t.Errorf("Incorrect result for source '%s'. Expected: %v. Got: %v.", source, result, vmachine.Out)
	}
}

func TestUnaryOp(t *testing.T) {
	tests := []struct {
		source string
		result interface{}
	}{
		{"-3", -3.0},
		{"--3", 3.0},
		{"-(3 + 5)", -8.0},
		{"!true", false},
		{"!!false", false},
		{"!nil", true},
	}

	for _, test := range tests {
		RunExpressionTest(t, test.source, test.result)
	}
}

func TestBinaryOp(t *testing.T) {
	tests := []struct {
		source string
		result interface{}
	}{
		{"3 + 4", 7.0},
		{"3 - 4", -1.0},
		{"3 * 4", 12.0},
		{"3 / 4", 0.75},
		{"3 + 6 / 2", 6.0},
		{"3 * (3 + 1)", 12.0},
		{"(5 - (3 - 1)) + -1", 2.0},
		{"3 == 4", false},
		{"3 != 4", true},
		{"3 > 4", false},
		{"3 < 4", true},
		{"3 >= 3", true},
		{"4 >= 3", true},
		{"3 <= 3", true},
		{"4 <= 3", false},
		{`"Hello" + " World!"`, "Hello World!"},
	}

	for _, test := range tests {
		RunExpressionTest(t, test.source, test.result)
	}
}
