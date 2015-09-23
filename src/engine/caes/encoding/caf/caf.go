// Copyright © 2015 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

// Import CAES argument graphs from the Carneades Argument Format
// (CAF) XML schema. The Carneades Argument Format (CAF) is the native format
// of Carneades 3, developed in the European IMACT project.
// For further information, see:
// https://github.com/carneades/carneades-3/blob/master/schemas/CAF.rnc
// http://www.policy-impact.eu/

package caf

import (
	"encoding/xml"
	// "errors"
	"fmt"
	"github.com/carneades/carneades-4/src/engine/caes"
	"io"
	"io/ioutil"
	"regexp"
)

type CAF struct {
	XMLName    xml.Name   `xml:"caf"`
	Version    string     `xml:"version,attr"`
	Metadata   Metadata   `xml:"metadata"`
	Statements Statements `xml:"statements"`
	Arguments  Arguments  `xml:"arguments"`
	References References `xml:"references"`
}

type Metadata struct {
	XMLName      xml.Name     `xml:"metadata"`
	Key          string       `xml:"key,attr"`
	Contributor  string       `xml:"contributor,attr"`
	Coverage     string       `xml:"coverage,attr"`
	Creator      string       `xml:"creator,attr"`
	Date         string       `xml:"date,attr"`
	Format       string       `xml:"format,attr"`
	Identifier   string       `xml:"identifier,attr"`
	Language     string       `xml:"language,attr"`
	Publisher    string       `xml:"publisher,attr"`
	Relation     string       `xml:"relation,attr"`
	Rights       string       `xml:"rights,attr"`
	Source       string       `xml:"source,attr"`
	Subject      string       `xml:"subject,attr"`
	Title        string       `xml:"title,attr"`
	Type         string       `xml:"type,attr"`
	Descriptions Descriptions `xml:"descriptions"`
}

func (md *Metadata) toMap() map[string]interface{} {
	m := make(map[string]interface{})
	if md.Key != "" {
		m["key"] = md.Key
	}
	if md.Contributor != "" {
		m["contributor"] = md.Contributor
	}
	if md.Coverage != "" {
		m["coverage"] = md.Coverage
	}
	if md.Creator != "" {
		m["creator"] = md.Creator
	}
	if md.Date != "" {
		m["date"] = md.Date
	}
	if md.Format != "" {
		m["format"] = md.Format
	}
	if md.Identifier != "" {
		m["identifier"] = md.Identifier
	}
	if md.Language != "" {
		m["language"] = md.Language
	}
	if md.Publisher != "" {
		m["publisher"] = md.Publisher
	}
	if md.Relation != "" {
		m["relation"] = md.Relation
	}
	if md.Rights != "" {
		m["rights"] = md.Rights
	}
	if md.Source != "" {
		m["source"] = md.Source
	}
	if md.Subject != "" {
		m["subject"] = md.Subject
	}
	if md.Title != "" {
		m["title"] = md.Title
	}
	if md.Type != "" {
		m["type"] = md.Type
	}
	if len(md.Descriptions.Content) > 0 {
		m["descriptions"] = make(map[string]string)
		for _, desc := range md.Descriptions.Content {
			ds := m["descriptions"].(map[string]string)
			ds[desc.Lang] = desc.Text
		}
	}
	return m
}

type Description struct {
	XMLName xml.Name `xml:"description"`
	Lang    string   `xml:"lang,attr"`
	Text    string   `xml:",innerxml"`
}

type Descriptions struct {
	XMLName xml.Name      `xml:"descriptions"`
	Content []Description `xml:"description"`
}

type Statements struct {
	XMLName xml.Name    `xml:"statements"`
	Content []Statement `xml:"statement"`
}

type Statement struct {
	XMLName  xml.Name     `xml:"statement"`
	Id       string       `xml:"id,attr"`
	Weight   float64      `xml:"weight,attr"`
	Value    float64      `xml:"value,attr"`
	Standard string       `xml:"standard,attr"`
	Atom     string       `xml:"atom,attr"`
	Main     bool         `xml:"main,attr"`
	Metadata Metadata     `xml:"metadata"`
	Texts    Descriptions `xml:"descriptions"`
}

type Arguments struct {
	XMLName xml.Name   `xml:"arguments"`
	Content []Argument `xml:"argument"`
}

type Argument struct {
	XMLName    xml.Name   `xml:"argument"`
	Id         string     `xml:"id,attr"`
	Strict     bool       `xml:"strict,attr"`
	Pro        bool       `xml:"pro,attr"`
	Scheme     string     `xml:"scheme,attr"`
	Weight     float64    `xml:"weight,attr"`
	Value      float64    `xml:"value,attr"`
	Metadata   Metadata   `xml:"metadata"`
	Conclusion Conclusion `xml:"conclusion"`
	Premises   Premises   `xml:"premises"`
}

type Premises struct {
	XMLName xml.Name  `xml:"premises"`
	Content []Premise `xml:"premise"`
}

type Premise struct {
	XMLName   xml.Name `xml:"premise"`
	Positive  bool     `xml:"positive,attr"`
	Role      string   `xml:"role,attr"`
	Implicit  string   `xml:"implicit,attr"`
	Statement string   `xml:"statement,attr"` // Id
}

type Conclusion struct {
	XMLName   xml.Name `xml:"conclusion"`
	Statement string   `xml:"statement,attr"` // Id
}

type References struct {
	XMLName xml.Name   `xml:"references"`
	Content []Metadata `xml:"metadata"`
}

var undercutRegExp = regexp.MustCompile(`^\((undercut|valid)[[:blank:]]+([^\)]*)\)$`)

