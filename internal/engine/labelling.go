// Copyright © 2015 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

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
