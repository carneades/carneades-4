// Copyright © 2015 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

// func Import(in io.Reader) (*caes.ArgGraph, error) - json-Import
// func Export(out io.Writer, ag *caes.ArgGraph) - fast technical json-Export, not for human reading
// func Json2Caes(jsonAG ArgGraph) (*caes.ArgGraph, error) - transform a json ArgGraph into a caes.ArgGraph
// func Caes2Json( ag *caes.ArgGraph) (ArgGraph, error) - transform a caes.ArgGraph into a json ArgGraph

package json

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/carneades/carneades-4/src/engine/caes"
	"io"
	"io/ioutil"
	"log"
)

type (
	Issue struct {
		id        string
		Meta      map[string]interface{} `json:"meta"`
		Positions []string               `json:"positions"`
		Standard  string                 `json:"standard"` //
	}

	Statement struct {
		id   string
		Meta map[string]interface{} `json:"meta"`
		Text string                 `json:"text"` // natural language
		//      Assumed bool                   `json:"assumed"` // true o false
		//		Issue   string                 `json:"issue"`   // "" if not at issue
		//		args    []string               // concluding with this statement
		Label string `json:"label"` // for storing the evaluated label
	}

	Argument struct {
		id          string
		Meta        map[string]interface{} `json:"meta"`
		Scheme      string                 `json:"scheme"`      // name of the scheme
		Premises    []interface{}          `json:"premises"`    // statement or role: statement
		Conclusion  string                 `json:"conclusion"`  // Statement
		Undercutter string                 `json:"undercutter"` // Statement
		Weight      float64                `json:"weight"`      // for storing the evaluated weight
	}

	Labels struct {
		In        []string `json:"in"`        // Statements
		Out       []string `json:"out"`       // Statements
		Undecided []string `json:"undecided"` // Statements
	}
	/* ArgGraph - for store the data in a database (i.e. couchdb) */
	ArgGraph struct {
		Meta        map[string]interface{}   `json:"meta"`
		Issues      map[string]Issue         `json:"issues"`
		Statements  map[string]Statement     `json:"statements"` // string or Statement
		Arguments   map[string]Argument      `json:"arguments"`
		References  map[string]caes.Metadata `json:"references"`
		Assumptions []string                 `json:"assumptions"` // statement ids
		//		Labels      Labels               `json:"labels"`
	}
)

func NewArgGraph() *ArgGraph {
	return &ArgGraph{
		Meta:        make(map[string]interface{}),
		Issues:      make(map[string]Issue),
		Statements:  make(map[string]Statement),
		Arguments:   make(map[string]Argument),
		References:  make(map[string]caes.Metadata),
		Assumptions: []string{},
	}
}

func Caes2Json(ag *caes.ArgGraph) (ArgGraph, error) {
	tmpAG := NewArgGraph()
	// Metadata
	tmpAG.Meta = ag.Metadata
	// References
	tmpAG.References = ag.References
	// Issues
	for _, iss := range ag.Issues {
		tmpIss := Issue{Meta: iss.Metadata}
		std := "??"
		switch iss.Standard {
		case caes.PE:
			std = "PE"
		case caes.CCE:
			std = "CCE"
		case caes.BRD:
			std = "BRD"
		}

		tmpIss.Standard = std
		first := true

		for _, pos := range iss.Positions {
			if first {
				tmpIss.Positions = []string{pos.Id}
				first = false
			} else {
				tmpIss.Positions = append(tmpIss.Positions, pos.Id)
			}
		}
		tmpAG.Issues[iss.Id] = tmpIss
	}
	// Statements
	for _, stat := range ag.Statements {
		tmpStat := Statement{Meta: stat.Metadata, Text: stat.Text}
		lbl := ""
		switch stat.Label {
		case caes.Undecided:
			lbl = "undecided"
		case caes.In:
			lbl = "in"
		case caes.Out:
			lbl = "out"
		}
		tmpStat.Label = lbl

		tmpAG.Statements[stat.Id] = tmpStat
	}
	//  Arguments
	for _, arg := range ag.Arguments {
		tmpArg := Argument{Meta: arg.Metadata, Weight: arg.Weight, Scheme: arg.Scheme.Id}
		if arg.Undercutter != nil {
			tmpArg.Undercutter = arg.Undercutter.Id
		}
		if arg.Conclusion != nil {
			tmpArg.Conclusion = arg.Conclusion.Id
		}

		first := true
		// Achtung!!! wenn prem.Role != "" dann map[prem.Role] = prem.Stmt.Id
		for _, prem := range arg.Premises {
			if prem.Role == "" {
				if first == true {
					tmpArg.Premises = []interface{}{prem.Stmt.Id}
					first = false
				} else {
					tmpArg.Premises = append(tmpArg.Premises, prem.Stmt.Id)
				}
			} else { // role: statement
				if first == true {
					tmpArg.Premises = []interface{}{map[string]string{prem.Role: prem.Stmt.Id}}
					first = false
				} else {
					tmpArg.Premises = append(tmpArg.Premises, map[string]string{prem.Role: prem.Stmt.Id})
				}
			}
		}

		tmpAG.Arguments[arg.Id] = tmpArg

		// Assumptions
		for k, _ := range ag.Assumptions {
			tmpAG.Assumptions = append(tmpAG.Assumptions, k)
		}

	}
	return *tmpAG, nil
}

