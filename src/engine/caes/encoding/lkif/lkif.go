// Copyright © 2015 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

// Import CAES argument graphs from the Legal Knowledge Interchange Format
// (LKIF) XML schema. For more information about LKIF, see:
// Gordon, T. F. The Legal Knowledge Interchange Format (LKIF).
// Deliverable 4.1, European Commission, 2008.
// http://www.tfgordon.de/publications/files/GordonLKIF2008.pdf

// LKIF is the native format of Carneades 2, a desktop argument mapping
// tool with a graphical user interface. Also known as the Carneades Editor.
// https://github.com/carneades/carneades-2/blob/master/schemas/LKIF.rnc

package lkif

import (
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/carneades/carneades-4/src/engine/caes"
	"io"
	"io/ioutil"
)

type LKIF struct {
	XMLName        xml.Name       `xml:"lkif"`
	Version        string         `xml:"version,attr"`
	ArgumentGraphs ArgumentGraphs `xml:"argument-graphs"`
}

type ArgumentGraphs struct {
	XMLName xml.Name        `xml:"argument-graphs"`
	Content []ArgumentGraph `xml:"argument-graph"`
}

type ArgumentGraph struct {
	XMLName    xml.Name   `xml:"argument-graph"`
	Id         string     `xml:"id,attr"`
	Title      string     `xml:"title,attr"`
	Main       string     `xml:"main-issue,attr"`
	Statements Statements `xml:"statements"`
	Arguments  Arguments  `xml:"arguments"`
}

type Statements struct {
	Content []Statement `xml:"statement"`
}

type Arguments struct {
	Content []Argument `xml:"argument"`
}

type Statement struct {
	XMLName    xml.Name `xml:"statement"`
	Id         string   `xml:"id,attr"`
	Value      string   `xml:"value,attr"`
	Assumption bool     `xml:"assumption,attr"`
	Standard   string   `xml:"standard,attr"`
	Atom       Atom     `xml:"s"`
}

type Atom struct {
	XMLName   xml.Name `xml:"s"`
	Predicate string   `xml:"pred,attr"`
	Assumable bool     `xml:"assumable,attr"`
	Text      string   `xml:",innerxml"`
}

type Argument struct {
	XMLName    xml.Name   `xml:"argument"`
	Id         string     `xml:"id,attr"`
	Title      string     `xml:"title,attr"`
	Direction  string     `xml:"direction,attr"`
	Scheme     string     `xml:"scheme,attr"`
	Weight     float64    `xml:"weight,attr"`
	Conclusion Conclusion `xml:"conclusion"`
	Premises   Premises   `xml:"premises"`
}

type Conclusion struct {
	XMLName   xml.Name `xml:"conclusion"`
	Statement string   `xml:"statement,attr"` // Id
}

type Premises struct {
	Content []Premise `xml:"premise"`
}

type Premise struct {
	XMLName   xml.Name `xml:"premise"`
	Polarity  string   `xml:"polarity,attr"`
	Type      string   `xml:"type,attr"`
	Role      string   `xml:"role,attr"`
	Statement string   `xml:"statement,attr"` // Id
}

