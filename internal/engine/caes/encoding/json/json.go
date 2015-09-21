// Copyright © 2015 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

// func Import(in io.Reader) (*caes.ArgGraph, error) - json-Import
// func Export(out io.Writer, ag *caes.ArgGraph) - fast technical json-Export, not for human reading

package json

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/carneades/carneades-4/internal/engine/caes"
	"io"
	"io/ioutil"
	"log"
)

type (
	TempIssue struct {
		id        string
		Meta      map[string]interface{} `json:"meta"`
		Positions []string               `json:"positions"`
		Standard  string                 `json:"standard"` //
	}

	TempStatement struct {
		id      string
		Meta    map[string]interface{} `json:"meta"`
		Text    string                 `json:"text"`    // natural language
		Assumed bool                   `json:"assumed"` // true o false
		//		Issue   string                 `json:"issue"`   // "" if not at issue
		//		args    []string               // concluding with this statement
		Label string `json:"label"` // for storing the evaluated label
	}

	TempArgument struct {
		id          string
		Meta        map[string]interface{} `json:"meta"`
		Scheme      string                 `json:"scheme"`      // name of the scheme
		Premises    []interface{}          `json:"premises"`    // statement or role: statement
		Conclusion  string                 `json:"conclusion"`  // Statement
		Undercutter string                 `json:"undercutter"` // Statement
		Weight      float64                `json:"weight"`      // for storing the evaluated weight
	}

	TempLabels struct {
		In        []string `json:"in"`        // Statements
		Out       []string `json:"out"`       // Statements
		Undecided []string `json:"undecided"` // Statements
	}
	/* TempArgGraphDB - for store the data in a database (i.e. couchdb) */
	TempArgGraphDB struct {
		Meta       map[string]interface{}   `json:"meta"`
		Issues     map[string]TempIssue     `json:"issues"`
		Statements map[string]TempStatement `json:"statements"` // string or TempStatement
		Arguments  map[string]TempArgument  `json:"arguments"`
		References map[string]caes.Metadata `json:"references"`
		//		Assumptions []string                 `json:"assumptions"`
		//		Labels      TempLabels               `json:"labels"`
	}
)

func Export(f io.Writer, ag *caes.ArgGraph) error {
	tmpAG := TempArgGraphDB{Issues: map[string]TempIssue{}, Statements: map[string]TempStatement{}, Arguments: map[string]TempArgument{}}
	// Metadata
	tmpAG.Meta = ag.Metadata
	// References
	tmpAG.References = ag.References
	// Issues
	for _, iss := range ag.Issues {
		tmpIss := TempIssue{Meta: iss.Metadata}
		std := "??"
		switch iss.Standard {
		case caes.DV:
			std = "DV"
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
		tmpStat := TempStatement{Meta: stat.Metadata, Text: stat.Text, Assumed: stat.Assumed}
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
		tmpArg := TempArgument{Meta: arg.Metadata, Weight: arg.Weight, Scheme: arg.Scheme}
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

	}
	d, err := json.Marshal(tmpAG)
	if err != nil {
		log.Fatal("error: %v", err)
		return err
	}
	fmt.Fprintf(f, "%s", string(d))
	return nil
}

func Import(inFile io.Reader) (*caes.ArgGraph, error) {

	data, err := ioutil.ReadAll(inFile)
	if err != nil {
		return nil, err
	}
	// log.Printf("Read-Datei: \nErr: %v len(data): %v \n", err, len(data))

	jsonAG := TempArgGraphDB{}
	err = json.Unmarshal(data, &jsonAG)
	if err != nil {
		return nil, err
	}
	// TempArgGraphDB --> cases.ArgGraph
	caesAG := caes.ArgGraph{Metadata: caes.Metadata{}, Issues: []*caes.Issue{},
		Statements: []*caes.Statement{},
		Arguments:  []*caes.Argument{}, References: map[string]caes.Metadata{}}
	// Metadata
	caesAG.Metadata = jsonAG.Meta
	// References
	caesAG.References = jsonAG.References

	// statements
	caesStatMap := map[string]*caes.Statement{}

	for statId, jsonStat := range jsonAG.Statements {
		caesStat := new(caes.Statement)
		caesStat.Id = statId
		caesStatMap[statId] = caesStat
		caesStat.Metadata = jsonStat.Meta
		caesStat.Text = jsonStat.Text
		caesStat.Assumed = jsonStat.Assumed
		switch jsonStat.Label {
		case "in":
			caesStat.Label = caes.In
		case "out":
			caesStat.Label = caes.Out
		default:
			caesStat.Label = caes.Undecided
		}
		//		if caesAG.Statements == nil { }
		caesAG.Statements = append(caesAG.Statements, caesStat)

	}
	// issues
	for issueId, jsonIssue := range jsonAG.Issues {
		refCaesIssue := new(caes.Issue)
		refCaesIssue.Id = issueId
		refCaesIssue.Metadata = jsonIssue.Meta
		switch jsonIssue.Standard {
		case "DV":
			refCaesIssue.Standard = caes.DV
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
						return &caesAG, errors.New(" *** Semantic Error: Statement: " + refCaes_Stat.Id + ", with two issues: " + jsonIssuePos_Id + ", " + refCaes_Stat.Issue.Id + "\n")
					}
				}
			}
		}
		if caesAG.Issues == nil {
			caesAG.Issues = []*caes.Issue{refCaesIssue}
		} else {
			caesAG.Issues = append(caesAG.Issues, refCaesIssue)
		}
	}

	// arguments

	for jsonArg_Id, jsonArg := range jsonAG.Arguments {
		refCaesArg := new(caes.Argument)
		if caesAG.Arguments == nil {
			caesAG.Arguments = []*caes.Argument{refCaesArg}
		} else {
			caesAG.Arguments = append(caesAG.Arguments, refCaesArg)
		}
		// Argument.Id
		refCaesArg.Id = jsonArg_Id
		// Argument.Metadata
		refCaesArg.Metadata = jsonArg.Meta
		// Argument.Scheme
		refCaesArg.Scheme = jsonArg.Scheme
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

	return &caesAG, nil
}
