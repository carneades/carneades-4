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
	"github.com/carneades/carneades-4/src/engine/caes"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	// "log"
	"strconv"
	"strings"
)

type (
	mapIface map[interface{}]interface{}
	umIssue  struct {
		id        string
		metadata  caes.Metadata
		positions []string
		standard  caes.Standard
	}
	umPremis struct {
		stmt string
		role string
	}
	umArgument struct {
		id          string
		metadata    caes.Metadata
		premises    []umPremis
		conclusion  string
		weigth      float64
		scheme      string
		undercutter string
	}
	assumptions   []string
	mapIssues     map[string]umIssue
	mapStatements map[string]*caes.Statement
	mapArguments  map[string]umArgument
	mapReferences map[string]caes.Metadata
	mapLabels     map[string]caes.Label
	argMapGraph   struct {
		Metadata   caes.Metadata
		Issues     mapIssues
		Statements mapStatements
		Arguments  mapArguments
		References mapReferences
		Assumtions assumptions
		Labels     mapLabels
	}
)

const spPlus = "    "

func Import(inFile io.Reader) (*caes.ArgGraph, error) {

	data, err := ioutil.ReadAll(inFile)
	if err != nil {
		return nil, err
	}
	// log.Printf("Read-Datei: \nErr: %v len(data): %v \n", err, len(data))

	m1 := make(mapIface)
	err = yaml.Unmarshal(data, &m1)
	if err != nil {
		return nil, err
	}

	return iface2caes(m1)
}

