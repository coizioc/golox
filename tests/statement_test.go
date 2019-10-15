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
		{`var a = "a"; var b = "b"; var c = "c"; a = b = c; print a;`, "c"},
		{`var a = "before"; var c = a = "var"; print a;`, "var"},
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

func TestIf(t *testing.T) {
	tests := []struct {
		source string
		result interface{}
	}{
		{`if(true) print "good";`, "good"},
		{`if(false) print "bad";`, nil},
		{`if(true) { print "good"; }`, "good"},
		{`var a  = false; if(a = true) print a;`, true},
		{`if(false) print "bad"; else print "false";`, "false"},
		{`if(nil) print "bad"; else print "nil";`, "nil"},
		{`if(0) print 0;`, 0.0},
		{`if("") print "empty";`, "empty"},
		{`if(true) print "good";`, "good"},
		{`if(true) print "good"; else print "bad";`, "good"},
		{`if(false) print "bad"; else print "good";`, "good"},
		{`if(false) nil; else { print "block"; }`, "block"},
		{`if(true) if(false) print "bad"; else print "good";`, "good"},
		{`if (false) if (true) print "bad"; else print "bad";`, nil},
	}

	for _, test := range tests {
		RunStatementTest(t, test.source, test.result)
	}
}

func TestFor(t *testing.T) {
	tests := []struct {
		source string
		result interface{}
	}{
		{`var sum = 0; for(var i = 0; i < 11; i = i + 1) sum = sum + i; print sum;`, 55.0},
		{`var sum = 0; for(var i = 0; i < 11; i = i + 1) { sum = sum + i; } print sum;`, 55.0},
		{`var sum = 0; var i = 0; for(; i < 11; i = i + 1) sum = sum + i; print sum;`, 55.0},
		{`var sum = 0; for(var i = 0; i < 11;) { sum = sum + i; i = i + 1; } print sum;`, 55.0},
		{`var i = "before"; for(var i = 0; i < 1; i = i + 1) print i;`, 0.0},
		{`for(var i = 0; i < 1; i = i + 1) {} var i = "after"; print i;`, "after"},
	}

	for _, test := range tests {
		RunStatementTest(t, test.source, test.result)
	}
}

func TestWhile(t *testing.T) {
	tests := []struct {
		source string
		result interface{}
	}{
		{`var sum = 0; var i = 0; while(sum < 55) { i = i + 1; sum = sum + i; } print i;`, 10.0},
		{`while(false) print "bad"; print "good";`, "good"},
	}

	for _, test := range tests {
		RunStatementTest(t, test.source, test.result)
	}
}
