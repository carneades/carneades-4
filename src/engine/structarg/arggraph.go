// Argument Graphs
package engine

// Proof Standards
type Standard int

const (
	PE  Standard = iota // preponderance of the evidence
	DV                  // dialectical validity
	CCE                 // clear and convincing evidence
	BRD                 // beyond reasonable doubt
)

type Statement struct {
	Id       string
	Atom     string // formalization
	Weight   float64
	Value    float64
	Standard Standard
	Main     bool // a main issue
	Metadata Metadata
	Text     map[Lang]string
}

type Literal struct {
	Positive bool
	StmtId   string
}

func PosLit(stmtId string) Literal {
	return Literal{true, stmtId}
}

func NegLit(stmtId string) Literal {
	return Literal{false, stmtId}
}

type ArgumentKind int

const (
	Strict ArgumentKind = iota
	Defeasible
	Cumulative
)

type Premise struct {
	Literal  Literal
	Role     string
	Implicit bool // implicit in some source document
}

type Argument struct {
	Id         string
	Kind       ArgumentKind
	Scheme     string
	Weight     float64
	Value      float64
	Premises   []Premise
	Conclusion Literal
}

type ArgGraph struct {
	Metadata   Metadata
	Statements map[string]Statement // key list
	Arguments  map[string]Argument
	References map[string]Metadata // indexed by key
}
