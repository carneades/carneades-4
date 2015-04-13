package engine

type Label int

const (
	Out Label = iota
	In
	Undecided
)

func (l Label) String() string {
	switch l {
	case In:
		return "in"
	case Out:
		return "out"
	default:
		return "undecided"
	}
}

type Labelling map[Arg]Label

func NewLabelling() Labelling {
	return Labelling(make(map[Arg]Label))
}

func (l Labelling) Get(arg Arg) Label {
	v, found := l[arg]
	if found {
		return v
	} else {
		return Undecided
	}
}