func iface2caes(m mapIface) (caesAg *caes.ArgGraph, err error) {
	var yamlArgMapGraph argMapGraph
	yamlArgMapGraph.Metadata = make(caes.Metadata)
	yamlMetaData := yamlArgMapGraph.Metadata
	yamlArgMapGraph.Issues = make(mapIssues)
	yamlIssues := yamlArgMapGraph.Issues
	yamlArgMapGraph.Statements = make(map[string]*caes.Statement)
	yamlStats := yamlArgMapGraph.Statements
	yamlArgMapGraph.Arguments = make(mapArguments)
	yamlArgs := yamlArgMapGraph.Arguments
	yamlArgMapGraph.References = make(map[string]caes.Metadata)
	yamlRefs := yamlArgMapGraph.References
	var yamlAssumps assumptions
	yamlArgMapGraph.Assumtions = yamlAssumps
	yamlArgMapGraph.Labels = make(mapLabels)
	yamlLbls := yamlArgMapGraph.Labels
	// caes.ArgGraph
	caesAg = &caes.ArgGraph{}

	for key, value := range m {
		keyStr := strings.ToLower(key.(string))
		switch keyStr {
		case "statements":
			yamlStats, err = iface2statement(value, yamlStats)
			if err != nil {
				return caesAg, err
			}
		case "assumptions":
			yamlAssumps, err = iface2assumps(value, yamlAssumps)
			if err != nil {
				return caesAg, err
			}
		case "issues":
			yamlIssues, err = iface2issues(value, yamlIssues)
			if err != nil {
				return caesAg, err
			}
		case "arguments":
			yamlArgs, err = iface2arguments(value, yamlArgs)
			if err != nil {
				return caesAg, err
			}
		case "premise":
			// log.Printf("Premise: \n")
		case "meta", "metadata":
			// log.Printf("Meta: \n")
			yamlMetaData, err = iface2metadata(value, yamlMetaData)
			if err != nil {
				return caesAg, err
			}
		case "references":
			yamlRefs, err = iface2references(value, yamlRefs)
			if err != nil {
				return caesAg, err
			}
		case "labels":
			yamlLbls, err = iface2labels(value, yamlLbls)
		default:
			// log.Printf("Default: \n")
		}
	}

	// create ArgGraph
	// ===============
	// Metadata
	// --------
	caesAg.Metadata = yamlMetaData
	// log.Printf("   ---  Metadata --- \n %v \n ------End Metadata --- \n", caesAg.Metadata)
	// Statement
	// ---------
	first := true
	found := false
	for _, refYamlStat := range yamlStats {

		if first {
			caesAg.Statements = []*caes.Statement{refYamlStat}
			first = false
		} else {
			caesAg.Statements = append(caesAg.Statements, refYamlStat)
		}
	}
	// log.Printf("   ---  Statements --- \n %v \n ------End Statements --- \n", caesAg.Statements)
	// Issue
	first = true
	for yamlIssue_Id, yamlIssue_Val := range yamlIssues {
		caes_Issue := &caes.Issue{Id: yamlIssue_Id, Metadata: yamlIssue_Val.metadata, Standard: yamlIssue_Val.standard}
		if first {
			caesAg.Issues = []*caes.Issue{caes_Issue}
			first = false
		} else {
			caesAg.Issues = append(caesAg.Issues, caes_Issue)
		}
		// References: Issue.Positions --> []*Statement, Statement.Issue --> *Issue
		for _, yamlIssue_Pos := range yamlIssue_Val.positions {
			found = false
		LoopIss:
			for _, caesAg_Stat := range caesAg.Statements {
				if yamlIssue_Pos == caesAg_Stat.Id {
					found = true
					// log.Printf("   Position: %s \n", yamlIssue_Pos)
					if caes_Issue.Positions == nil {
						caes_Issue.Positions = []*caes.Statement{caesAg_Stat}
					} else {
						caes_Issue.Positions = append(caes_Issue.Positions, caesAg_Stat)
					}
					if caesAg_Stat.Issue == nil {
						caesAg_Stat.Issue = caes_Issue
					} else {
						return caesAg, errors.New(" *** Semantic Error: Statement: " + caesAg_Stat.Id + ", with two issues: " + caes_Issue.Id + ", " + caesAg_Stat.Issue.Id + "\n")
					}
					break LoopIss
				}
			}
			if !found {
				return caesAg, errors.New(" *** Semantic Error: Position " + yamlIssue_Pos + ", from Issue: " + caes_Issue.Id + ", is not a Statement-ID\n")
			}
		}
	}

	// log.Printf("   ---  Issues --- \n %v \n ------End Issuess --- \n", caesAg.Issues)

	// Arguments
	first = true
	for yamlArg_Id, yamlArg_Val := range yamlArgs {
		caesArg := &caes.Argument{Id: yamlArg_Id, Metadata: yamlArg_Val.metadata, Weight: yamlArg_Val.weigth}
		if first {
			caesAg.Arguments = []*caes.Argument{caesArg}
			first = false
		} else {
			caesAg.Arguments = append(caesAg.Arguments, caesArg)
		}
		// References: Argument.Conclusion --> *Statement, Statement.Args --> []*Argument
		found := false
	LoopC:
		for _, caesArg_Stat := range caesAg.Statements {
			if yamlArg_Val.conclusion == caesArg_Stat.Id {
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
			return caesAg, errors.New(" *** Semantic Error: Conclusion: " + yamlArg_Val.conclusion + ", from Argument: " + yamlArg_Id + ", is not a Statement-ID\n")
		}

		// References: Argument.undercutter --> *Statement,
		// No undercutter in Statement.Args --> []*Argument
		if yamlArg_Val.undercutter != "" {
			found = false
		LoopN:
			for _, caesArg_Stat := range caesAg.Statements {
				if yamlArg_Val.undercutter == caesArg_Stat.Id {
					found = true
					caesArg.Undercutter = caesArg_Stat
					break LoopN
				}
			}
			if !found {
				return caesAg, errors.New(" *** Semantic Error: Undercutter: " + yamlArg_Val.undercutter + ", from Argument: " + yamlArg_Id + ", is not a Statement-ID\n")
			}
		}
		// Argument.Premises
		for _, yamlArg_Prem := range yamlArg_Val.premises {
			prem_stat, ok := yamlStats[yamlArg_Prem.stmt]
			if !ok {
				return caesAg, errors.New(" *** Semantic Error: Premise: " + yamlArg_Prem.stmt + ", from Argument: " + yamlArg_Id + ", is not a Statement-ID\n")
			}
			if prem_stat == nil {
				// log.Printf("\n *** Prem Stat == nil für %s \n", yamlArg_Prem.stmt)
			} else {
				// log.Printf(" \n +++ Prem_Stat: %s für %s \n", prem_stat.Id, yamlArg_Prem.stmt)
			}
			caes_prem := caes.Premise{Stmt: prem_stat, Role: yamlArg_Prem.role}
			if caesArg.Premises == nil {
				caesArg.Premises = []caes.Premise{caes_prem}
			} else {
				caesArg.Premises = append(caesArg.Premises, caes_prem)
			}
		}
		// Scheme
		if yamlArg_Val.scheme != "" {
			caesArg.Scheme = yamlArg_Val.scheme
		}
	}
	// log.Printf("   ---  Arguments --- \n %v \n ------End Arguments --- \n", caesAg.Arguments)
	// Assumptions
	for _, yamlAss := range yamlAssumps {
		found = false
		for _, caesArg_Stat := range caesAg.Statements {
			if yamlAss == caesArg_Stat.Id {
				found = true
				// log.Printf(" Set assumtions: %s\n", yamlAss)
				caesArg_Stat.Assumed = true
			}
		}
		if !found {
			return caesAg, errors.New(" *** Semantic Error: Assumption: " + yamlAss + ", is not a Statement-ID\n")
		}
	}

	// References
	caesAg.References = yamlRefs
	// log.Printf("   ---  References --- \n %v \n ------End References --- \n", caesAg.References)
	// Labels
	// if yamlLbls not empty
	for _, caesArg_Stat := range caesAg.Statements {
		lbl, found := yamlLbls[caesArg_Stat.Id]
		if found == true {
			// log.Printf(" Label %s:%v\n", caesArg_Stat.Id, lbl)
			caesArg_Stat.Label = lbl
		}
	}
	//check
	for lbl_Id, lbl_val := range yamlLbls {
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

	return caesAg, nil

}

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

func ExportWithReferences(f io.Writer, caesAg *caes.ArgGraph) {
	writeArgGraph1(false, f, caesAg)
}

func Export(f io.Writer, caesAg *caes.ArgGraph) {
	writeArgGraph1(true, f, caesAg)
}

func writeArgGraph1(noRefs bool, f io.Writer, caesAg *caes.ArgGraph) {
	sp0 := ""
	sp1 := spPlus
	sp2 := sp1 + spPlus
	sp3 := sp2 + spPlus

	writeMetaData(f, sp0, sp1, caesAg.Metadata)

	is := caesAg.Issues
	if is != nil {
		fmt.Fprintf(f, "issues: \n")
		for _, is_val := range is {
			fmt.Fprintf(f, "%s%s: \n", sp1, is_val.Id)
			writeMetaData(f, sp2, sp3, is_val.Metadata)
			fmt.Fprintf(f, "%spositions: ", sp2)
			first := true
			for _, ref_stat := range is_val.Positions {
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
			}
			fmt.Fprintf(f, "        standard: ")
			s := "???"
			switch is_val.Standard {
			case 0:
				s = "DV"
			case 1:
				s = "PE"
			case 2:
				s = "CCE"
			case 3:
				s = "BRD"
			}
			fmt.Fprintf(f, "%s\n", s)
		}
	}

	std := caesAg.Statements
	if std != nil {
		fmt.Fprintf(f, "statements: \n")
		for _, ref_stat := range std {

			if ref_stat.Metadata == nil && ref_stat.Assumed == false && ref_stat.Label == caes.Undecided && (noRefs == true || (ref_stat.Issue == nil && ref_stat.Args == nil)) {
				fmt.Fprintf(f, "%s%s: %s\n", sp1, ref_stat.Id, mkYamlString(ref_stat.Text))
			} else {

				fmt.Fprintf(f, "%s%s: \n", sp1, ref_stat.Id)
				writeMetaData(f, sp2, sp3, ref_stat.Metadata)

				if ref_stat.Text != "" {
					fmt.Fprintf(f, "%stext: %s \n", sp2, mkYamlString(ref_stat.Text))
				}
				if ref_stat.Assumed == true {
					fmt.Fprintf(f, "%sassumed: true\n", sp2)
				}
				if ref_stat.Label != caes.Undecided {
					fmt.Fprintf(f, "%slabel: %v\n", sp2, ref_stat.Label)
				}
				if noRefs == false && ref_stat.Issue != nil {
					fmt.Fprintf(f, "%sissue: %s \n", sp2, ref_stat.Issue.Id)
				}
				if noRefs == false && ref_stat.Args != nil {
					fmt.Fprintf(f, "%sarguments: ", sp2)
					first := true
					for _, arg := range ref_stat.Args {
						if first == true {
							fmt.Fprintf(f, "[%s", arg.Id)
							first = false
						} else {
							fmt.Fprintf(f, ",%s", arg.Id)
						}
					}
					fmt.Fprintf(f, "]\n")
				}
			}
		}
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
						fmt.Fprintf(f, "%spremises: [%s", sp2, s)
						first = false
					} else {
						fmt.Fprintf(f, ",%s", s)
					}
				}
			}
			if first == false { // && list == true
				fmt.Fprintf(f, "]\n")
			}
			if ref_caesAg_Arg.Conclusion != nil {
				fmt.Fprintf(f, "%sconclusion: %s\n", sp2, ref_caesAg_Arg.Conclusion.Id)
			}
			if ref_caesAg_Arg.Weight != 0.0 {
				fmt.Fprintf(f, "%sweight: %4.2f\n", sp2, ref_caesAg_Arg.Weight)
			}
			if ref_caesAg_Arg.Scheme != "" {
				fmt.Fprintf(f, "%sscheme: %s\n", sp2, ref_caesAg_Arg.Scheme)
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

func iface2labels(value interface{}, yamlLbls mapLabels) (mapLabels, error) {
	var err error
	switch subT := value.(type) {
	case mapIface:
		for lblkey, lblvalue := range value.(mapIface) {
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
		return yamlLbls, errors.New("*** Error labels: (Type)" + subT.(string) + "\n")

	}
	return yamlLbls, err
}

func iface2lbl_1(inArg interface{}, yamlLbls mapLabels, label caes.Label) (mapLabels, error) {
	// log.Printf("labels:\n   %v: ", label)
	switch intype := inArg.(type) {
	case []interface{}:
		// log.Printf(" [")
		for idx, stat := range inArg.([]interface{}) {
			switch stype := stat.(type) {
			case string:
				yamlLbls[stat.(string)] = label
				if idx != 0 {
					// log.Printf(", ")
				}
				// log.Printf("%s", stat.(string))
			default:
				return yamlLbls, errors.New("*** Error labels [(Type)]:" + stype.(string) + "\n")
			}
		}
		// log.Printf("]")
	case string:
		yamlLbls[inArg.(string)] = label
	default:
		return yamlLbls, errors.New("*** Error labels (Type):" + intype.(string) + "\n")

	}
	return yamlLbls, nil
}

func iface2metadata(value interface{}, meta caes.Metadata) (caes.Metadata, error) {
	switch subT := value.(type) {
	case mapIface:
		for metakey, metavalue := range value.(mapIface) {
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
			case mapIface:
				var err error
				meta[metakey.(string)], err = iface2metadata(metavalue, caes.Metadata{})
				return meta, err
			}
		}
	default:
		return meta, errors.New("*** Error metadata (Type):" + subT.(string) + "\n")
	}
	return meta, nil
}

func iface2statement(value interface{}, yamlStats mapStatements) (mapStatements, error) {

	var err error
	// log.Printf("Statements: \n")
	// switch t := value.(type) {
	switch value.(type) {
	case mapIface:
		for st_key, st_value := range value.(mapIface) {
			keyStr := st_key.(string)
			switch st_value.(type) {
			case string:
				yamlStats[keyStr] = &caes.Statement{
					Id:    keyStr,
					Text:  st_value.(string),
					Label: caes.Undecided,
				}
				// log.Printf(" %v: %v \n", st_key.(string), st_value.(string))
			case int, float32, float64, bool:
				yamlStats[keyStr] = &caes.Statement{
					Id:    keyStr,
					Text:  iface2string(st_value),
					Label: caes.Undecided,
				}
				// log.Printf(" %v: %v \n", st_key.(string), iface2string(st_value))
			case mapIface:
				// log.Printf("   %v:\n", st_key.(string))
				yamlStats[keyStr], err = iface2xstatement(st_value, &caes.Statement{Id: st_key.(string), Label: caes.Undecided})
				if err != nil {
					return yamlStats, err
				}
			}
		}
	default:
		// log.Printf(" Type: %v \n", t)
	}
	return yamlStats, nil
}

func iface2xstatement(st_value interface{}, stat *caes.Statement) (*caes.Statement, error) {
	var err error
	for st_subkey, st_subvalue := range st_value.(mapIface) {
		st_subkey_lo := strings.ToLower(st_subkey.(string))
		switch st_subkey_lo {
		case "meta", "metadata":
			// log.Printf("      meta:\n")
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
			// log.Printf("      text: %s \n", stat.Text)
		case "assumed":
			switch t := st_subvalue.(type) {
			case bool:
				stat.Assumed = st_subvalue.(bool)
				// log.Printf("     assumed: %v \n", stat.Assumed)
			case int:
				if st_subvalue.(int) == 0 {
					stat.Assumed = false
				} else {
					stat.Assumed = true
				}
				// log.Printf("     assumed: %v \n", stat.Assumed)
			default:
				return stat, errors.New("*** ERROR: Assumed value not bool: " + t.(string) + " \n")
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

func iface2issues(value interface{}, yamlIssues mapIssues) (mapIssues, error) {
	// log.Printf("issues:\n")
	switch value.(type) {
	case mapIface:
		for issueName, issueMap := range value.(mapIface) {
			switch iT := issueName.(type) {
			case string:
				issueNameStr := issueName.(string)
				// log.Printf("   %s:\n", issueNameStr)
				issue := umIssue{id: issueNameStr, standard: caes.PE}
				switch issueMap.(type) {
				case mapIface:
					for issueKey, issueValue := range issueMap.(mapIface) {
						switch issueKey.(type) {
						case string:
							// log.Printf("      %s: ", issueKey.(string))
							issueKey_lo := strings.ToLower(issueKey.(string))
							switch issueKey_lo {
							case "positions":
								// switch ivT := issueValue.(type) {
								switch issueValue.(type) {
								case []interface{}:
									issueArr := issueValue.([]interface{})
									issue.positions = make([]string, len(issueArr))
									// log.Printf(" [")
									for i, ele_i := range issueArr {
										if i != 0 {
											// log.Printf(" ,")
										}
										switch eleT := ele_i.(type) {
										case string:
											issue.positions[i] = ele_i.(string)
											// log.Printf("%v", issue.positions[i])
										default:
											return yamlIssues,
												errors.New("*** Error Position value-type: " + eleT.(string) + " \n")
										}
									}
									// log.Printf("]\n")
								case string: /* a singel position */
									issue.positions = make([]string, 1)
									issue.positions[0] = issueValue.(string)
									// log.Printf(" = %s \n", issueValue.(string))
								default:
									// log.Printf("      position-type: %v \n", ivT)
								}

							case "standard":
								switch ivT := issueValue.(type) {
								case string:
									issueValueStr := issueValue.(string)
									// log.Printf("%s\n", issueValueStr)
									switch issueValueStr {
									case "DV", "dv":
										issue.standard = caes.DV
									case "PE", "pe":
										issue.standard = caes.PE
									case "CCE", "cce":
										issue.standard = caes.CCE
									case "BRD", "brd":
										issue.standard = caes.BRD
									default:
										return yamlIssues,
											errors.New("*** Error: position expected DV,PE, CCE, BRD, wrong: " + issueValue.(string) + " \n")
									}
								default:
									return yamlIssues,
										errors.New("*** Error: unexpected issue-value-type:" + ivT.(string) + "\n")
								}

							case "meta", "metadata":
								var err error
								if issue.metadata == nil {
									issue.metadata, err = iface2metadata(issueValue, caes.Metadata{})
								} else {
									issue.metadata, err = iface2metadata(issueValue, issue.metadata)
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
					errors.New("*** ERROR: Not a issue Name: " + iT.(string) + "\n")
			}
		}
	}
	return yamlIssues, nil
}

func iface2assumps(value interface{}, yamlAssumps assumptions) (assumptions, error) {
	// log.Printf("Assumptions: ")
	switch value.(type) {
	case []interface{}:
		// log.Printf("[")
		for i, str := range value.([]interface{}) {
			if i != 0 {
				// log.Printf(", ")
			}
			switch str.(type) {
			case string:
				if yamlAssumps == nil {
					yamlAssumps = assumptions{str.(string)}
				} else {
					yamlAssumps = append(yamlAssumps, str.(string))
				}
				// log.Printf("%s", str.(string))
			default:
				if yamlAssumps == nil {
					yamlAssumps = assumptions{iface2string(str)}
				} else {
					yamlAssumps = append(yamlAssumps, iface2string(str))
				}
				// log.Printf("if=%v", iface2string(str))
			}
		}
		// log.Printf("]\n")
	case string:
		if yamlAssumps == nil {
			yamlAssumps = assumptions{value.(string)}
		} else {
			yamlAssumps = append(yamlAssumps, value.(string))
		}
		// log.Printf(" %s \n", value.(string))
	default:
		return yamlAssumps,
			errors.New("*** ERROR: not a right assumption-value: " + value.(string) + "\n")
	}
	return yamlAssumps, nil
}

func iface2arguments(value interface{}, yamlArgs mapArguments) (mapArguments, error) {
	// log.Printf("Arguments: \n")
	switch value.(type) {
	case mapIface:
		for argName, arg := range value.(mapIface) {
			switch aT := argName.(type) {
			case string:
				// log.Printf("    %s: \n", argName.(string))
				var err error
				yamlArgs[argName.(string)], err = iface2argument(arg, umArgument{})
				if err != nil {
					return yamlArgs, err
				}
			default:
				return yamlArgs,
					errors.New("*** ERROR: Arguement-name is not a string: " + aT.(string) + "\n")
			}
		}
	}
	return yamlArgs, nil
}

func iface2argument(inArg interface{}, outArg umArgument) (umArgument, error) {
	var err error
	switch inArg.(type) {
	case mapIface:
		for attName, attValue := range inArg.(mapIface) {
			switch ant := attName.(type) {
			case string:
				attNameStr := strings.ToLower(attName.(string))
				// log.Printf("      %s:", attNameStr)
				switch attNameStr {
				case "conclusion":
					outArg.conclusion = iface2string(attValue)
					// log.Printf(" %s \n", outArg.conclusion)
				case "premises":
					outArg.premises, err = iface2premises(attValue)
					if err != nil {
						return outArg, err
					}
					// log.Printf("\n")
				case "metadata", "meta":
					outArg.metadata, err = iface2metadata(attValue, caes.Metadata{})
					if err != nil {
						return outArg, err
					}
				case "weight":
					outArg.weigth, err = iface2weight(attValue)
					if err != nil {
						return outArg, err
					}
				case "nas", "undercutter", "not app statement":
					outArg.undercutter = iface2string(attValue)
					// log.Printf(" %s \n", outArg.undercutter)
				case "scheme":
					outArg.scheme = iface2string(attValue)
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
					//					// log.Printf(" %s \n", outArg.scheme)
				default:
					return outArg,
						errors.New("*** ERROR: Wrong argument attribute: " + attName.(string) + " (expected: conclusion, premises, weight, undercutter, scheme or metadata)\n")
				}
			default:
				return outArg,
					errors.New("*** ERROR: Wrong argument attribute type (string expected): " + ant.(string) + "\n")
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
			return 0.0, errors.New(("*** ERROR: Wrong weigth type (float expeced) integer value: " + attValue.(string)))
		}
	default:
		return 0.0, errors.New("*** ERROR: Wrong weigth type (float expected): " + atype.(string) + "\n")
	}
	// log.Printf("     weight: %v\n", *weight)
	return weight, nil
}

func iface2premises(inArg interface{}) ([]umPremis, error) {
	var outArg []umPremis
	switch inArg.(type) {
	case []interface{}:
		// log.Printf(" [")
		for idx, stat := range inArg.([]interface{}) {
			switch stat.(type) {
			case string:
				umP := umPremis{stmt: stat.(string)}
				if outArg == nil {
					outArg = []umPremis{umP}
				} else {
					outArg = append(outArg, umP)
				}
				if idx != 0 {
					// log.Printf(", ")
				}
				// log.Printf("%s", stat.(string))

			}
		}
		// log.Printf("]")
	case mapIface:
		// log.Printf("\n")
		for key, val := range inArg.(mapIface) {
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
			premis.stmt = valStr
			if outArg == nil {
				outArg = []umPremis{premis}
			} else {
				outArg = append(outArg, premis)
			}
		}
	}
	return outArg, nil
}

func iface2references(reference interface{}, yamlRefs mapReferences) (mapReferences, error) {
	var err error
	// log.Printf("references: \n")
	switch reference.(type) {
	case mapIface:
		for refName, refBody := range reference.(mapIface) {
			refNameStr := "???"
			switch refName.(type) {
			case string:
				refNameStr = refName.(string)
			default:
				refNameStr = iface2string(refName)
			}
			// log.Printf("    %s:\n", refNameStr)
			yamlRefs[refNameStr], err = iface2metadata(refBody, caes.Metadata{})
			if err != nil {
				return yamlRefs, err
			}
		}
	}
	return yamlRefs, nil
}
