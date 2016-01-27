package terms_test

import (
	"fmt"
	"github.com/carneades/carneades-4/src/engine/terms"
	"testing"
)

func Test_terms(t *testing.T) {
	// src is the input that we want to tokenize.
	// src := []byte("cos(x) + 1i*sin(x) // Euler")
	// ReadString("[function/2, p(x,Y) ==> Y < 23.4 | r(x),( gcd(N) \\ gcd(M) <=> N<=M| L is M mod N, gcd(L)), Z, ]")
	// ReadString("prime(N) ==> N>2 | M is N-1, prime(m). prime(A) \\ prime(B) <=> B mod A =:= 0 | true")

	// tt("E")
	tt("23")
	tt("\"str\"")
	tt("3.147")
	tt("abc")
	tt("foo(a, b, c)")
	tt("baz(a+3, b*7, bar(x,Y))")
}

func tt(str string) {
	fmt.Printf("----> %s \n", str)
	term, ok := terms.ReadString(str)

	fmt.Printf("<---- %s = OK: %v Term: %s \n", str, ok, term.String())
}