func (ag ArgGraph) String() string {
	d, err := json.Marshal(ag)
	if err != nil {
		log.Fatal("error: %v", err)
		return ""
	}
	return string(d)
}

func Export(f io.Writer, ag *caes.ArgGraph) error {
	tmpAG, err := Caes2Json(ag)
	if err != nil {
		log.Fatal("error: %v", err)
		return err
	}
	d, err := json.Marshal(tmpAG)
	if err != nil {
		log.Fatal("error: %v", err)
		return err
	}
	fmt.Fprintf(f, "%s", string(d))
	return nil
}

func Json2Caes(jsonAG ArgGraph) (*caes.ArgGraph, error) {

	// ArgGraph --> cases.ArgGraph
	caesAG := caes.NewArgGraph()
	// Metadata
	caesAG.Metadata = jsonAG.Meta
	// References
	caesAG.References = jsonAG.References
	// Scheme
	schemes := map[string]*caes.Scheme{}

	// statements
	caesStatMap := map[string]*caes.Statement{}

	for statId, jsonStat := range jsonAG.Statements {
		caesStat := new(caes.Statement)
		caesStat.Id = statId
		caesStatMap[statId] = caesStat
		caesStat.Metadata = jsonStat.Meta
		caesStat.Text = jsonStat.Text
		switch jsonStat.Label {
		case "in":
			caesStat.Label = caes.In
		case "out":
			caesStat.Label = caes.Out
		default:
			caesStat.Label = caes.Undecided
		}
		//		if caesAG.Statements == nil { }
		caesAG.Statements[caesStat.Id] = caesStat

	}
	// issues
	for issueId, jsonIssue := range jsonAG.Issues {
		refCaesIssue := new(caes.Issue)
		refCaesIssue.Id = issueId
		refCaesIssue.Metadata = jsonIssue.Meta
		switch jsonIssue.Standard {
		case "PE":
			refCaesIssue.Standard = caes.PE
		case "CCE":
			refCaesIssue.Standard = caes.CCE
		case "BRD":
			refCaesIssue.Standard = caes.BRD
		}
		// refCaesIssue.Positions --> []*caesStat && caesStat.Issue --> *refCaesIssue
		for _, jsonIssuePos_Id := range jsonIssue.Positions {
			for _, refCaes_Stat := range caesAG.Statements {
				if jsonIssuePos_Id == refCaes_Stat.Id {
					if refCaes_Stat.Issue == nil {
						refCaes_Stat.Issue = refCaesIssue
						if refCaesIssue.Positions == nil {
							refCaesIssue.Positions = []*caes.Statement{refCaes_Stat}
						} else {
							refCaesIssue.Positions = append(refCaesIssue.Positions, refCaes_Stat)
						}
					} else {
						return caesAG, errors.New(" *** Semantic Error: Statement: " + refCaes_Stat.Id + ", with two issues: " + jsonIssuePos_Id + ", " + refCaes_Stat.Issue.Id + "\n")
					}
				}
			}
		}
		caesAG.Issues[refCaesIssue.Id] = refCaesIssue
	}

	// arguments

	for jsonArg_Id, jsonArg := range jsonAG.Arguments {
		refCaesArg := new(caes.Argument)
		caesAG.Arguments[refCaesArg.Id] = refCaesArg

		// Argument.Id
		refCaesArg.Id = jsonArg_Id
		// Argument.Metadata
		refCaesArg.Metadata = jsonArg.Meta
		// Argument.Scheme
		if s := schemes[jsonArg.Scheme]; s != nil {
			refCaesArg.Scheme = s
		} else {
			s := caes.Scheme{Id: jsonArg.Scheme, Weight: caes.LinkedWeighingFunction}
			refCaesArg.Scheme = &s
		}
		// Argument.Weight
		// if jsonArg.Weight != 0.0 {
		refCaesArg.Weight = jsonArg.Weight
		// }
		// Argument.Premise
		for _, jsonArg_prem := range jsonArg.Premises {
			jsonArgPremRole := ""
			jsonArgPremStat := ""
			switch jsonArg_prem.(type) {
			case string:
				jsonArgPremStat = jsonArg_prem.(string)
			case map[string]string:
				for role, stat := range jsonArg_prem.(map[string]string) {
					jsonArgPremRole = role
					jsonArgPremStat = stat
				}
			case map[string]interface{}:
				for role, stat := range jsonArg_prem.(map[string]interface{}) {
					jsonArgPremRole = role
					jsonArgPremStat = stat.(string)
				}
				//			default:
				//				fmt.Printf(" *** Premises %T \n", jsonArg_prem)
			}
			//			if jsonArgPremRole != "" {
			//				fmt.Printf(" +++ Role; %s \n", jsonArgPremRole)
			//			}
			// serch Statement
		PrimeStatLoop:
			for _, refCaes_Stat := range caesAG.Statements {
				if refCaes_Stat.Id == jsonArgPremStat {
					if refCaesArg.Premises == nil {
						refCaesArg.Premises = []caes.Premise{caes.Premise{Stmt: refCaes_Stat, Role: jsonArgPremRole}}
					} else {
						refCaesArg.Premises = append(refCaesArg.Premises, caes.Premise{Stmt: refCaes_Stat, Role: jsonArgPremRole})
					}
					break PrimeStatLoop
				}
			}
		}
		// Argument.Conclusion
		// Reference: Argument.Concliusion --> *Statement && Statement.Args --> []*Argument

		if jsonArg.Conclusion != "" {
		ConclusionLoop:
			for _, refCaes_Stat := range caesAG.Statements {
				if refCaes_Stat.Id == jsonArg.Conclusion {
					if refCaes_Stat.Args == nil {
						refCaes_Stat.Args = []*caes.Argument{refCaesArg}
					} else {
						refCaes_Stat.Args = append(refCaes_Stat.Args, refCaesArg)
					}
					refCaesArg.Conclusion = refCaes_Stat
					break ConclusionLoop
				}
			}
		}
		// Argument.Undercutter
		// Reference: Argument.Undercutter --> *Statement &&
		// No undercutter in  Statement.Args --> []*Argument

		if jsonArg.Undercutter != "" {
		UndercutterLoop:
			for _, refCaes_Stat := range caesAG.Statements {
				if refCaes_Stat.Id == jsonArg.Undercutter {
					// if refCaes_Stat.Args == nil {
					//  	refCaes_Stat.Args = []*caes.Argument{refCaesArg}
					// } else {
					//  	refCaes_Stat.Args = append(refCaes_Stat.Args, refCaesArg)
					// }
					refCaesArg.Undercutter = refCaes_Stat
					break UndercutterLoop
				}
			}
		}
	}

	// Assumptions
	for _, s := range jsonAG.Assumptions {
		caesAG.Assumptions[s] = true
	}

	return caesAG, nil

}

func Import(inFile io.Reader) (*caes.ArgGraph, error) {
	data, err := ioutil.ReadAll(inFile)
	if err != nil {
		return nil, err
	}
	// log.Printf("Read-Datei: \nErr: %v len(data): %v \n", err, len(data))

	jsonAG := ArgGraph{}
	err = json.Unmarshal(data, &jsonAG)
	if err != nil {
		return nil, err
	}

	return Json2Caes(jsonAG)
}
