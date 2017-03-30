// Copyright © 2015 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

// func Import(in io.Reader) (*caes.ArgGraph, error)
// func Export(out io.Writer, caesAg *caes.ArgGraph)
// func ExportWithReferences(out io.Writer, caesAg *caes.ArgGraph)

package yaml

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/carneades/carneades-4/src/engine/caes"
	"github.com/carneades/carneades-4/src/engine/terms"
	"gopkg.in/yaml.v2"
	// "log"
	"strconv"
	"strings"
)

type (
	argMapGraph struct {
		Arguments             map[string]*umArgument
		Argument_schemes      []*umArgScheme
		Assumptions           []string
		caesArgSchemes        []*caes.Scheme
		caesLabels            map[string]caes.Label
		caesStatements        map[string]*caes.Statement
		caesWeighingFunctions map[string]caes.WeighingFunction
		Issues                map[string]*umIssue
		Issue_schemes         map[string]*caes.IssueScheme //[]string
		Tests                 *umLabel
		Language              caes.Language
		Meta                  caes.Metadata
		References            map[string]caes.Metadata
		Statements            map[interface{}]interface{} // string || text: label:
		Weighing_functions    map[string]interface{}
	}
	// mapIface map[interface{}]interface{}

	umArgScheme struct {
		Id          string
		Assumptions []string
		caesWeight  caes.WeighingFunction
		Conclusions []string
		Deletions   []string
		Exceptions  []string
		Guards      []string
		Meta        caes.Metadata
		Premises    []string
		Variables   []string
		Weight      interface{}
		// string
		// Constant: float64
		// Criteria: {Hard: []int Soft: map[string]{Factor: float64 Values: map[string]float64}
		// Preference: []{Property: string Order: sting || []string
	}
	umArgument struct {
		id          string
		Conclusion  string
		Meta        caes.Metadata
		Premises    interface{}
		Scheme      string
		Parameters  []string
		umpremises  []umPremis
		Undercutter string
		Weigth      float64 //
	}
	umIssue struct {
		id           string
		Meta         caes.Metadata
		Positions    []string
		Standard     string
		caesStandard caes.Standard
	}
	umLabel struct {
		In        []string // to do string || [string, ..]
		Out       []string // to do string || [string, ..]
		Undecided []string // to do string || [string, ..]
	}
	umPremis struct {
		role string
		stmt string
	}
	constantWF struct {
		constant float64
	}
)

const spPlus = "    "

var sp0, sp1, sp2, sp3, sp4, sp5, sp6, sp7 string

var collOfSchemes map[string]*caes.Scheme                    // collecion of all defined schemes
var collOfWeighingFunctions map[string]caes.WeighingFunction // collection of all defined weighing functions
var collOfAssumptions []string                               // collection of all assumptions
var collOfStatements map[string]*caes.Statement              // collection of statements
var collOfWF2source map[string]interface{}                   // collection of weighing function to her definition
// string (Basic)

// Import

func Import(inFile io.Reader) (*caes.ArgGraph, error) {

	collOfWeighingFunctions = caes.BasicWeighingFunctions
	collOfSchemes = caes.BasicSchemes
	collOfAssumptions = []string{}
	collOfWF2source = map[string]interface{}{}

	data, err := ioutil.ReadAll(inFile)
	if err != nil {
		return nil, err
	}
	// fmt.Printf("Read-Datei: \nErr: %v len(data): %v \n", err, len(data))
	m := new(argMapGraph)

	//	m := make(map[interface{}]interface{})
	err = yaml.Unmarshal(data, m)
	if err != nil {
		return nil, err
	}
	c, err := argMapGraph2caes(m)
	if err != nil {
		return nil, err
	}
	c.Theory.InitSchemeIndex()
	return c, nil
	//	return iface2caes(m)
}

func argMapGraph2caes(m *argMapGraph) (*caes.ArgGraph, error) {

	m, err := scanArgMapGraph(m)
	if err != nil {
		return caes.NewArgGraph(), err
	}
	c, err := caesArgMapGraph2caes(m)
	// fmt.Printf(" >>> End-Assumptions: %v \n", c.Assumptions)
	return c, err
	// return caesArgMapGraph2caes(m)
}

// func iface2caes(m map[interface{}]interface{}) (caesAg *caes.ArgGraph, err error) {
func scanArgMapGraph(m *argMapGraph) (*argMapGraph, error) {

	// set issue-id
	// ------------

	for id, iss := range m.Issues {
		iss.id = id
		// fmt.Printf(" issue: %s standard: \"%s\"\n", id, iss.Standard)
		switch iss.Standard {
		case "", "PE", "pe":
			iss.caesStandard = caes.PE
		case "CCE", "cce":
			iss.caesStandard = caes.CCE
		case "BRD", "brd":
			iss.caesStandard = caes.BRD
		default:
			return nil,
				errors.New("*** Error: issues: ... standard: expected PE, CCE, BRD, wrong: " + iss.Standard + " \n")
		}
	}
	// scan Stantements set caesStatements
	// -----------------------------------
	cs, err := iface2statement(m.Statements, map[string]*caes.Statement{})
	if err != nil {
		return nil, err
	}
	m.caesStatements = cs
	collOfStatements = cs
	// set argument-id and scan premises
	// ---------------------------------
	for id, arg := range m.Arguments {
		// for id, _ := range m.Arguments {
		arg.id = id
		// fmt.Printf(" Arg.: %s, Premises: %v \n", id, arg.Premises)
		arg.umpremises, err = iface2premises(arg.Premises)
		if err != nil {
			return nil, err
		}
	}
	// for _, arg := range m.Arguments {
	// 	fmt.Printf(" Arg.: %s, umpremises: %v \n", arg.id, arg.umpremises)
	// }

	// scan Labels
	// -----------
	m.caesLabels = labels2caes(m.Tests)
	// scan Assumtions
	// ---------------
	if m.Assumptions != nil && len(m.Assumptions) != 0 {
		for _, stat := range m.Assumptions {
			stat = normString(stat)
			collOfAssumptions = append(collOfAssumptions, stat)
		}
	}
	// scan weighing_functions
	// -----------------------
	m.caesWeighingFunctions = map[string]caes.WeighingFunction{}
	for name, body := range m.Weighing_functions {
		wf, err := iface2weighfunc(body, name, collOfWeighingFunctions)
		if err != nil {
			return nil, err
		}
		if wf != nil {
			m.caesWeighingFunctions[name] = wf
			collOfWeighingFunctions[name] = wf
		}
	}

	// scan argument_scheme
	// --------------------

	for _, argS := range m.Argument_schemes {
		name := argS.Id
		// scan weight in argument_schemes
		argS.caesWeight, err = iface2weighfunc(argS.Weight, name, collOfWeighingFunctions)
		if err != nil {
			return nil, err
		}
		//		// scan premises in argument_schemes
		//		//  to do
		//		argS.caesPremises, err = iface2mapStringString("premises", argS.Premises)
		//		if err != nil {
		//			return nil, err
		//		}
		//		// scan assumptions in argument_schemes
		//		// to do
		//		argS.caesAssumptions, err = iface2mapStringString("assumption", argS.Assumptions)
		//		if err != nil {
		//			return nil, err
		//		}
		//		// scan exceptions in argument_schemes
		//		// to do
		//		argS.caesExceptions, err = iface2mapStringString("exceptions", argS.Exceptions)
		//		if err != nil {
		//			return nil, err
		//		}
	}
	// scan argument_scheme and set caesArgSchemes
	m.caesArgSchemes = []*caes.Scheme{}
	for _, as := range m.Argument_schemes {
		id := as.Id
		s := caes.Scheme{Id: id, Metadata: as.Meta, Variables: as.Variables, Weight: as.caesWeight,
			Premises: as.Premises, Assumptions: as.Assumptions, Exceptions: as.Exceptions,
			Deletions: normStringVec(as.Deletions),
			Guards:    normStringVec(as.Guards), Conclusions: normStringVec(as.Conclusions)}
		m.caesArgSchemes = append(m.caesArgSchemes, &s)
		collOfSchemes[id] = &s
	}
	return m, nil
}

func labels2caes(ul *umLabel) map[string]caes.Label {
	ml := map[string]caes.Label{}
	if ul != nil {
		for _, in := range ul.In {
			ml[in] = caes.In
		}
		for _, out := range ul.Out {
			ml[out] = caes.Out
		}
		for _, undec := range ul.Undecided {
			ml[undec] = caes.Undecided
		}
	}
	return ml
}

