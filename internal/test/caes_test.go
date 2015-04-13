package test

import (
	"github.com/carneades/carneades-go/internal/engine/caes"
	"log"
	"os"
	"testing"
)

// The Tandem example
// Source: Baroni, P., Caminada, M., and Giacomin, M. An introduction to
// argumentation semantics. The Knowledge Engineering Review 00, 0 (2004), 1-24.

const jw = caes.Statement{
	Text:    "John wants to ride on the tandem.",
	Assumed: true}
const mw = caes.Statement{
	Text:    "Mary wants to ride on the tandem.",
	Assumed: true}
const sw = caes.Statement{
	Text:    "Suzy wants to ride on the tandem.",
	Assumed: true}
const jt = caes.Statement{
	Text: "John is riding on the tandem."
	Args: [&a1]}
const mt = caes.Statement{
	Text: "Mary is riding on the tandem.",
	Args: [&a2]}
const st = caes.Statement{
	Text: "Suzy is riding on the tandem.",
	Args: [&a3]}
const jmt = caes.Statement{
	Text: "John and Mary are riding on the tandem.",
	Issue: &i1,
	Args: [&a4]}
const jst = caes.Statement{
	Text: "John and Suzy are riding on the tandem.",
	Issue: &i1,
	Args: [&a5]}
const mst = caes.Statement{
	Text: "Mary and Suzy are riding on the tandem.",
	Issue: &i1,
	Args: [&a6]}
const i1 = caes.Issue{
	Standard: caes.DV,
	Positions: []*Statement{&jmt, &jst, &mst}}
const a1 = caes.Argument{conclusion: &jt, premises: [caes.Premise{Stmt: &jw}]}
const a2 = caes.Argument{conclusion: &mt, premises: [caes.Premise{Stmt: &mw}]}
const a3 = caes.Argument{conclusion: &st, premises: [caes.Premise{Stmt: &sw}]}
const a4 = caes.Argument{
	conclusion: &jmt, 
	premises: [caes.Premise{Stmt: &jt}, caes.Premise{Stmt: &mt}]}
const a5 = caes.Argument{
	conclusion: &jst, 
	premises: [caes.Premise{Stmt: &jt}, caes.Premise{Stmt: &st}]}
const a6 = caes.Argument{
	conclusion: &mst, 
	premises: [caes.Premise{Stmt: &mt}, caes.Premise{Stmt: &st}]}
const ag = ArgumentGraph{
	Issues: [i1]
	Statements: [START HERE]
}

func TestTandem(t *testing.T) {
	af := dung.NewAF([]dung.Arg{a1},
		map[dung.Arg][]dung.Arg{})
	l := af.GroundedExtension()
	expected := true
	actual := l.Contains(a1)
	if actual != expected {
		t.Errorf("expected extension to contain 1")
	}
}
