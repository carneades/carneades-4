// Copyright © 2015 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

// func Import(in io.Reader) (*caes.ArgGraph, error)
// func Export(out io.Writer, ag *caes.ArgGraph)
// func ExportWithReferences(out io.Writer, ag *caes.ArgGraph)

package yaml

import (
	"errors"
	"fmt"
	"github.com/carneades/carneades-4/internal/engine/caes"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	//	"log"
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

func iface2caes(m mapIface) (ag *caes.ArgGraph, err error) {
	var argMapGraph argMapGraph
	argMapGraph.Metadata = make(caes.Metadata)
	mData := argMapGraph.Metadata
	argMapGraph.Issues = make(mapIssues)
	issues := argMapGraph.Issues
	argMapGraph.Statements = make(map[string]*caes.Statement)
	stats := argMapGraph.Statements
	argMapGraph.Arguments = make(mapArguments)
	args := argMapGraph.Arguments
	argMapGraph.References = make(map[string]caes.Metadata)
	refs := argMapGraph.References
	var assumps assumptions
	argMapGraph.Assumtions = assumps
	argMapGraph.Labels = make(mapLabels)
	lbls := argMapGraph.Labels
	ag = &caes.ArgGraph{}

	for key, value := range m {
		keyStr := strings.ToLower(key.(string))
		switch keyStr {
		case "statements":
			stats, err = iface2statement(value, stats)
			if err != nil {
				return ag, err
			}
		case "assumptions":
			assumps, err = iface2assumps(value, assumps)
			if err != nil {
				return ag, err
			}
		case "issues":
			issues, err = iface2issues(value, issues)
			if err != nil {
				return ag, err
			}
		case "arguments":
			args, err = iface2arguments(value, args)
			if err != nil {
				return ag, err
			}
		case "premise":
			// log.Printf("Premise: \n")
		case "meta", "metadata":
			// log.Printf("Meta: \n")
			mData, err = iface2metadata(value, mData)
			if err != nil {
				return ag, err
			}
		case "references":
			refs, err = iface2references(value, refs)
			if err != nil {
				return ag, err
			}
		case "labels":
			lbls, err = iface2labels(value, lbls)
		default:
			// log.Printf("Default: \n")
		}
	}

	// create ArgGraph
	// ---------------
	// Metadata
	ag.Metadata = mData
	// log.Printf("   ---  Metadata --- \n %v \n ------End Metadata --- \n", ag.Metadata)
	// Statement
	first := true
	found := false
	for _, ref_stat := range stats {

		if first {
			ag.Statements = []*caes.Statement{ref_stat}
			first = false
		} else {
			ag.Statements = append(ag.Statements, ref_stat)
		}
	}
	// log.Printf("   ---  Statements --- \n %v \n ------End Statements --- \n", ag.Statements)
	// Issue
	first = true
	for issue_id, issue_val := range issues {
		iss := &caes.Issue{Id: issue_id, Metadata: issue_val.metadata, Standard: issue_val.standard}
		if first {
			ag.Issues = []*caes.Issue{iss}
			first = false
		} else {
			ag.Issues = append(ag.Issues, iss)
		}
		// References: Issue.Positions --> []*Statement, Statement.Issue --> *Issue
		for _, pos := range issue_val.positions {
			found = false
		LoopIss:
			for _, stat := range ag.Statements {
				if pos == stat.Id {
					found = true
					// log.Printf("   Position: %s \n", pos)
					if iss.Positions == nil {
						iss.Positions = []*caes.Statement{stat}
					} else {
						iss.Positions = append(iss.Positions, stat)
					}
					if stat.Issue == nil {
						stat.Issue = iss
					} else {
						return ag, errors.New(" *** Semantic Error: Statement: " + stat.Id + ", with two issues: " + iss.Id + ", " + stat.Issue.Id + "\n")
					}
					break LoopIss
				}
			}
			if !found {
				return ag, errors.New(" *** Semantic Error: Position " + pos + ", from Issue: " + iss.Id + ", is not a Statement-ID\n")
			}
		}
	}
	/* // No issues
	     	if first == true {
	   		iss := &caes.Issue{Id: "defoult_issue", Standard: caes.PE}
	   		ag.Issues = []*caes.Issue{iss}
	   	}
	*/
	// log.Printf("   ---  Issues --- \n %v \n ------End Issuess --- \n", ag.Issues)

	// Arguments
	first = true
	for arg_id, arg_val := range args {
		arg := &caes.Argument{Id: arg_id, Metadata: arg_val.metadata, Weight: arg_val.weigth}
		if first {
			ag.Arguments = []*caes.Argument{arg}
			first = false
		} else {
			ag.Arguments = append(ag.Arguments, arg)
		}
		// References: Argument.Conclusion --> *Statement, Statement.Args --> []*Argument
		found := false
	LoopC:
		for _, stat := range ag.Statements {
			if arg_val.conclusion == stat.Id {
				arg.Conclusion = stat
				found = true
				if stat.Args == nil {
					stat.Args = []*caes.Argument{arg}
				} else {
					stat.Args = append(stat.Args, arg)
				}
				break LoopC
			}
		}
		if !found {
			return ag, errors.New(" *** Semantic Error: Conclusion: " + arg_val.conclusion + ", from Argument: " + arg_id + ", is not a Statement-ID\n")
		}

		// References: Argument.undercutter --> *Statement, Statement.Args --> []*Argument
		if arg_val.undercutter != "" {
			found = false
		LoopN:
			for _, stat := range ag.Statements {
				if arg_val.undercutter == stat.Id {
					found = true
					arg.Undercutter = stat
					break LoopN
				}
			}
			if !found {
				return ag, errors.New(" *** Semantic Error: Undercutter: " + arg_val.undercutter + ", from Argument: " + arg_id + ", is not a Statement-ID\n")
			}
		}
		for _, prem := range arg_val.premises {
			prem_stat, ok := stats[prem.stmt]
			if !ok {
				return ag, errors.New(" *** Semantic Error: Premise: " + prem.stmt + ", from Argument: " + arg_id + ", is not a Statement-ID\n")
			}
			if prem_stat == nil {
				// log.Printf("\n *** Prem Stat == nil für %s \n", prem.stmt)
			} else {
				// log.Printf(" \n +++ Prem_Stat: %s für %s \n", prem_stat.Id, prem.stmt)
			}
			caes_prem := caes.Premise{Stmt: prem_stat, Role: prem.role}
			if arg.Premises == nil {
				arg.Premises = []caes.Premise{caes_prem}
			} else {
				arg.Premises = append(arg.Premises, caes_prem)
			}
		}
		// Scheme
		if arg_val.scheme != "" {
			sch := caes.BasicSchemes[arg_val.scheme]
			arg.Scheme = &caes.Scheme{Id: sch.Id, Metadata: sch.Metadata, Eval: sch.Eval, Valid: sch.Valid}
		}
	}
	// log.Printf("   ---  Arguments --- \n %v \n ------End Arguments --- \n", ag.Arguments)
	for _, ass := range assumps {
		found = false
		for _, stat := range ag.Statements {
			if ass == stat.Id {
				found = true
				// log.Printf(" Set assumtions: %s\n", ass)
				stat.Assumed = true
			}
		}
		if !found {
			return ag, errors.New(" *** Semantic Error: Assumption: " + ass + ", is not a Statement-ID\n")
		}
	}

	// References
	ag.References = refs
	// log.Printf("   ---  References --- \n %v \n ------End References --- \n", ag.References)
	// Labels
	// if lbls not empty
	for _, stat := range ag.Statements {
		lbl, found := lbls[stat.Id]
		if found == true {
			// log.Printf(" Label %s:%v\n", stat.Id, lbl)
			stat.Label = lbl
		}
	}
	//check
	for lbl_Id, lbl_val := range lbls {
		found = false
	LoopLbl:
		for _, stat := range ag.Statements {
			if lbl_Id == stat.Id {
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
			return ag, errors.New(" *** Semantic Error: " + lbl_str + "- Label: " + lbl_Id + ", is not a Statement-ID\n")
		}
	}

	return ag, nil

}

func writeMetaData(f io.Writer, sp1 string, sp2 string, md caes.Metadata) {
	if md != nil {
		fmt.Fprintf(f, "%smetadata: \n", sp1)
		for md_key, md_val := range md {
			fmt.Fprintf(f, "%s%s: %s\n", sp2, md_key, md_val)
		}
	}
}

func ExportWithReferences(f io.Writer, ag *caes.ArgGraph) {
	writeArgGraph1(false, f, ag)
}

func Export(f io.Writer, ag *caes.ArgGraph) {
	writeArgGraph1(true, f, ag)
}

func writeArgGraph1(noRefs bool, f io.Writer, ag *caes.ArgGraph) {
	sp0 := ""
	sp1 := "    "
	sp2 := "        "
	sp3 := "            "

	writeMetaData(f, sp0, sp1, ag.Metadata)

	is := ag.Issues
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

	std := ag.Statements
	if std != nil {
		fmt.Fprintf(f, "statements: \n")
		for _, ref_stat := range std {

			if ref_stat.Metadata == nil && ref_stat.Assumed == false && ref_stat.Label == caes.Undecided && (noRefs == true || (ref_stat.Issue == nil && ref_stat.Args == nil)) {
				fmt.Fprintf(f, "%s%s: %s\n", sp1, ref_stat.Id, ref_stat.Text)
			} else {

				fmt.Fprintf(f, "%s%s: \n", sp1, ref_stat.Id)
				writeMetaData(f, sp2, sp3, ref_stat.Metadata)

				if ref_stat.Text != "" {
					fmt.Fprintf(f, "%stext: %s \n", sp2, ref_stat.Text)
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
	if ag.Arguments != nil {
		fmt.Fprintf(f, "arguments: \n")
		for _, ref_arg := range ag.Arguments {
			fmt.Fprintf(f, "%s%s:\n", sp1, ref_arg.Id)

			writeMetaData(f, sp2, sp3, ref_arg.Metadata)

			first := true
			list := true
			for _, prem := range ref_arg.Premises {
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
			if ref_arg.Conclusion != nil {
				fmt.Fprintf(f, "%sconclusion: %s\n", sp2, ref_arg.Conclusion.Id)
			}
			if ref_arg.Weight != 0.0 {
				fmt.Fprintf(f, "%sweight: %v\n", sp2, ref_arg.Weight)
			}
			if ref_arg.Scheme != nil {
				fmt.Fprintf(f, "%sscheme: %s\n", sp2, ref_arg.Scheme.Id)
			}
			if ref_arg.Undercutter != nil {
				fmt.Fprintf(f, "%sundercutter: %s\n", sp2, ref_arg.Undercutter.Id)
			}
		}
	}
	first := true
	for key, md := range ag.References {
		if first == true {
			fmt.Fprintf(f, "references:\n")
			first = false
		}
		fmt.Fprintf(f, "%s%s:\n", sp2, key)
		for md_key, md_val := range md {
			fmt.Fprintf(f, "%s%s: %v\n", sp3, md_key, md_val)
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

func iface2labels(value interface{}, lbls mapLabels) (mapLabels, error) {
	var err error
	switch subT := value.(type) {
	case mapIface:
		for lblkey, lblvalue := range value.(mapIface) {
			switch strings.ToLower(lblkey.(string)) {
			case "in":
				lbls, err = iface2lbl_1(lblvalue, lbls, caes.In)
			case "out":
				lbls, err = iface2lbl_1(lblvalue, lbls, caes.Out)
			case "undecided":
				lbls, err = iface2lbl_1(lblvalue, lbls, caes.Undecided)
			}
		}
	default:
		return lbls, errors.New("*** Error labels: (Type)" + subT.(string) + "\n")

	}
	return lbls, err
}

func iface2lbl_1(inArg interface{}, lbls mapLabels, label caes.Label) (mapLabels, error) {
	// log.Printf("labels:\n   %v: ", label)
	switch intype := inArg.(type) {
	case []interface{}:
		// log.Printf(" [")
		for idx, stat := range inArg.([]interface{}) {
			switch stype := stat.(type) {
			case string:
				lbls[stat.(string)] = label
				if idx != 0 {
					// log.Printf(", ")
				}
				// log.Printf("%s", stat.(string))
			default:
				return lbls, errors.New("*** Error labels [(Type)]:" + stype.(string) + "\n")
			}
		}
		// log.Printf("]")
	case string:
		lbls[inArg.(string)] = label
	default:
		return lbls, errors.New("*** Error labels (Type):" + intype.(string) + "\n")

	}
	return lbls, nil
}

func iface2metadata(value interface{}, meta caes.Metadata) (caes.Metadata, error) {
	switch subT := value.(type) {
	case mapIface:
		for metakey, metavalue := range value.(mapIface) {
			meta[metakey.(string)] = metavalue
			// log.Printf("         %s: %s\n", metakey.(string), meta[metakey.(string)])
		}
	default:
		return meta, errors.New("*** Error metadata (Type):" + subT.(string) + "\n")
	}
	return meta, nil
}

func iface2statement(value interface{}, stats mapStatements) (mapStatements, error) {

	var err error
	// log.Printf("Statements: \n")
	// switch t := value.(type) {
	switch value.(type) {
	case mapIface:
		for st_key, st_value := range value.(mapIface) {
			keyStr := st_key.(string)
			switch st_value.(type) {
			case string:
				stats[keyStr] = &caes.Statement{
					Id:    keyStr,
					Text:  st_value.(string),
					Label: caes.Undecided,
				}
				// log.Printf(" %v: %v \n", st_key.(string), st_value.(string))
			case int, float32, float64, bool:
				stats[keyStr] = &caes.Statement{
					Id:    keyStr,
					Text:  iface2string(st_value),
					Label: caes.Undecided,
				}
				// log.Printf(" %v: %v \n", st_key.(string), iface2string(st_value))
			case mapIface:
				// log.Printf("   %v:\n", st_key.(string))
				stats[keyStr], err = iface2xstatement(st_value, &caes.Statement{Id: st_key.(string), Label: caes.Undecided})
				if err != nil {
					return stats, err
				}
			}
		}
	default:
		// log.Printf(" Type: %v \n", t)
	}
	return stats, nil
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

func iface2issues(value interface{}, issues mapIssues) (mapIssues, error) {
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
											return issues,
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
										return issues,
											errors.New("*** Error: position expected DV,PE, CCE, BRD, wrong: " + issueValue.(string) + " \n")
									}
								default:
									return issues,
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
									return issues, err
								}
							}
						}
					}
				}
				issues[issueNameStr] = issue
			default:
				return issues,
					errors.New("*** ERROR: Not a issue Name: " + iT.(string) + "\n")
			}
		}
	}
	return issues, nil
}

func iface2assumps(value interface{}, assumps assumptions) (assumptions, error) {
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
				if assumps == nil {
					assumps = assumptions{str.(string)}
				} else {
					assumps = append(assumps, str.(string))
				}
				// log.Printf("%s", str.(string))
			default:
				if assumps == nil {
					assumps = assumptions{iface2string(str)}
				} else {
					assumps = append(assumps, iface2string(str))
				}
				// log.Printf("if=%v", iface2string(str))
			}
		}
		// log.Printf("]\n")
	case string:
		if assumps == nil {
			assumps = assumptions{value.(string)}
		} else {
			assumps = append(assumps, value.(string))
		}
		// log.Printf(" %s \n", value.(string))
	default:
		return assumps,
			errors.New("*** ERROR: not a right assumption-value: " + value.(string) + "\n")
	}
	return assumps, nil
}