func caesArgMapGraph2caes(m *argMapGraph) (caesAg *caes.ArgGraph, err error) {
	// create Theory
	// =============

	theory := caes.NewTheory()
	theory.Language = m.Language
	theory.WeighingFunctions = m.caesWeighingFunctions
	theory.ArgSchemes = m.caesArgSchemes
	theory.IssueSchemes = m.Issue_schemes

	// create ArgGraph
	// ===============
	caesAg = caes.NewArgGraph()
	caesAg.Metadata = m.Meta
	caesAg.References = m.References
	caesAg.Theory = theory

	// Metadata
	// --------
	// fmt.Printf("   ---  Metadata --- \n %v \n ------End Metadata --- \n", caesAg.Metadata)
	// References
	// ----------
	// fmt.Printf("   ---  References --- \n %v \n ------End References --- \n", caesAg.References)
	// Statement
	// ---------
	found := false
	caesAg.Statements = map[string]*caes.Statement{}
	for _, refYamlStat := range m.caesStatements {
		// fmt.Printf(" Statement: %s \n", refYamlStat.Id)
		caesAg.Statements[refYamlStat.Id] = refYamlStat
	}
	// fmt.Printf("   ---  Statements --- \n %v \n ------End Statements --- \n", caesAg.Statements)
	// Issue
	// -----
	// first = true
	caesAg.Issues = map[string]*caes.Issue{}
	for yamlIssue_Id, yamlIssue_Val := range m.Issues {
		caes_Issue := &caes.Issue{Id: yamlIssue_Id, Metadata: yamlIssue_Val.Meta, Standard: yamlIssue_Val.caesStandard}
		caesAg.Issues[yamlIssue_Id] = caes_Issue
		// References: Issue.Positions --> []*Statement, Statement.Issue --> *Issue
		for _, yamlIssue_Pos := range yamlIssue_Val.Positions {
			yamlIssue_Pos = normString(yamlIssue_Pos)
			found = false
		LoopIss:
			for _, caesAg_Stat := range caesAg.Statements {
				// fmt.Printf("   Compare Position: %s in statement %s \n", yamlIssue_Pos, caesAg_Stat.Id)
				if yamlIssue_Pos == caesAg_Stat.Id {
					found = true
					// fmt.Printf("   Position: %s \n", yamlIssue_Pos)
					if caes_Issue.Positions == nil {
						caes_Issue.Positions = []*caes.Statement{caesAg_Stat}
					} else {
						caes_Issue.Positions = append(caes_Issue.Positions, caesAg_Stat)
					}
					if caesAg_Stat.Issue == nil {
						caesAg_Stat.Issue = caes_Issue
					} else {
						if caes_Issue.Id != caesAg_Stat.Issue.Id {
							return caesAg, errors.New(" *** Semantic Error: Statement: " + caesAg_Stat.Id + ", with two issues: " + caes_Issue.Id + ", " + caesAg_Stat.Issue.Id + "\n")
						}
					}
					break LoopIss
				}
			}
			if !found {
				return caesAg, errors.New(" *** Semantic Error: Position " + yamlIssue_Pos + ", from Issue: " + caes_Issue.Id + ", is not a Statement-ID\n")
			}
		}
	}

	// fmt.Printf("   ---  Issues --- \n %v \n ------End Issuess --- \n", caesAg.Issues)

	// Arguments
	// first = true
	caesAg.Arguments = map[string]*caes.Argument{}
	for yamlArg_Id, yamlArg_Val := range m.Arguments {
		// fmt.Printf(" Argument2caes: %s\n", yamlArg_Id)
		caesArg := &caes.Argument{Id: yamlArg_Id, Metadata: yamlArg_Val.Meta, Weight: yamlArg_Val.Weigth}
		caesAg.Arguments[yamlArg_Id] = caesArg
		/* if first {
			caesAg.Arguments = []*caes.Argument{caesArg}
			first = false
		} else {
			caesAg.Arguments = append(caesAg.Arguments, caesArg)
		}*/
		// References: Argument.Conclusion --> *Statement, Statement.Args --> []*Argument
		found := false
		conc := normString(yamlArg_Val.Conclusion)
	LoopC:
		for _, caesArg_Stat := range caesAg.Statements {
			if conc == caesArg_Stat.Id {
				caesArg.Conclusion = caesArg_Stat
				found = true
				if caesArg_Stat.Args == nil {
					caesArg_Stat.Args = []*caes.Argument{caesArg}
				} else {
					caesArg_Stat.Args = append(caesArg_Stat.Args, caesArg)
				}
				break LoopC
			}
		}
		if !found {
			return caesAg, errors.New(" *** Semantic Error: Conclusion: " + yamlArg_Val.Conclusion + ", from Argument: " + yamlArg_Id + ", is not a Statement-ID\n")
		}

		// References: Argument.undercutter --> *Statement,
		// No undercutter in Statement.Args --> []*Argument
		if yamlArg_Val.Undercutter != "" {
			found = false
			ucut := normString(yamlArg_Val.Undercutter)
		LoopN:
			for _, caesArg_Stat := range caesAg.Statements {
				if ucut == caesArg_Stat.Id {
					found = true
					caesArg.Undercutter = caesArg_Stat
					break LoopN
				}
			}
			if !found {
				return caesAg, errors.New(" *** Semantic Error: Undercutter: " + yamlArg_Val.Undercutter + ", from Argument: " + yamlArg_Id + ", is not a Statement-ID\n")
			}
		}
		// Argument.Premises
		// fmt.Printf(" umpremises:%s \n", yamlArg_Val.umpremises)
		for _, yamlArg_Prem := range yamlArg_Val.umpremises {
			// fmt.Printf(" premises:%s \n", yamlArg_Prem.stmt)
			prem_stat, ok := collOfStatements[yamlArg_Prem.stmt]
			if !ok {
				return caesAg, errors.New(" *** Semantic Error: Premise: " + yamlArg_Prem.stmt + ", from Argument: " + yamlArg_Id + ", is not a Statement-ID\n")
			}
			if prem_stat == nil {
				// fmt.Printf("\n *** Prem Stat == nil für %s \n", yamlArg_Prem.stmt)
			} else {
				// fmt.Printf(" \n +++ Prem_Stat: %s für %s \n", prem_stat.Id, yamlArg_Prem.stmt)
			}
			caes_prem := caes.Premise{Stmt: prem_stat, Role: yamlArg_Prem.role}
			if caesArg.Premises == nil {
				caesArg.Premises = []caes.Premise{caes_prem}
			} else {
				caesArg.Premises = append(caesArg.Premises, caes_prem)
			}
		}
		// Scheme
		if yamlArg_Val.Scheme != "" {
			// fmt.Printf(" Coll of Schemes: %v \n", collOfSchemes)
			scheme, ok := collOfSchemes[yamlArg_Val.Scheme]
			if !ok {
				return caesAg, errors.New(" *** Semantic Error: Scheme: " + yamlArg_Val.Scheme + ", is not defined\n")
			}
			caesArg.Scheme = scheme
		} else {
			caesArg.Scheme = collOfSchemes["linked"]
		}
		// Parameters
		// fmt.Printf(" set parameter arg: %s Parameter: %v\n", caesArg.Id, yamlArg_Val.Parameters)
		caesArg.Parameters = yamlArg_Val.Parameters

	}
	// fmt.Printf("   ---  Arguments --- \n %v \n ------End Arguments --- \n", caesAg.Arguments)
	// Assumptions
	caesAg.Assumptions = []string{}
	for _, yamlAss := range collOfAssumptions {
		caesAg.Assumptions = append(caesAg.Assumptions, yamlAss)
		if caesAg.Theory == nil {
			found = false
			for _, caesArg_Stat := range caesAg.Statements {
				if yamlAss == caesArg_Stat.Id {
					found = true
				}
			}
			if !found {
				return caesAg, errors.New(" *** Semantic Error: Assumption: " + yamlAss + ", is not a Statement-ID\n")
			}
		}
	}
	// fmt.Printf(" caes-Assamptions: %v \n", caesAg.Assumptions)
	// Labels
	// if yamlLbls not empty
	// fmt.Printf(" Labels: %v \n", m.caesLabels)
	if m.caesLabels != nil && len(m.caesLabels) != 0 {
		caesAg.ExpectedLabeling = m.caesLabels
		for _, caesArg_Stat := range caesAg.Statements {
			lbl, found := m.caesLabels[caesArg_Stat.Id]
			if found == true {
				// fmt.Printf(" Label %s:%v\n", caesArg_Stat.Id, lbl)
				caesArg_Stat.Label = lbl
				// } else {
				// 	fmt.Printf(" ## Label not found %s:%v\n", caesArg_Stat.Id)
			}
		}
	}
	//check-Labels
	if caesAg.Theory == nil {
		for lbl_Id, lbl_val := range m.caesLabels {
			found = false
		LoopLbl:
			for _, caesArg_Stat := range caesAg.Statements {
				if lbl_Id == caesArg_Stat.Id {
					found = true
					break LoopLbl
				}
			}
			if !found {
				lbl_str := ""
				switch lbl_val {
				case caes.In:
					lbl_str = "In"
				case caes.Out:
					lbl_str = "Out"
				case caes.Undecided:
					lbl_str = "Undecided"
				}
				return caesAg, errors.New(" *** Semantic Error: " + lbl_str + "- Label: " + lbl_Id + ", is not a Statement-ID\n")
			}
		}
	}

	return caesAg, nil

}

