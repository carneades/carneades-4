package terms_test

import (
	"fmt"
	"github.com/carneades/carneades-4/src/engine/terms"
	"testing"
)

func Test_terms(t *testing.T) {
	// src is the input that we want to tokenize.
	// src := []byte(t, "cos(x) + 1i*sin(x) // Euler")
	// ReadString(t, "[function/2, p(x,Y) ==> Y < 23.4 | r(x),( gcd(N) \\ gcd(M) <=> N<=M| L is M mod N, gcd(L)), Z, ]")
	// ReadString(t, "prime(N) ==> N>2 | M is N-1, prime(m). prime(A) \\ prime(B) <=> B mod A =:= 0 | true")

	tt(t, "E")
	tt(t, "23")
	tt(t, "\"str\"")
	tt(t, "3.147")
	tt(t, "abc")
	tt(t, "'Joe Smith'")

	tt(t, "foo(a, b, c)")
	tt(t, "baz(a+3, b*7, bar(x,Y))")
	tt(t, "[foo(), baz(2*3+4*5,VAR,atom), X]")

	tt(t, "2*3*4+5*6*7")
	tt(t, "2+3+4*5+6+7")
	tt(t, "2+(3+4)*(5+6)+7")
	tt(t, "A/B/C")
	tt(t, " a==b && a<b && a>b || a=<b || a!= b && a <= b || a >= b && a in b")
	tt(t, "a+b-c or d^e")
	tt(t, "a*b/c%d<<e>>f&g&^h")
	tt(t, "[a | B]")
	tt(t, "[a,b,c|D]")
	tt(t, "'foo'('Joe Smith')")
	tt(t, "--A")
	tt(t, "+A")
	tt(t, "-a+-b+^c+!d")
	tt(t, "---A+!!!B++++C")
	// tt(t, "_t(-_a,_B)")

	// Fehler
	/*tt(t, "Fehler")
	tt(t, "[a, b,]")
	tt(t, "[a|b]")
	*/
}

func tt(t *testing.T, str string) {
	// fmt.Printf("----> %s \n", str)
	term, ok := terms.ReadString(str)
	fmt.Printf("===RUN Scan %s \n               Term: %s \n", str, term.String())
	if !ok {
		t.Errorf(fmt.Sprintf("Scan \"%s\" failed, term: %s", str, term.String()))
	}
}
