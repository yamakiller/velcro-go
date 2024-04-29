package test

import "reflect"

// testingTB is a subset of common methods between *testing.T and *testing.B.
type testingTB interface {
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Helper()
}

// Assert asserts cond is true, otherwise fails the test.
func Assert(t testingTB, cond bool, val ...interface{}) {
	t.Helper()
	if !cond {
		if len(val) > 0 {
			val = append([]interface{}{"assertion failed:"}, val...)
			t.Fatal(val...)
		} else {
			t.Fatal("assertion failed")
		}
	}
}

// Assertf asserts cond is true, otherwise fails the test.
func Assertf(t testingTB, cond bool, format string, val ...interface{}) {
	t.Helper()
	if !cond {
		t.Fatalf(format, val...)
	}
}

// DeepEqual asserts a and b are deep equal, otherwise fails the test.
func DeepEqual(t testingTB, a, b interface{}) {
	t.Helper()
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("assertion failed: %v != %v", a, b)
	}
}

// Panic asserts fn should panic and recover it, otherwise fails the test.
func Panic(t testingTB, fn func()) {
	t.Helper()
	defer func() {
		if err := recover(); err == nil {
			t.Fatal("assertion failed: did not panic")
		}
	}()
	fn()
}

// PanicAt asserts fn should panic and recover it, otherwise fails the test. The expect function can be provided to do further examination of the error.
func PanicAt(t testingTB, fn func(), expect func(err interface{}) bool) {
	t.Helper()
	defer func() {
		if err := recover(); err == nil {
			t.Fatal("assertion failed: did not panic")
		} else {
			if expect != nil && !expect(err) {
				t.Fatal("assertion failed: panic but not expected")
			}
		}
	}()
	fn()
}