func iface2mapStringString(lbl string, value interface{}) (result map[string]string, err error) {
	// and normelaise string
	result = map[string]string{}
	if value == nil {
		return nil, nil
	}
	switch tValue := value.(type) {
	case map[interface{}]interface{}:
		for key, val := range value.(map[interface{}]interface{}) {
			keyStr := ""
			switch key.(type) {
			case string:
				keyStr = key.(string)
			default:
				keyStr = fmt.Sprintf("%v", key)
			}
			valStr := ""
			switch val.(type) {
			case string:
				valStr = val.(string)
			default:
				valStr = fmt.Sprintf("%v", val)
			}
			result[keyStr] = normString(valStr)
		}
	case []interface{}:
		for idx, val := range value.([]interface{}) {
			result[fmt.Sprintf("%d", idx+1)] = normString(fmt.Sprintf("%v", val))
		}
	default:
		return result, errors.New(fmt.Sprintf(" *** Syntax Error: In %s not a map nor a list. Wrong type: %v \n", lbl, tValue))

	}
	return
}

func iface2string(value interface{}) string {
	var text string
	switch value.(type) {
	case string:
		text = value.(string)
	case int:
		text = strconv.FormatInt(int64(value.(int)), 10)
	case float32:
		text = strconv.FormatFloat(float64(value.(float32)), 'G', 10, 32)
	case float64:
		text = strconv.FormatFloat(value.(float64), 'G', 10, 64)
	case bool:
		if value.(bool) {
			text = "true"
		} else {
			text = "false"
		}
	default:
		text = "???"
	}
	return text
}

func iface2language(value interface{}, yamlLang caes.Language) (caes.Language, error) {

	switch subT := value.(type) {
	case map[interface{}]interface{}:
		for langkey, langval := range value.(map[interface{}]interface{}) {
			switch keyT := langkey.(type) {
			case string:
				switch valT := langval.(type) {
				case string:
					yamlLang[langkey.(string)] = langval.(string)
				default:
					return yamlLang, errors.New("*** Error language: (value-Type)" + fmt.Sprintf("%v", valT) + " of key " + fmt.Sprintf("%v", langkey) + "\n")
				}
			default:
				return yamlLang, errors.New("*** Error language: (key-Type)" + fmt.Sprintf("%v", keyT) + "\n")

			}

		}
		return yamlLang, nil
	default:
		return yamlLang, errors.New("*** Error language-elements: (Type)" + fmt.Sprintf("%v", subT) + "\n")

	}
}

func iface2namedweighfunc(value interface{}, yamlWeighFunc map[string]caes.WeighingFunction) (map[string]caes.WeighingFunction, error) {
	switch subT := value.(type) {
	case map[interface{}]interface{}:
		for wf_name, wf_body := range value.(map[interface{}]interface{}) {
			switch nameT := wf_name.(type) {
			case string:
				wf, err := iface2weighfunc(wf_body, wf_name.(string), yamlWeighFunc)
				if err != nil {
					return yamlWeighFunc, err
				}
				if wf != nil {
					yamlWeighFunc[wf_name.(string)] = wf
					collOfWeighingFunctions[wf_name.(string)] = wf
				}
			default:
				return yamlWeighFunc, errors.New("*** Error weighing function name-key, string: expected, not type " + fmt.Sprintf("%v", nameT) + "\n")
			}
		}
	default:
		return yamlWeighFunc, errors.New("*** Error <weighing function name:> <body> expectes, not type " + fmt.Sprintf("%v", subT) + "\n")
	}
	return yamlWeighFunc, nil
}

func iface2weighfunc(value interface{}, name string, yamlWeighFunc map[string]caes.WeighingFunction) (caes.WeighingFunction, error) {
	var wf caes.WeighingFunction
	var err error = nil
	switch subT := value.(type) {
	case map[interface{}]interface{}:
		// preference: ... , criteria: ..., constant: <0.0 ... 1.0>
		for wf_type, wf_body := range value.(map[interface{}]interface{}) {
			switch typeT := wf_type.(type) {
			case string:
				switch strings.ToLower(wf_type.(string)) {
				case "preference":
					po := []caes.PropertyOrder{}
					wf, po, err = iface2preference(wf_body)
					collOfWF2source[name] = po
				case "criteria":
					c := &caes.Criteria{}
					wf, c, err = iface2criteria(wf_body)
					collOfWF2source[name] = c
				case "constant":
					fl := 0.0
					wf, fl, err = iface2floatWF(wf_body)
					collOfWF2source[name] = constantWF{constant: fl}
				default:
					return nil, errors.New("*** Error in weighing function 'preference:', 'criteria:' or 'constant:' expected, not " + fmt.Sprintf("%v", wf_type) + "\n")
				}
			default:
				return wf, errors.New("*** Error 'preference:', 'criteria:' or 'constant:' expected, not type " + fmt.Sprintf("%v", typeT) + "\n")
			}
		}
	case string:
		str := strings.ToLower(value.(string))
		// wf, in := caes.BasicWeighingFunctions[str]
		// if in {
		// 	return wf, nil
		//}
		// search in defines named weighing functions
		wf, in := yamlWeighFunc[str]
		// name of a defined weighing function
		if in {
			return wf, nil
		} else {
			return nil, errors.New("*** Error weighing function '" + fmt.Sprintf("%v", value) + "' is not defined")
		}
	case nil:
		return caes.LinkedWeighingFunction, nil
		// return nil, errors.New("*** Internal Error: Cannot find the defoult weighting function 'linked'\n")
	default:
		return nil, errors.New("*** Error: Not a weighting function " + fmt.Sprintf("%v", value) + ", type " + fmt.Sprintf("%v", subT) + "\n")
	}
	return wf, err
}

func iface2preference(pref interface{}) (caes.WeighingFunction, []caes.PropertyOrder, error) {
	// [[property: ... order: ..],[property: ... order: ...] ...]
	switch pT := pref.(type) {
	case []interface{}:
		prefArr := pref.([]interface{})
		po := make([]caes.PropertyOrder, len(prefArr))
		// fmt.Printf(" [")
		for i, po_ele := range prefArr {
			switch poT := po_ele.(type) {
			case map[interface{}]interface{}:
				property, order, values := "", caes.Descending, []string{}
				for key, val := range po_ele.(map[interface{}]interface{}) {
					switch key.(type) {
					case string:
						str := strings.ToLower(key.(string))
						switch str {
						case "property":
							switch valT := val.(type) {
							case string:
								property = val.(string)
							default:
								return nil, nil,
									errors.New("*** Error property-value must be a string and not type" + fmt.Sprintf("%v", valT) + "\n")
							}
						case "order":
							switch valT := val.(type) {
							case string: // ascending or descending
								switch strings.ToLower(val.(string)) {
								case "ascending":
									order = caes.Ascending
								case "descending":
									order = caes.Descending
								default:
									return nil, nil,
										errors.New("*** Error expected in preference: order: 'ascending' or 'descending' and not '" + fmt.Sprintf("%v", val) + "'\n")
								}
							case []interface{}: // [val1, val2, ...]
								orderArr := val.([]interface{})
								values = make([]string, len(orderArr))
								// fmt.Printf(" [")
								for io, o_ele := range orderArr {
									switch oT := o_ele.(type) {
									case string:
										values[io] = strings.ToLower(o_ele.(string))
									default:
										return nil, nil,
											errors.New("*** Error expected string-list in preference: order: not type '" + fmt.Sprintf("%v", oT) + "'\n")
									}
								}
							default:
								return nil, nil,
									errors.New("*** Error in preference: order: wrong type '" + fmt.Sprintf("%v", valT) + "'\n")
							}
						default:
							return nil, nil,
								errors.New("*** Error in preference: wrong key '" + fmt.Sprintf("%v", key) + "' \n")
						}
					}

				}
				po[i] = caes.PropertyOrder{Property: property, Order: order, Values: values}
			default:
				return nil, nil,
					errors.New("*** Error preference: has a wrong type '" + fmt.Sprintf("%v", poT) + "'\n")

			}

		}
		return caes.PreferenceWeighingFunction(po), po, nil
	case map[interface{}]interface{}: // preference: and single property: order:
		po := make([]caes.PropertyOrder, 1)
		property, order, values := "", caes.Descending, []string{}
		for key, val := range pref.(map[interface{}]interface{}) {
			switch key.(type) {
			case string:
				str := strings.ToLower(key.(string))
				switch str {
				case "property":
					switch valT := val.(type) {
					case string:
						property = val.(string)
					default:
						return nil, nil,
							errors.New("*** Error property-value must be a string and not type" + fmt.Sprintf("%v", valT) + "\n")
					}
				case "order":
					switch valT := val.(type) {
					case string: // ascending or descending
						switch strings.ToLower(val.(string)) {
						case "ascending":
							order = caes.Ascending
						case "descending":
							order = caes.Descending
						default:
							return nil, nil,
								errors.New("*** Error expected in preference: order: 'ascending' or 'descending' and not '" + fmt.Sprintf("%v", val) + "'\n")
						}
					case []interface{}: // [val1, val2, ...]
						orderArr := val.([]interface{})
						values = make([]string, len(orderArr))
						// fmt.Printf(" [")
						for io, o_ele := range orderArr {
							switch oT := o_ele.(type) {
							case string:
								values[io] = strings.ToLower(o_ele.(string))
							default:
								return nil, nil,
									errors.New("*** Error expected string-list in preference: order: not type '" + fmt.Sprintf("%v", oT) + "'\n")
							}
						}
					default:
						return nil, nil,
							errors.New("*** Error in preference: order: wrong type '" + fmt.Sprintf("%v", valT) + "'\n")
					}
				default:
					return nil, nil,
						errors.New("*** Error in preference: wrong key '" + fmt.Sprintf("%v", key) + "'\n")
				}
			}
		}
		po[0] = caes.PropertyOrder{Property: property, Order: order, Values: values}
		return caes.PreferenceWeighingFunction(po), po, nil
	default:
		return nil, nil,
			errors.New("*** ERROR wrong type afer preference: '" + fmt.Sprintf("%v", pT) + "'\n")
	}
	return nil, nil, nil // dummy return
}

