package tests

import (
	"golox/vm"
	"testing"
)

func RunStatementTest(t *testing.T, source string, result interface{}) {
	vmachine := vm.New()
	vmachine.Interpret(source)
	if vmachine.Out != result {
		t.Errorf("Incorrect result for source '%s'. Expected: %v. Got: %v.", source, result, vmachine.Out)
	}
}

func TestVarDecl(t *testing.T) {
	tests := []struct {
		source string
		result interface{}
	}{
		{`var x = "hello"; print x;`, "hello"},
		{`var x = "hello"; x = "world"; print x;`, "world"},
		{`var x = "hello"; var y = "world"; x = y; print x;`, "world"},
		{`var x = "hello"; var x; print x;`, nil},
		{`var x = "hello"; var x = x; print x;`, "hello"},
		{`var x; print x;`, nil},
	}

	for _, test := range tests {
		RunStatementTest(t, test.source, test.result)
	}
}

func TestLocalVars(t *testing.T) {
	tests := []struct {
		source string
		result interface{}
	}{
		{`{var a = "a"; var b = a + " b"; print b;}`, "a b"},
		{`{var a = "a"; var b = a + " b"; var c = a + " c"; print c;}`, "a c"},
		{`{var a = "outer";{print a;}}`, "outer"},
		{`{var a = "first";}{var a = "second"; print a;}`, "second"},
		{`{var a = "outer";{var a = "inner"; print a;}}`, "inner"},
		{`{var a = "first";}{var a = "second"; print a;}`, "second"},
	}

	for _, test := range tests {
		RunStatementTest(t, test.source, test.result)
	}
}