// Convert to an LKIF argument graph to a CAES argument graph
func (lag *ArgumentGraph) Caes() *caes.ArgGraph {
	cag := caes.NewArgGraph()
	if lag.Id != "" {
		cag.Metadata["id"] = lag.Id
	}
	if lag.Title != "" {
		cag.Metadata["title"] = lag.Title
	}
	if lag.Main != "" {
		cag.Metadata["main-issue"] = lag.Main
	}
	stmts := make(map[string]*caes.Statement)
	standards := make(map[string]caes.Standard) // proof standards
	args := make(map[string]*caes.Argument)
	issues := make(map[string]*caes.Issue)
	issueCounter := 0

	// hasComplement: returns true if the statement with the given id
	// has a complement in the argument graph
	hasComplement := func(stmtId string) bool {
		negId := "¬" + stmtId
		if _, ok := stmts[negId]; ok {
			return true
		} else {
			return false
		}
	}

	// complement: returns the id of the complement of statement
	// with the given id, creating the statement for the complement
	// if it does not already exist
	complement := func(stmtId string) string {
		negId := "¬" + stmtId
		if _, ok := stmts[negId]; ok {
			return negId
		} else {
			if s, ok := stmts[stmtId]; ok {
				neg := caes.NewStatement()
				neg.Id = "¬" + s.Id
				neg.Text = "¬" + s.Text
				stmts[neg.Id] = neg
				return neg.Id
			} else {
				return "" // shouldn't happen
			}
		}
	}

	for _, s := range lag.Statements.Content {
		stmt := caes.NewStatement()
		stmt.Id = s.Id
		stmt.Text = s.Atom.Text
		switch s.Standard {
		case "CCE":
			standards[stmt.Id] = caes.CCE
		case "BRD":
			standards[stmt.Id] = caes.BRD
		default:
			standards[stmt.Id] = caes.PE
		}

		stmts[stmt.Id] = stmt
		// s.Assumption ignored as CAES does not distinguish
		// between facts and assumptions
		switch s.Value {
		case "true":
			stmt.Assumed = true
		case "false":
			c := complement(stmt.Id)
			stmts[c].Assumed = true
		default:
			continue
		}
	}

	for _, a := range lag.Arguments.Content {
		arg := caes.NewArgument()
		if a.Id != "" {
			arg.Id = a.Id
		}
		if a.Title != "" {
			arg.Metadata["title"] = a.Title
		}
		if a.Scheme != "" {
			arg.Scheme = a.Scheme
		}
		if a.Weight != 0.0 {
			arg.Weight = a.Weight
		}
		args[arg.Id] = arg
		switch a.Direction {
		case "con":
			arg.Conclusion = stmts[complement(a.Conclusion.Statement)]
		default:
			arg.Conclusion = stmts[a.Conclusion.Statement]
		}
		arg.Conclusion.Args = append(arg.Conclusion.Args, arg)
		i := 0 // premise index
		for _, p := range a.Premises.Content {
			var s *caes.Statement
			if p.Polarity == "positive" {
				s = stmts[p.Statement]
			} else {
				s = stmts[complement(p.Statement)]
			}
			pr := caes.Premise{Stmt: s, Role: p.Role}
			switch p.Type {
			case "ordinary":
				arg.Premises = append(arg.Premises, pr)
			case "exception":
				var us *caes.Statement        // undercutter statement
				uid := "¬app(" + arg.Id + ")" // undercutter id
				if s, ok := stmts[uid]; ok {
					us = s
				} else {
					s2 := caes.NewStatement()
					us = s2
					us.Id = uid
					us.Text = uid
					stmts[us.Id] = us
					arg.Undercutter = us
				}
				e := caes.NewArgument() // for the exception
				e.Id = arg.Id + "." + fmt.Sprintf("%v", i)
				e.Conclusion = us
				e.Premises = []caes.Premise{pr}
				args[e.Id] = e
			case "assumption":
				arg.Premises = append(arg.Premises, pr)
				pr.Stmt.Assumed = true
			default:
				continue
			}
			i++
		}
	}
	for sid, s := range stmts {
		cag.Statements = append(cag.Statements, s)
		if hasComplement(sid) {
			// create issues for statements with complements
			issueCounter++
			issue := caes.NewIssue()
			issue.Id = fmt.Sprintf("i%v", issueCounter)
			issue.Standard = standards[sid]
			issue.Positions = []*caes.Statement{s, stmts[complement(sid)]}
			for _, p := range issue.Positions {
				p.Issue = issue
			}
			issues[issue.Id] = issue
		}
	}
	for _, issue := range issues {
		cag.Issues = append(cag.Issues, issue)

	}
	for _, arg := range args {
		cag.Arguments = append(cag.Arguments, arg)
	}
	return cag
}

func Import(inFile io.Reader) (*caes.ArgGraph, error) {
	data, err := ioutil.ReadAll(inFile)
	if err != nil {
		return nil, err
	}
	// lkif := LKIF{ArgumentGraphs: ArgumentGraphs{Content: []ArgumentGraph{}}}
	lkif := LKIF{}
	err = xml.Unmarshal(data, &lkif)
	if err != nil {
		return nil, err
	}
	if len(lkif.ArgumentGraphs.Content) == 0 {
		return nil, errors.New("No argument graphs found in the input file.\n")
	}
	ag := lkif.ArgumentGraphs.Content[0].Caes()
	return ag, err
}