func iface2criteria(cr interface{}) (caes.WeighingFunction, *caes.Criteria, error) {
	c := caes.Criteria{HardConstraints: []int{}, SoftConstraints: map[string]caes.SoftConstraint{}}

	switch subT := cr.(type) {
	case map[interface{}]interface{}:
		for key, value := range cr.(map[interface{}]interface{}) {
			switch key.(type) {
			case string:
				switch strings.ToLower(key.(string)) {
				case "hard":
					switch valT := value.(type) {
					case []interface{}:
						for _, ele := range value.([]interface{}) {
							switch eleT := ele.(type) {
							case int:
								c.HardConstraints = append(c.HardConstraints, ele.(int))
							default:
								return nil, nil,
									errors.New("*** Error element of criteria: hard: is not a 'string', found type '" + fmt.Sprintf("%v", eleT) + "'\n")
							}
						}
					default:
						return nil, nil,
							errors.New("*** Error value of criteria: hard: is not a list, found type '" + fmt.Sprintf("%v", valT) + "'\n")
					}
				case "soft":
					so, err := iface2soft(value, c.SoftConstraints)
					if err != nil {
						return nil, nil, err
					}
					c.SoftConstraints = so
				default:
					return nil, nil,
						errors.New("*** Error unknown key in criteria: '" + fmt.Sprintf("%v", key) + "', expected keys 'hard:' or 'soft:' \n")
				}
			default:
				return nil, nil,
					errors.New("*** Error key in criteria mus be a string, not'" + fmt.Sprintf("%v", key) + "'\n")
			}
		}
		return caes.CriteriaWeighingFunction(&c), &c, nil
	default:
		return nil, nil,
			errors.New("*** Error unknown type after criteria: '" + fmt.Sprintf("%v", subT) + "', expected keys-struct with keys 'hard:' and 'soft'\n")
	}
	return caes.CriteriaWeighingFunction(&c), &c, nil
}

func iface2soft(body interface{}, soft map[string]caes.SoftConstraint) (map[string]caes.SoftConstraint, error) {

	// fmt.Printf(" soft='%v'\n", body)

	switch body.(type) {
	case map[interface{}]interface{}:
		for key, value := range body.(map[interface{}]interface{}) {
			switch keyT := key.(type) { //
			case string: // name of the soft criteria for ex. speed, safety, price, type
				// --------------------------

				switch valT := value.(type) {
				case map[interface{}]interface{}:

					factor := 0.0
					normValues := map[string]float64{} // collect the NormelizedValues

					for subkey, subvalue := range value.(map[interface{}]interface{}) {
						switch subkT := subkey.(type) { // name of the NormalizedValues
						case string:
							// factor or values
							switch strings.ToLower(subkey.(string)) {
							case "factor":
								switch subvT := subvalue.(type) {
								case float64:
									factor = subvalue.(float64)
									//									if factor < 0.0 || factor > 1.0 {
									//										return soft,
									//											errors.New("*** Error value 0.00 ... 1.00 exspected in soft criteria '" + subkey.(string) + "' factor:, not '" +
									//												fmt.Sprintf("%v'\n", subvalue))
									//									}
								case int:
									factor = float64(subvalue.(int))
								default:
									return soft,
										errors.New("*** Error float-value exspected in soft criteria '" + subkey.(string) + "' factor:, not '" +
											fmt.Sprintf("%v' (type=%v)\n", subvalue, subvT))
								}
							case "values":
								switch subvT := subvalue.(type) {
								case map[interface{}]interface{}:
									float := 0.0
									for ky, fl := range subvalue.(map[interface{}]interface{}) {
										k := ""
										switch kyT := ky.(type) {
										case string:
											k = ky.(string)
										default:
											return soft,
												errors.New("*** Error in soft criteria '" + subkey.(string) + "' values: a key: expected, not:" +
													fmt.Sprintf("'%v' (type=%v)\n", ky, kyT))
										}
										switch flT := fl.(type) {
										case float64:
											float = fl.(float64)
										case int:
											float = float64(fl.(int))
										default:
											return soft,
												errors.New("*** Error in soft criteria '" + subkey.(string) + "' values: " +
													k + ":, float or int expected, not" +
													fmt.Sprintf("'%v' (type=%v)", fl, flT))
										}

										if float < 0.0 || float > 1.0 {
											return soft,
												errors.New("*** Error value 0.00 ... 1.00 expected in soft criteria '" + subkey.(string) + "' values: " +
													k + ":, not '" +
													fmt.Sprintf("%v'\n", float))
										}
										normValues[k] = float
									}
								default:
									return soft,
										errors.New("*** Error expected a 'key: float'-list, not " +
											fmt.Sprintf("'%v'\n", subvT))
								}
							default:
								return soft,
									errors.New("*** Error expected 'factor' or 'values' in soft criteria '" +
										fmt.Sprintf("%s", key) + "'and not '" + fmt.Sprintf("%v", subkey) + "'\n")
							}
						default:
							return soft,
								errors.New("*** ERROR in creteria: soft: " + fmt.Sprintf("%v", key) + ": expected a string-key and not '" + fmt.Sprintf("%v", subkT) + "'\n")
						}
					}
					soft[key.(string)] = caes.SoftConstraint{Factor: factor, NormalizedValues: normValues} // set the NormelizesValues
				default:
					return soft,
						errors.New("*** Error in criteria: soft: key expected factor: ... values: ...-list not '" + fmt.Sprintf("%v' (type=", value, valT) + ")\n")

				}
			default:
				return soft,
					errors.New("*** Error in criteria: soft: expected a string-key and not '" + fmt.Sprintf("%v' (type=", key, keyT) + ")\n")
			}
		}
	default:
		return soft,
			errors.New("*** Error after soft: a key-list expected, not '" +
				fmt.Sprintf("'%v'\n", body))
	}
	return soft, nil
}

func iface2floatWF(fl interface{}) (caes.WeighingFunction, float64, error) {
	f := 0.0
	switch flT := fl.(type) {
	case float32:
		f = float64(fl.(float32))
	case float64:
		f = fl.(float64)
	case int:
		f = float64(fl.(int))
	case string:
		_, err := fmt.Scanf(fl.(string), "%f", &f)
		if err != nil {
			return nil, 0.0,
				errors.New("*** ERROR in weighing function constant: expected fload and not type '" + flT + "'\n")
		}
	}
	return caes.ConstantWeighingFunction(f), f, nil
}

