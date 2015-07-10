// Copyright © 2015 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

// Import CAES argument graphs from a JSON serialization of the
// Argument Interchange Format (AIF). For more information about AIF, see:
// http://www.argumentinterchange.org/home
// http://www.arg-tech.org/index.php/projects/

// See also:
// Chesnevar,C., McGinnis,J., Modgil,S. Rahwan,I., Reed,C., Simari,G.,
// South,M., Vreeswijk,G. & Willmott,S. (2006)
// "Towards an Argument Interchange Format",
// Knowledge Engineering Review, 21 (4), pp. 293-316

package aif

import (
	"encoding/json"
	"github.com/carneades/carneades-4/internal/engine/caes"
	"io"
	"io/ioutil"
)

type AIF struct {
	Edges []Edge `json:"edges"`
	Nodes []Node `json:"nodes"`
}

func newAIF() AIF {
	return AIF{
		Edges: []Edge{},
		Nodes: []Node{},
	}
}

type Edge struct {
	Id       string `json:"edgeID"`
	From     string `json:"fromID"`
	FormEdge string `json:"formEdgeID"`
	To       string `json:"toID"`
}

type Node struct {
	Id        string `json:"nodeID"`
	Text      string `json:"text"`
	Scheme    string `json:"scheme"`
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"`
}

func (ag AIF) Caes() *caes.ArgGraph {
	stmts := make(map[string]*caes.Statement)
	args := make(map[string]*caes.Argument)
	issues := make(map[string]*caes.Issue)

	nodeType := func(id string) string {
		if stmts[id] != nil {
			return "Statement"
		} else if args[id] != nil {
			return "Argument"
		} else if issues[id] != nil {
			return "Issue"
		} else {
			return ""
		}
	}

	for _, node := range ag.Nodes {
		switch node.Type {
		case "I":
			s := caes.NewStatement()
			s.Id = node.Id
			s.Text = node.Text
			s.Assumed = true // may be overridden below
			stmts[s.Id] = &s
		case "CA":
			i := caes.NewIssue()
			i.Id = node.Id
			issues[i.Id] = &i
		case "RA":
			arg := caes.NewArgument()
			arg.Id = node.Id
			arg.Scheme = node.Scheme
			args[arg.Id] = &arg
		default:
			continue
		}
	}

	for _, edge := range ag.Edges {
		switch nodeType(edge.From) {
		case "Statement":
			s := stmts[edge.From]
			switch nodeType(edge.To) {
			case "Argument":
				arg := args[edge.To]
				p := caes.Premise{Stmt: s}
				arg.Premises = append(arg.Premises, p)
			case "Issue":
				i := issues[edge.To]
				i.Positions = append(i.Positions, s)
				s.Issue = i
			default:
				continue
			}
		case "Issue":
			i := issues[edge.From]
			switch nodeType(edge.To) {
			case "Statement":
				s := stmts[edge.To]
				i.Positions = append(i.Positions, s)
			default:
				continue
			}
		case "Argument":
			arg := args[edge.From]
			switch nodeType(edge.To) {
			case "Statement":
				s := stmts[edge.To]
				arg.Conclusion = s
				s.Args = append(s.Args, arg)
			default:
				continue
			}
		default:
			continue
		}
	}

	cag := caes.NewArgGraph()

	for i, _ := range stmts {
		s := stmts[i]
		if len(s.Args) > 0 || s.Issue != nil {
			// do not assume statements supported by arguments or at issue
			s.Assumed = false
		}
		cag.Statements = append(cag.Statements, s)
	}

	for _, issue := range issues {
		cag.Issues = append(cag.Issues, issue)
	}

	for _, arg := range args {
		cag.Arguments = append(cag.Arguments, arg)
	}

	return &cag
}

func Import(inFile io.Reader) (*caes.ArgGraph, error) {
	data, err := ioutil.ReadAll(inFile)
	if err != nil {
		return nil, err
	}
	aif := newAIF()
	err = json.Unmarshal(data, &aif)
	if err != nil {
		return nil, err
	}
	ag := aif.Caes()
	return ag, err
}