func iface2arguments(value interface{}, args mapArguments) (mapArguments, error) {
	// log.Printf("Arguments: \n")
	switch value.(type) {
	case mapIface:
		for argName, arg := range value.(mapIface) {
			switch aT := argName.(type) {
			case string:
				// log.Printf("    %s: \n", argName.(string))
				var err error
				args[argName.(string)], err = iface2argument(arg, umArgument{})
				if err != nil {
					return args, err
				}
			default:
				return args,
					errors.New("*** ERROR: Arguement-name is not a string: " + aT.(string) + "\n")
			}
		}
	}
	return args, nil
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
					outArg.weigth, err = iface2weigth(attValue)
					if err != nil {
						return outArg, err
					}
				case "nas", "undercutter", "not app statement":
					outArg.undercutter = iface2string(attValue)
					// log.Printf(" %s \n", outArg.undercutter)
				case "scheme":
					outArg.scheme = iface2string(attValue)
					_, isIn := caes.BasicSchemes[outArg.scheme]
					if isIn == false {
						errStr := "*** ERROR: In argument wrong scheme value: " + outArg.scheme + "(expected: "
						first := true
						for schemeKey, _ := range caes.BasicSchemes {
							if first {
								errStr = errStr + schemeKey
								first = false
							} else {
								errStr = errStr + ", " + schemeKey
							}
						}
						return outArg,
							errors.New(errStr + ")\n")
					}
					// log.Printf(" %s \n", outArg.scheme)
				default:
					return outArg,
						errors.New("*** ERROR: Wrong argument attribut: " + attName.(string) + " (expected: conclusion, premises, weight, undercutter, scheme or metadata)\n")
				}
			default:
				return outArg,
					errors.New("*** ERROR: Wrong argument attribut type (string expected): " + ant.(string) + "\n")
			}
		}
	}
	return outArg, nil
}

func iface2weigth(attValue interface{}) (float64, error) {
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
			switch keyStr {
			case "major":
				premis.role = keyStr
				premis.stmt = valStr
				// log.Printf("      major: %s\n", valStr)
			case "minor":
				premis.role = keyStr
				premis.stmt = valStr
				// log.Printf("      minor: %s\n", valStr)
			default:
				return outArg,
					errors.New("*** ERROR: minor/majo expected and not" + keyStr + ": (value " + valStr + ")")

			}
			if outArg == nil {
				outArg = []umPremis{premis}
			} else {
				outArg = append(outArg, premis)
			}
		}
	}
	return outArg, nil
}

func iface2references(reference interface{}, refs mapReferences) (mapReferences, error) {
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
			refs[refNameStr], err = iface2metadata(refBody, caes.Metadata{})
			if err != nil {
				return refs, err
			}
		}
	}
	return refs, nil
}