func iface2labels(value interface{}, yamlLbls map[string]caes.Label) (map[string]caes.Label, error) {
	var err error
	switch subT := value.(type) {
	case map[interface{}]interface{}:
		for lblkey, lblvalue := range value.(map[interface{}]interface{}) {
			switch strings.ToLower(lblkey.(string)) {
			case "in":
				yamlLbls, err = iface2lbl_1(lblvalue, yamlLbls, caes.In)
			case "out":
				yamlLbls, err = iface2lbl_1(lblvalue, yamlLbls, caes.Out)
			case "undecided":
				yamlLbls, err = iface2lbl_1(lblvalue, yamlLbls, caes.Undecided)
			}
		}
	default:
		return yamlLbls, errors.New("*** Error labels: (Type)" + fmt.Sprintf("%v", subT) + "\n")

	}
	return yamlLbls, err
}

func iface2lbl_1(inArg interface{}, yamlLbls map[string]caes.Label, label caes.Label) (map[string]caes.Label, error) {
	// fmt.Printf("labels:\n   %v: ", label)
	switch intype := inArg.(type) {
	case []interface{}:
		// fmt.Printf(" [")
		for idx, stat := range inArg.([]interface{}) {
			switch stype := stat.(type) {
			case string:
				yamlLbls[stat.(string)] = label
				if idx != 0 {
					// fmt.Printf(", ")
				}
				// fmt.Printf("%s", stat.(string))
			default:
				return yamlLbls, errors.New("*** Error labels [(Type)]:" + fmt.Sprintf("%v", stype) + "\n")
			}
		}
		// fmt.Printf("]")
	case string:
		yamlLbls[inArg.(string)] = label
	default:
		return yamlLbls, errors.New("*** Error labels (Type):" + fmt.Sprintf("%v", intype) + "\n")

	}
	return yamlLbls, nil
}

func iface2metadata(value interface{}, meta caes.Metadata) (caes.Metadata, error) {
	switch subT := value.(type) {
	case map[interface{}]interface{}:
		for metakey, metavalue := range value.(map[interface{}]interface{}) {
			if metakey == nil || metavalue == nil {
				continue
			}
			switch metavalue.(type) {
			case string:
				if metavalue.(string) != "" {
					meta[metakey.(string)] = metavalue.(string)
				}
			case int, float32, float64, bool:
				meta[metakey.(string)] = iface2string(metavalue)
			case map[interface{}]interface{}:
				var err error
				meta[metakey.(string)], err = iface2metadata(metavalue, caes.Metadata{})
				return meta, err
			}
		}
	default:
		return meta, errors.New("*** Error metadata (Type):" + fmt.Sprintf("%v", subT) + "\n")
	}
	return meta, nil
}

func iface2statement(value interface{}, yamlStats map[string]*caes.Statement) (map[string]*caes.Statement, error) {

	// fmt.Printf("Statements: \n")
	// switch t := value.(type) {
	switch value.(type) {
	case map[interface{}]interface{}:
		for st_key, st_value := range value.(map[interface{}]interface{}) {
			switch st_key.(type) {
			case string: // ok
			default:
				return nil,
					errors.New("*** Error statemment name must be a string not '" + fmt.Sprintf("%v", st_key) + "'\n")
			}
			keyStr := normString(st_key.(string))
			switch st_value.(type) {
			case string:
				yamlStats[keyStr] = &caes.Statement{
					Id:    keyStr,
					Text:  st_value.(string),
					Label: caes.Undecided,
				}
				// fmt.Printf(" %v: %v \n", keyStr, st_value.(string))
			case int, float32, float64, bool:
				yamlStats[keyStr] = &caes.Statement{
					Id:    keyStr,
					Text:  iface2string(st_value),
					Label: caes.Undecided,
				}
				// fmt.Printf(" %v: %v \n", keyStr, iface2string(st_value))
			case map[interface{}]interface{}:
				// fmt.Printf("   %v:\n", keyStr)
				v, err := iface2xstatement(st_value, &caes.Statement{Id: keyStr, Label: caes.Undecided})
				if err != nil {
					return yamlStats, err
				}
				yamlStats[keyStr] = v
			}
		}
	default:
		// fmt.Printf(" Type: %v \n", t)
	}
	return yamlStats, nil
}

func iface2xstatement(st_value interface{}, stat *caes.Statement) (*caes.Statement, error) {
	var err error
	switch st_value.(type) {
	case map[interface{}]interface{}: // OK
	default:
		return nil, errors.New("*** Error subexpression of statement should be a 'key: value'-list, not '" + fmt.Sprintf("%v", st_value) + "'\n")
	}
	for st_subkey, st_subvalue := range st_value.(map[interface{}]interface{}) {
		switch st_subkey.(type) {
		case string: // OK
		default:
			return nil,
				errors.New("*** Error key of a subexpression of statement must be a string, not '" + fmt.Sprintf("%v", st_subkey) + "'\n")
		}
		st_subkey_lo := strings.ToLower(st_subkey.(string))
		switch st_subkey_lo {
		case "meta", "metadata":
			// fmt.Printf("      meta:\n")
			if stat.Metadata == nil {
				stat.Metadata, err = iface2metadata(st_subvalue, caes.Metadata{})
			} else {
				stat.Metadata, err = iface2metadata(st_subvalue, stat.Metadata)
			}
			if err != nil {
				return stat, err
			}
		case "text":
			stat.Text = iface2string(st_subvalue)
			// fmt.Printf("      text: %s \n", stat.Text)
		case "assumed":
			switch t := st_subvalue.(type) {
			case bool:
				if st_subvalue.(bool) {
					collOfAssumptions = append(collOfAssumptions, stat.Id)
				}
				// fmt.Printf("     assumed: %v \n", stat.Assumed)
			case int:
				if st_subvalue.(int) == 1 {
					collOfAssumptions = append(collOfAssumptions, stat.Id)
				}
				// fmt.Printf("     assumed: %v \n", stat.Assumed)
			default:
				return stat, errors.New("*** ERROR: Assumed value not bool: " + fmt.Sprintf("%v", t) + " \n")
			}
		case "value", "label":
			switch strings.ToLower(iface2string(st_subvalue)) {
			case "in":
				stat.Label = caes.In
			case "out":
				stat.Label = caes.Out
			case "undecided":
				stat.Label = caes.Undecided
			}
		}
	}
	return stat, nil
}

func iface2issues(value interface{}, yamlIssues map[string]umIssue) (map[string]umIssue, error) {
	// fmt.Printf("issues:\n")
	switch value.(type) {
	case map[interface{}]interface{}:
		for issueName, issueMap := range value.(map[interface{}]interface{}) {
			switch iT := issueName.(type) {
			case string:
				issueNameStr := issueName.(string)
				// fmt.Printf("   %s:\n", issueNameStr)
				issue := umIssue{id: issueNameStr, caesStandard: caes.PE}
				switch issueMap.(type) {
				case map[interface{}]interface{}:
					for issueKey, issueValue := range issueMap.(map[interface{}]interface{}) {
						switch issueKey.(type) {
						case string:
							// fmt.Printf("      %s: ", issueKey.(string))
							issueKey_lo := strings.ToLower(issueKey.(string))
							switch issueKey_lo {
							case "positions":
								// switch ivT := issueValue.(type) {
								switch issueValue.(type) {
								case []interface{}:
									issueArr := issueValue.([]interface{})
									issue.Positions = make([]string, len(issueArr))
									// fmt.Printf(" [")
									for i, ele_i := range issueArr {
										if i != 0 {
											// fmt.Printf(" ,")
										}
										switch eleT := ele_i.(type) {
										case string:
											issue.Positions[i] = ele_i.(string)
											// fmt.Printf("%v", issue.positions[i])
										default:
											return yamlIssues,
												errors.New("*** Error Position value-type: " + eleT.(string) + " \n")
										}
									}
									// fmt.Printf("]\n")
								case string: /* a singel position */
									issue.Positions = make([]string, 1)
									issue.Positions[0] = issueValue.(string)
									// fmt.Printf(" = %s \n", issueValue.(string))
								default:
									// fmt.Printf("      position-type: %v \n", ivT)
								}

							case "standard":
								switch ivT := issueValue.(type) {
								case string:
									issueValueStr := issueValue.(string)
									// fmt.Printf("%s\n", issueValueStr)
									switch issueValueStr {
									case "PE", "pe":
										issue.caesStandard = caes.PE
									case "CCE", "cce":
										issue.caesStandard = caes.CCE
									case "BRD", "brd":
										issue.caesStandard = caes.BRD
									default:
										return yamlIssues,
											errors.New("*** Error: issues: ... standard: expected PE, CCE, BRD, wrong: " + fmt.Sprintf("%v", issueValue) + " \n")
									}
								default:
									return yamlIssues,
										errors.New("*** Error: unexpected issue-value-type:" + fmt.Sprintf("%v", ivT) + "\n")
								}

							case "meta", "metadata":
								var err error
								if issue.Meta == nil {
									issue.Meta, err = iface2metadata(issueValue, caes.Metadata{})
								} else {
									issue.Meta, err = iface2metadata(issueValue, issue.Meta)
								}
								if err != nil {
									return yamlIssues, err
								}
							}
						}
					}
				}
				yamlIssues[issueNameStr] = issue
			default:
				return yamlIssues,
					errors.New("*** ERROR: Not a issue Name: " + fmt.Sprintf("%v", iT) + "\n")
			}
		}
	}
	return yamlIssues, nil
}