// Convert to an CAF argument graph to a CAES argument graph
func (caf *CAF) Caes() *caes.ArgGraph {
	cag := caes.NewArgGraph()
	cag.Metadata = caf.Metadata.toMap()
	for _, md := range caf.References.Content {
		cag.References[md.Key] = md.toMap()
	}

	stmts := make(map[string]*caes.Statement)
	standards := make(map[string]caes.Standard) // proof standards
	args := make(map[string]*caes.Argument)
	uids := make(map[string]string) // from URIs to shorter ids
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

	// isUndercutter: checks whether the atom of the statement with the given id
	// has the form (undercut X) or (valid X). Returns the predicate of the
	// undercutter, "valid" or "undercut", in the first result parameter and
	// the UID of the argument undercut in the second result parameter.
	// The boolean result parameter is true only if the atom of the
	// statement is an undercutter.
	isUndercutter := func(id string) (string, string, bool) {
		if stmt, ok := stmts[id]; ok {
			if atom, ok := stmt.Metadata["atom"]; ok {
				v := undercutRegExp.FindStringSubmatch(atom.(string))
				if len(v) != 3 {
					return "", "", false
				} else {
					return v[1], v[2], true
				}
			}
		}
		return "", "", false
	}

	sid := 0 // counter for generating short statement ids
	for _, s := range caf.Statements.Content {
		sid++
		stmt := caes.NewStatement()
		stmt.Id = fmt.Sprintf("s%v", sid)
		uids[s.Id] = stmt.Id
		stmts[stmt.Id] = stmt
		stmt.Metadata = s.Metadata.toMap()
		if s.Atom != "" {
			stmt.Metadata["atom"] = s.Atom
		}
		if len(s.Texts.Content) > 0 {
			stmt.Text = s.Texts.Content[0].Text
			stmt.Metadata["texts"] = make(map[string]string)
			for _, desc := range s.Texts.Content {
				txts := stmt.Metadata["texts"].(map[string]string)
				txts[desc.Lang] = desc.Text
			}
		} else if s.Atom != "" {
			stmt.Text = s.Atom
		} else {
			stmt.Text = stmt.Id
		}
		switch s.Standard {
		case "CCE":
			standards[stmt.Id] = caes.CCE
		case "BRD":
			standards[stmt.Id] = caes.BRD
		default:
			standards[stmt.Id] = caes.PE
		}

		if s.Weight > 0.5 {
			stmt.Assumed = true
		} else if s.Weight < 0.5 {
			c := complement(stmt.Id)
			stmts[c].Assumed = true
		}

		if s.Value > 0.5 {
			stmt.Label = caes.In
			if hasComplement(stmt.Id) {
				stmts[complement(stmt.Id)].Label = caes.Out
			}
		} else if s.Value < 0.5 {
			stmt.Label = caes.Out
			if hasComplement(stmt.Id) {
				stmts[complement(stmt.Id)].Label = caes.In
			}
		}

	}

	aid := 0 // counter for generating short statement ids
	for _, a := range caf.Arguments.Content {
		aid++
		var arg *caes.Argument
		if uids[a.Id] == "" {
			arg = caes.NewArgument()
			arg.Id = fmt.Sprintf("a%v", aid)
			args[arg.Id] = arg
		} else {
			arg = args[uids[a.Id]]
		}
		uids[a.Id] = arg.Id
		arg.Metadata = a.Metadata.toMap()
		arg.Scheme = a.Scheme
		arg.Weight = a.Weight
		// a.Value ignored, since arguments are no longer labelled

		if a.Pro {
			arg.Conclusion = stmts[uids[a.Conclusion.Statement]]
		} else {
			arg.Conclusion = stmts[complement(uids[a.Conclusion.Statement])]
		}
		arg.Conclusion.Args = append(arg.Conclusion.Args, arg)

		// handle undercutters for both Carneades 3.5 (not valid)
		// and Carneades 3.7 (undercut) versions of CAF
		pred, uid, ok := isUndercutter(arg.Conclusion.Id)
		if ok && ((pred == "undercut" && a.Pro) ||
			(pred == "valid" && !a.Pro)) {
			// the conclusion undercuts some argument
			var arg2 *caes.Argument
			if uids[uid] == "" {
				arg2 := caes.NewArgument()
				aid++
				arg2.Id = fmt.Sprintf("a%v", aid)
				args[arg2.Id] = arg2
				uids[uid] = arg2.Id
			} else {
				arg2 = args[uids[uid]]
			}
			arg2.Undercutter = arg.Conclusion
			// replace the uid in the atoms and texts of the undercutter
			// and its complement with uids[uid]
			arg.Conclusion.Metadata["atom"] = "¬valid(" + arg2.Id + ")"
			arg.Conclusion.Text = "¬valid(" + arg2.Id + ")"
			notS := stmts[complement(arg.Conclusion.Id)]
			notS.Metadata["atom"] = "valid(" + arg2.Id + ")"
			notS.Text = "valid(" + arg2.Id + ")"
		}

		i := 0 // premise index
		for _, p := range a.Premises.Content {
			var s *caes.Statement
			if p.Positive {
				s = stmts[uids[p.Statement]]
			} else {
				s = stmts[complement(uids[p.Statement])]
			}
			pr := caes.Premise{Stmt: s, Role: p.Role}
			// ignore p.Implicit
			arg.Premises = append(arg.Premises, pr)
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
	caf := CAF{}
	err = xml.Unmarshal(data, &caf)
	if err != nil {
		return nil, err
	}
	return caf.Caes(), nil
}
