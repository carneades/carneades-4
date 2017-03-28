// Copyright © 2015 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

// Import CAES argument graphs from Andreas Peldszus's XML
// format for argument graphs, for use with his Arg-Microtexts corpus. See:
// https://github.com/peldszus/arg-microtexts/blob/master/corpus/arggraph.dtd

// See also:
// Andreas Peldszus and Manfred Stede. From argument diagrams to argumentation
// mining in texts: a survey. International Journal of Cognitive Informatics
// and Natural Intelligence (IJCINI), 7(1):1–31, 2013.

package agxml

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/carneades/carneades-4/src/engine/caes"
)

type Edu struct {
	Id      string `xml:"id,attr"`
	Content string `xml:",chardata"`
}

type Joint struct {
	Id string `xml:"id,attr"`
}

type Adu struct {
	Id   string `xml:"id,attr"`
	Type string `xml:"type,attr"`
}

type Edge struct {
	Id   string `xml:"id,attr"`
	Src  string `xml:"src,attr"`
	Trg  string `xml:"trg,attr"`
	Type string `xml:"type,attr"`
}

type Arggraph struct {
	Id      string  `xml:"id,attr"`
	Stance  string  `xml:"stance,attr"`
	TopicId string  `xml:"topic_id,attr"`
	Edus    []Edu   `xml:"edu"`
	Joints  []Joint `xml:"joint"`
	Adus    []Adu   `xml:"adu"`
	Edges   []Edge  `xml:"edge"`
}

// Convert to an Arggraph to a CAES argument graph
func (pag *Arggraph) Caes() *caes.ArgGraph {
	cag := caes.NewArgGraph()
	cag.Metadata["id"] = pag.Id
	cag.Metadata["stance"] = pag.Stance
	cag.Metadata["topic_id"] = pag.TopicId
	edus := make(map[string][]string) // elementary discourse units
	stmts := make(map[string]*caes.Statement)
	args := make(map[string]*caes.Argument)
	issues := make(map[string]*caes.Issue)
	issueCounter := 0
	assums := map[string]bool{}

	for _, e := range pag.Edus {
		edus[e.Id] = []string{e.Content}
	}

	for _, j := range pag.Joints {
		edus[j.Id] = nil // initialize
	}

	for _, a := range pag.Adus {
		s := caes.NewStatement()
		s.Id = a.Id
		assums[s.Id] = true // overridden below if supported by arguments
		s.Metadata["type"] = a.Type
		stmts[a.Id] = s
	}

	// first pass
	for _, e := range pag.Edges {
		switch e.Type {
		case "seg":
			if edus[e.Trg] != nil {
				// the target node is an edu
				// join them together
				edus[e.Trg] = append(edus[e.Trg], edus[e.Src]...)
			}
		case "sup", "exa", "reb":
			s := stmts[e.Src]
			p := caes.Premise{Stmt: s}
			ps := []caes.Premise{p}
			cid := e.Trg // conclusion id
			if e.Type == "reb" {
				cid = "¬" + e.Trg
				if _, ok := stmts[cid]; !ok {
					// create the contrary statement for rebuttals
					stmts[cid] = &caes.Statement{Id: cid}
					// create the issue
					issueCounter++
					issueId := fmt.Sprintf("i%v", issueCounter)
					s1 := stmts[e.Trg]
					s2 := stmts[cid]
					positions := []*caes.Statement{s1, s2}
					issue := caes.Issue{Id: issueId,
						Standard:  caes.PE,
						Positions: positions}
					issues[issueId] = &issue
					s1.Issue = &issue
					s2.Issue = &issue
				}
			}
			c := stmts[cid]
			arg := caes.Argument{Id: e.Id, Premises: ps, Conclusion: c}
			c.Args = append(c.Args, &arg)
			args[e.Id] = &arg
			if e.Type == "exa" {
				arg.Metadata["note"] = "support by example"
			}
		case "und":
			cid := "¬app(" + e.Trg + ")"
			if _, ok := stmts[cid]; !ok {
				// create the undercutter statement
				stmts[cid] = &caes.Statement{Id: cid}
			}
			c := stmts[cid]
			s := stmts[e.Src]
			p := caes.Premise{Stmt: s}
			ps := []caes.Premise{p}
			arg := caes.Argument{Id: e.Id, Premises: ps, Conclusion: c}
			c.Args = append(c.Args, &arg)
			args[e.Id] = &arg
		default:
			continue
		}
	}

	// second pass
	for _, e := range pag.Edges {
		switch e.Type {
		case "seg":
			if s, ok := stmts[e.Trg]; ok {
				// the target node is an adu, i.e. statement
				s.Text = edus[e.Src][0]
				s.Metadata["edus"] = edus[e.Src]
			}
		case "add":
			s := stmts[e.Src]
			p := caes.Premise{Stmt: s}
			args[e.Trg].Premises = append(args[e.Trg].Premises, p)
		default:
			continue
		}
	}

	for _, s := range stmts {
		// do not assume statements supported by arguments or at issue
		if len(s.Args) > 0 || s.Issue != nil {
			assums[s.Id] = false
		}
		cag.Statements[s.Id] = s
	}
	for _, issue := range issues {
		cag.Issues[issue.Id] = issue
		for _, p := range issue.Positions {
			p.Issue = issue
		}
	}
	for _, arg := range args {
		cag.Arguments[arg.Id] = arg
		arg.Conclusion.Args = append(arg.Conclusion.Args, arg)
		if s, ok := stmts["¬app("+arg.Id+")"]; ok {
			arg.Undercutter = s
		}
	}
	for k, _ := range assums {
		cag.Assumptions = append(cag.Assumptions, k)
	}
	return cag
}

func Import(inFile io.Reader) (*caes.ArgGraph, error) {
	data, err := ioutil.ReadAll(inFile)
	if err != nil {
		return nil, err
	}
	v := Arggraph{}
	err = xml.Unmarshal(data, &v)
	if err != nil {
		return nil, err
	}
	ag := v.Caes()
	return ag, err
}