func iface2assumps(value interface{}, yamlAssumps map[string]bool) (map[string]bool, error) {
	// fmt.Printf("Assumptions: ")
	switch value.(type) {
	case []interface{}:
		// fmt.Printf("[")
		for i, str := range value.([]interface{}) {
			if i != 0 {
				// fmt.Printf(", ")
			}
			switch str.(type) {
			case string:
				yamlAssumps[str.(string)] = true
				// fmt.Printf("%s", str.(string))
			default:
				yamlAssumps[iface2string(str)] = true
				// fmt.Printf("if=%v", iface2string(str))
			}
		}
		// fmt.Printf("]\n")
	case string:
		yamlAssumps[value.(string)] = true
		// fmt.Printf(" %s \n", value.(string))
	default:
		return yamlAssumps,
			errors.New("*** ERROR: not a right assumption-value: " + fmt.Sprintf("%v", value) + "\n")
	}
	return yamlAssumps, nil
}

func iface2arguments(value interface{}, yamlArgs map[string]umArgument) (map[string]umArgument, error) {
	// fmt.Printf("Arguments: \n")
	switch value.(type) {
	case map[interface{}]interface{}:
		for argName, arg := range value.(map[interface{}]interface{}) {
			switch aT := argName.(type) {
			case string:
				// fmt.Printf("    %s: \n", argName.(string))
				var err error
				yamlArgs[argName.(string)], err = iface2argument(arg, umArgument{})
				if err != nil {
					return yamlArgs, err
				}
			default:
				return yamlArgs,
					errors.New("*** ERROR: Arguement-name is not a string: " + fmt.Sprintf("%v", aT) + "\n")
			}
		}
	}
	return yamlArgs, nil
}

func iface2argument(inArg interface{}, outArg umArgument) (umArgument, error) {
	var err error
	switch inArg.(type) {
	case map[interface{}]interface{}:
		for attName, attValue := range inArg.(map[interface{}]interface{}) {
			switch ant := attName.(type) {
			case string:
				attNameStr := strings.ToLower(attName.(string))
				// fmt.Printf("      %s:", attNameStr)
				switch attNameStr {
				case "conclusion":
					outArg.Conclusion = iface2string(attValue)
					// fmt.Printf(" %s \n", outArg.conclusion)
				case "premises":
					outArg.umpremises, err = iface2premises(attValue)
					if err != nil {
						return outArg, err
					}
					// fmt.Printf("\n")
				case "metadata", "meta":
					outArg.Meta, err = iface2metadata(attValue, caes.Metadata{})
					if err != nil {
						return outArg, err
					}
				case "weight":
					outArg.Weigth, err = iface2weight(attValue)
					if err != nil {
						return outArg, err
					}
				case "nas", "undercutter", "not app statement":
					outArg.Undercutter = iface2string(attValue)

					// fmt.Printf(" %s \n", outArg.undercutter)
				case "scheme":
					outArg.Scheme = iface2string(attValue)

					//					_, isIn := caes.BasicSchemes[outArg.scheme]
					//					if isIn == false {
					//						errStr := "*** ERROR: In argument wrong scheme value: " + outArg.scheme + "(expected: "
					//						first := true
					//						for schemeKey, _ := range caes.BasicSchemes {
					//							if first {
					//								errStr = errStr + schemeKey
					//								first = false
					//							} else {
					//								errStr = errStr + ", " + schemeKey
					//							}
					//						}
					//						return outArg,
					//							errors.New(errStr + ")\n")
					//					}
					//					// fmt.Printf(" %s \n", outArg.scheme)
				case "parameters":

					fmt.Printf(" parameters: %v\n", attValue)
					outArg.Parameters, err = iface2parameters(attValue)
					if err != nil {
						return outArg, err
					}
				default:
					return outArg,
						errors.New("*** ERROR: Wrong argument attribute: " + fmt.Sprintf("%v", attName) + " (expected: conclusion, premises, weight, undercutter, scheme or metadata)\n")
				}
			default:
				return outArg,
					errors.New("*** ERROR: Wrong argument attribute type (string expected): " + fmt.Sprintf("%v", ant) + "\n")
			}
		}
	}
	return outArg, nil
}

func iface2weight(attValue interface{}) (float64, error) {
	// weight := new(float64)
	weight := 0.0
	switch atype := attValue.(type) {
	case float32:
		// fl32 := attValue.(float32)
		// *weight = fl32.(float64)
		weight = attValue.(float64)
	case float64:
		weight = attValue.(float64)
	case int:
		intvalue := attValue.(int)
		if intvalue == 0 || intvalue == 1 {
			weight = attValue.(float64)
		} else {
			return 0.0, errors.New(("*** ERROR: Wrong weigth type (float expeced) integer value: " + fmt.Sprintf("%v", attValue)))
		}
	default:
		return 0.0, errors.New("*** ERROR: Wrong weigth type (float expected): " + fmt.Sprintf("%v", atype) + "\n")
	}
	// fmt.Printf("     weight: %v\n", *weight)
	return weight, nil
}

func iface2parameters(inArg interface{}) ([]string, error) {
	outArg := []string{}
	switch tinArg := inArg.(type) {
	case []interface{}:
		for _, para := range inArg.([]interface{}) {
			switch tpara := para.(type) {
			case string:
				outArg = append(outArg, para.(string))
			default:
				return nil,
					errors.New("*** Error: parameters: expected strings, not type" + fmt.Sprintf("%v", tpara) + "\n")
			}
		}
	default:
		return nil,
			errors.New("*** Error: parameters: expected list of strings, not type" + fmt.Sprintf("%v", tinArg) + "\n")
	}
	return outArg, nil
}

func iface2premises(inArg interface{}) ([]umPremis, error) {
	var outArg []umPremis
	switch inArg.(type) {
	case []interface{}:
		// fmt.Printf(" [")
		for idx, stat := range inArg.([]interface{}) {
			switch stat.(type) {
			case string:
				umP := umPremis{stmt: normString(stat.(string))}
				if outArg == nil {
					outArg = []umPremis{umP}
				} else {
					outArg = append(outArg, umP)
				}
				if idx != 0 {
					// fmt.Printf(", ")
				}
				// fmt.Printf("%s", stat.(string))

			}
		}
		// fmt.Printf("]")
	case map[interface{}]interface{}:
		// fmt.Printf("\n")
		for key, val := range inArg.(map[interface{}]interface{}) {
			premis := umPremis{}
			valStr := ""
			switch val.(type) {
			case string:
				valStr = val.(string)
			default:
				valStr = iface2string(val)
			}
			keyStr := ""
			switch key.(type) {
			case string:
				keyStr = key.(string)
			default:
				keyStr = iface2string(key)
			}
			keyStr = strings.ToLower(keyStr)
			premis.role = keyStr
			premis.stmt = normString(valStr)
			if outArg == nil {
				outArg = []umPremis{premis}
			} else {
				outArg = append(outArg, premis)
			}
		}
	}
	return outArg, nil
}

func iface2references(reference interface{}, yamlRefs map[string]caes.Metadata) (map[string]caes.Metadata, error) {
	var err error
	// fmt.Printf("references: \n")
	switch reference.(type) {
	case map[interface{}]interface{}:
		for refName, refBody := range reference.(map[interface{}]interface{}) {
			refNameStr := "???"
			switch refName.(type) {
			case string:
				refNameStr = refName.(string)
			default:
				refNameStr = iface2string(refName)
			}
			// fmt.Printf("    %s:\n", refNameStr)
			yamlRefs[refNameStr], err = iface2metadata(refBody, caes.Metadata{})
			if err != nil {
				return yamlRefs, err
			}
		}
	}
	return yamlRefs, nil
}

func normString(src string) (scr2 string) {
	// fmt.Printf(">%s<\n", src)
	t, ok := terms.ReadString(src)
	if ok {
		return t.String()
	}
	return src
}

func normStringVec(in []string) []string {
	for i, s := range in {
		in[i] = normString(s)
	}
	return in
}

// Export

func mkYamlString(str string) string {
	newstr := str
	if strings.ContainsAny(str, "\n\r:-?,[]{}#&!*|>'\"%@`") {
		if strings.Contains(str, "\"") {
			if strings.Contains(str, "'") {
				// mark \"
				newstr1 := []rune{}
				for _, ch := range str {
					if ch == '"' {
						newstr1 = append(newstr1, '\\')
					}
					newstr1 = append(newstr1, ch)
				}
				newstr = "\"" + string(newstr1) + "\""
			} else {
				newstr = "'" + newstr + "'"
			}
		} else {
			newstr = "\"" + newstr + "\""
		}
	}
	return newstr
}

func writeMetaData(f io.Writer, sp1 string, sp2 string, md caes.Metadata) {
	if md != nil && len(md) != 0 {
		fmt.Fprintf(f, "%smeta: \n", sp1)
		writeKeyValue(f, sp2, md)
	}
}

func writeKeyValue(f io.Writer, sp string, keyVal caes.Metadata) {

	for md_key, md_val := range keyVal {
		writeKeyValue1(f, sp, md_key, md_val)
	}
}

func writeKeyValue1(f io.Writer, sp string, md_key string, md_val interface{}) {

	switch md_val.(type) {
	case string:
		fmt.Fprintf(f, "%s%s: %s\n", sp, md_key, mkYamlString(md_val.(string)))
	case int, float32, float64, bool:
		fmt.Fprintf(f, "%s%s: %s\n", sp, md_key, md_val)
	case map[string]string:
		fmt.Fprintf(f, "%s%s: \n", sp, md_key)
		sp = sp + spPlus
		for key02, val02 := range md_val.(map[string]string) {
			fmt.Fprintf(f, "%s%s: %s\n", sp, key02, mkYamlString(val02))
		}
	case map[string]interface{}:
		fmt.Fprintf(f, "%s%s: \n", sp, md_key)
		for key02, val02 := range md_val.(map[string]interface{}) {
			writeKeyValue1(f, sp+spPlus, key02, val02)
		}
	default:
		fmt.Fprintf(f, "%s%s: %v\n", sp, md_key, md_val)
		// fmt.Fprintf(f, "--- Key: TYPE >> %s: %T << TYPE\n   VALU >> %s << VALU\n", md_key, md_val, md_val)
	}
}

func writeMapStrStr(f io.Writer, space1 string, name string, space2 string, mapstr map[string]string) {
	if mapstr != nil && len(mapstr) != 0 {
		fmt.Fprintf(f, "# %s%s:\n", space1, name)
		for key, val := range mapstr {
			fmt.Fprintf(f, "# %s%s: %s\n", space2, key, val)
		}
	}
}

func writeStrings(f io.Writer, space1 string, name string, space2 string, vars []string) {
	if vars != nil && len(vars) != 0 {
		fmt.Fprintf(f, "# %s%s:\n", space1, name)
		for _, v := range vars {
			fmt.Fprintf(f, "# %s- %s\n", space2, v)
		}
	}
}

func writeWF(f io.Writer, key, name string, weight caes.WeighingFunction) {
	val, found := collOfWF2source[name]
	if found {
		fmt.Fprintf(f, "# %s%s:\n", sp2, key)
		switch val.(type) {
		case constantWF:
			fmt.Fprintf(f, "# %sconstant: %3.2f\n", sp3, val.(constantWF).constant)
		case *caes.Criteria:
			c := val.(*caes.Criteria)
			fmt.Fprintf(f, "# %scriteria:\n", sp3)
			fmt.Fprintf(f, "# %shard: [", sp4)
			for ix, val := range c.HardConstraints {
				if ix != 0 {
					fmt.Fprintf(f, ", %s", val)
				} else {
					fmt.Fprintf(f, "%s", val)
				}
			}
			fmt.Fprintf(f, "]\n")
			fmt.Fprintf(f, "# %ssoft:\n", sp4)
			for role, sc := range c.SoftConstraints {
				fmt.Fprintf(f, "# %s%s:\n", sp5, role)
				fmt.Fprintf(f, "# %sfactor: %3.2f\n", sp6, sc.Factor)
				fmt.Fprintf(f, "# %svalues:\n", sp6)
				for nval, fl := range sc.NormalizedValues {
					fmt.Fprintf(f, "# %s%s: %3.2f\n", sp7, nval, fl)
				}
			}
		case []caes.PropertyOrder:
			povec := val.([]caes.PropertyOrder)
			fmt.Fprintf(f, "# %spreference:\n", sp3)
			sp4ms := sp4 + "  "
			for _, po := range povec {
				fmt.Fprintf(f, "# %s- property: %s\n", sp4, po.Property)
				vals := po.Values
				if vals != nil && len(vals) != 0 {
					fmt.Fprintf(f, "# %sorder: [", sp4ms)
					for ix, v := range vals {
						if ix == 0 {
							fmt.Fprintf(f, "%s", v)
						} else {
							fmt.Fprintf(f, ", %s", v)
						}
					}
					fmt.Fprintf(f, "]\n")
				} else {
					fmt.Fprintf(f, "# %sorder: ", sp4ms)
					if po.Order == caes.Ascending {
						fmt.Fprintf(f, "ascending\n")
					} else {
						fmt.Fprintf(f, "descending\n")
					}
				}
			}
		default:
			fmt.Fprintf(f, "# %s %v # defined in %s\n", sp3, weight, name)
		}
	} else {
		fmt.Fprintf(f, "# %s%s: %v\n", sp2, key, weight)
	}
}

func ExportWithReferences(f io.Writer, caesAg *caes.ArgGraph) {
	writeArgGraph1(false, f, caesAg)
}

func Export(f io.Writer, caesAg *caes.ArgGraph) {
	writeArgGraph1(true, f, caesAg)
}

func writeArgGraph1(noRefs bool, f io.Writer, caesAg *caes.ArgGraph) {
	sp0 = ""
	sp1 = spPlus
	sp2 = sp1 + spPlus
	sp3 = sp2 + spPlus
	sp4 = sp3 + spPlus
	sp5 = sp4 + spPlus
	sp6 = sp5 + spPlus
	sp7 = sp6 + spPlus

	writeMetaData(f, sp0, sp1, caesAg.Metadata)

	is := caesAg.Issues
	if is != nil {
		fmt.Fprintf(f, "issues: \n")
		for _, is_val := range is {
			fmt.Fprintf(f, "%s%s: \n", sp1, is_val.Id)
			writeMetaData(f, sp2, sp3, is_val.Metadata)
			fmt.Fprintf(f, "%spositions:\n", sp2)
			// first := true
			for _, ref_stat := range is_val.Positions {
				fmt.Fprintf(f, "%s- %s\n", sp3, ref_stat.Id)
			}
			/*
						if first == true {
							fmt.Fprintf(f, "[%s", ref_stat.Id)
							first = false
						} else {
							fmt.Fprintf(f, ",%s", ref_stat.Id)
						}
					}

				if first == true {
					fmt.Fprintf(f, "[]\n")
				} else {
					fmt.Fprintf(f, "]\n")
				}*/
			fmt.Fprintf(f, "        standard: ")
			s := "???"
			switch is_val.Standard {
			case 0:
				s = "PE"
			case 1:
				s = "CCE"
			case 2:
				s = "BRD"
			}
			fmt.Fprintf(f, "%s\n", s)
		}
	}

	std := caesAg.Statements
	if std != nil {
		fmt.Fprintf(f, "statements: \n")
		for _, ref_stat := range std {

			if ref_stat.Metadata == nil && ref_stat.Label == caes.Undecided && (noRefs == true || (ref_stat.Issue == nil && ref_stat.Args == nil)) {
				fmt.Fprintf(f, "%s%s: %s\n", sp1, ref_stat.Id, mkYamlString(ref_stat.Text))
			} else {

				fmt.Fprintf(f, "%s%s: \n", sp1, ref_stat.Id)
				writeMetaData(f, sp2, sp3, ref_stat.Metadata)

				if ref_stat.Text != "" {
					fmt.Fprintf(f, "%stext: %s \n", sp2, mkYamlString(ref_stat.Text))
				}
				if ref_stat.Label != caes.Undecided {
					fmt.Fprintf(f, "%slabel: %v\n", sp2, ref_stat.Label)
				}
				if noRefs == false && ref_stat.Issue != nil {
					fmt.Fprintf(f, "%sissue: %s \n", sp2, ref_stat.Issue.Id)
				}
				if noRefs == false && ref_stat.Args != nil {
					fmt.Fprintf(f, "%sarguments:\n", sp2)
					// first := true
					for _, arg := range ref_stat.Args {
						fmt.Fprintf(f, "%s- %s\n", sp3, arg.Id)
						/*if first == true {
							fmt.Fprintf(f, "[%s", arg.Id)
							first = false
						} else {
							fmt.Fprintf(f, ",%s", arg.Id)
						}*/
					}
					// fmt.Fprintf(f, "]\n")
				}
			}
		}
	}
	if caesAg.Assumptions != nil && len(caesAg.Assumptions) != 0 {
		fmt.Fprintf(f, "assumptions:\n")
		for _, stat := range caesAg.Assumptions {
			fmt.Fprintf(f, "%s- %s\n", sp1, stat)
		}
		/*		fmt.Fprintf(f, "assumptions: [")
				first := true
				for stat, boolval := range caesAg.Assumptions {
					if boolval {
						if first {
							first = false
						} else {
							fmt.Fprintf(f, ", ")
						}
						fmt.Fprintf(f, "%s", stat)
					}
				}
				fmt.Fprintf(f, "]\n")
				//	} else {
				//		fmt.Fprintf(f, " assumptins [%v, len, \"%v\"]\n", caesAg.Assumptions, len(caesAg.Assumptions))
		*/
	}
	if caesAg.Arguments != nil {
		fmt.Fprintf(f, "arguments: \n")
		for _, ref_caesAg_Arg := range caesAg.Arguments {
			fmt.Fprintf(f, "%s%s:\n", sp1, ref_caesAg_Arg.Id)

			writeMetaData(f, sp2, sp3, ref_caesAg_Arg.Metadata)

			first := true
			list := true
			for _, prem := range ref_caesAg_Arg.Premises {
				if prem.Role != "" {
					if list == true && first == true {
						fmt.Fprintf(f, "%spremises:\n", sp2)
					}
					s := "nil?"
					if prem.Stmt != nil {
						s = prem.Stmt.Id
					}
					if first == false && list == true {
						fmt.Fprintf(f, "\n")
					}
					fmt.Fprintf(f, "%s%s: %s\n", sp3, prem.Role, s)
					list = false
				} else {
					s := "nil?"
					if prem.Stmt != nil {
						s = prem.Stmt.Id
					}
					if first == true {
						// fmt.Fprintf(f, "%spremises: [%s", sp2, s)
						fmt.Fprintf(f, "%spremises:\n", sp2)
						fmt.Fprintf(f, "%s- %s\n", sp3, s)
						first = false
					} else {
						// fmt.Fprintf(f, ",%s", s)
						fmt.Fprintf(f, "%s- %s\n", sp3, s)
					}
				}
			}
			/* if first == false { // && list == true
				fmt.Fprintf(f, "]\n")
			}*/
			if ref_caesAg_Arg.Conclusion != nil {
				fmt.Fprintf(f, "%sconclusion: %s\n", sp2, ref_caesAg_Arg.Conclusion.Id)
			}
			if ref_caesAg_Arg.Weight != 0.0 {
				fmt.Fprintf(f, "%sweight: %4.2f\n", sp2, ref_caesAg_Arg.Weight)
			}
			if ref_caesAg_Arg.Scheme != nil {
				fmt.Fprintf(f, "%sscheme: %s\n", sp2, ref_caesAg_Arg.Scheme.Id)
			}
			if ref_caesAg_Arg.Parameters != nil && len(ref_caesAg_Arg.Parameters) != 0 {
				fmt.Fprintf(f, "%sparameters: ", sp2)
				for idx, para := range ref_caesAg_Arg.Parameters {
					if idx == 0 {
						fmt.Fprintf(f, "[%s", para)
					} else {
						fmt.Fprintf(f, ", %s", para)
					}
				}
				fmt.Fprintf(f, "]\n")
			}
			if ref_caesAg_Arg.Undercutter != nil {
				fmt.Fprintf(f, "%sundercutter: %s\n", sp2, ref_caesAg_Arg.Undercutter.Id)
			}
		}
	}
	first := true
	for key, md := range caesAg.References {
		if first == true {
			fmt.Fprintf(f, "references:\n")
			first = false
		}
		fmt.Fprintf(f, "%s%s:\n", sp2, key)
		for md_key, md_val := range md {
			writeKeyValue1(f, sp3, md_key, md_val)
		}
	}
	if caesAg.ExpectedLabeling != nil && len(caesAg.ExpectedLabeling) != 0 {
		fmt.Fprintf(f, "labels:\n")
		in := []string{}
		out := []string{}
		undec := []string{}
		for stat, lbl := range caesAg.ExpectedLabeling {
			switch lbl {
			case caes.Undecided:
				undec = append(undec, stat)
			case caes.In:
				in = append(in, stat)
			case caes.Out:
				out = append(out, stat)
			}

		}
		if len(in) != 0 {
			fmt.Fprintf(f, "%sin: ", sp2)
			first := true
			for _, stat := range in {
				if !first {
					fmt.Fprintf(f, ", %s", mkYamlString(stat))
				} else {
					fmt.Fprintf(f, "[%s", mkYamlString(stat))
					first = false
				}
			}
			fmt.Fprintf(f, "]\n")
		}

		if len(out) != 0 {
			fmt.Fprintf(f, "%sout: ", sp2)
			first := true
			for _, stat := range out {
				if !first {
					fmt.Fprintf(f, ", %s", mkYamlString(stat))
				} else {
					fmt.Fprintf(f, "[%s", mkYamlString(stat))
					first = false
				}
			}
			fmt.Fprintf(f, "]\n")
		}

		if len(undec) != 0 {
			fmt.Fprintf(f, "%sundecided: ", sp2)
			first := true
			for _, stat := range undec {
				if !first {
					fmt.Fprintf(f, ", %s", mkYamlString(stat))
				} else {
					fmt.Fprintf(f, "[%s", mkYamlString(stat))
					first = false
				}
			}
			fmt.Fprintf(f, "]\n")
		}

	}

	//	// Write out the theory, as a YAML comment, for debugging
	//	t := caesAg.Theory
	//	if t == nil {
	//		fmt.Fprintf(f, "# !!! No Theory !!!\n")
	//		return
	//	}
	//	fmt.Fprintf(f, "#  -----------------------------------------------\n")
	//	fmt.Fprintf(f, "# Theory\n")
	//	fmt.Fprintf(f, "#  -----------------------------------------------\n")
	//	if t.Language != nil && len(t.Language) != 0 {
	//		fmt.Fprintf(f, "# language:\n")
	//		for key, val := range t.Language {
	//			fmt.Fprintf(f, "# %s%s: %s\n", sp1, key, val)
	//		}
	//	}
	//	wf := t.WeighingFunctions
	//	if wf != nil && len(wf) != 0 {
	//		fmt.Fprintf(f, "# weighing_functions:\n")
	//		for key, weight := range wf {
	//			writeWF(f, key, key, weight)
	//			// fmt.Fprintf(f, "# %s- %s\n", sp1, key)
	//		}
	//	}
	//	as := t.ArgSchemes
	//	if as != nil && len(as) != 0 {
	//		fmt.Fprintf(f, "# argument_schemes:\n")
	//		for name, scheme := range as {
	//			if name != scheme.Id {
	//				fmt.Fprintf(f, "# ERROR name '%s' ist not scheme-id '%s'\n", name, scheme.Id)
	//			} else {
	//				fmt.Fprintf(f, "# %s%s:\n", sp1, name)
	//			}
	//			meta := scheme.Metadata
	//			if meta != nil && len(meta) != 0 {
	//				fmt.Fprintf(f, "# %smeta:\n", sp2)
	//				for mkey, mval := range meta {
	//					fmt.Fprintf(f, "# %s%s: %v\n", sp3, mkey, mval)
	//				}
	//			}

	//			writeStrings(f, sp2, "variables", sp3, scheme.Variables)

	//			weight := scheme.Weight
	//			if weight != nil {
	//				writeWF(f, "weight", name, weight)

	//			} // if weigth != nil
	//			writeMapStrStr(f, sp2, "premises", sp3, scheme.Premises)
	//			writeMapStrStr(f, sp2, "assumptions", sp3, scheme.Assumptions)
	//			writeMapStrStr(f, sp2, "exceptions", sp3, scheme.Exceptions)
	//			writeStrings(f, sp2, "deletions", sp3, scheme.Deletions)
	//			writeStrings(f, sp2, "guards", sp3, scheme.Guards)
	//			writeStrings(f, sp2, "conclusions", sp3, scheme.Conclusions)

	//		}
	//	}

}
