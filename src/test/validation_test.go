package test

import (
	"fmt"
	"os"
	"path"
	"testing"

	"strings"

	"github.com/carneades/carneades-4/src/engine/caes"
	"github.com/carneades/carneades-4/src/engine/caes/encoding/yaml"
	"github.com/carneades/carneades-4/src/engine/validation"
)

func TestValidateExamples(t *testing.T) {

	var ag *caes.ArgGraph
	var err error

	d, err := os.Open(yamlDir)
	defer d.Close()
	check(t, err)
	files, err := d.Readdir(0)
	check(t, err)
	for _, fi := range files {
		file, err := os.Open(yamlDir + fi.Name())
		defer file.Close()
		if err != nil {
			e := fmt.Errorf("%s, %s", file.Name(), err)
			t.Errorf(e.Error())
			continue
		}
		if path.Ext(file.Name()) == ".yml" {
			// skip non-YAML files
			//t.Logf(" =  =  =  =  =  =  Import %s =  =  =  =  =  = \n", fi.Name())

			ag, err = yaml.Import(file)
			if err != nil {
				e := fmt.Errorf("%s, %s", file.Name(), err)
				t.Errorf(e.Error())
				continue
			}

			// fmt.Printf("---------- CheckArgGraph %s ----------\n", filename1)
			// yaml.ExportWithReferences(os.Stdout, ag)
			// fmt.Printf("---------- End: WriteArgGraph %s ----------\n", filename1)
			//t.Logf(" -   -  -  -  -  -   Checking -  -  -  - \n")
			problems := validation.Validate(ag)

			if len(problems) > 0 {
				var p []string
				for _, prob := range problems {
					p = append(p, fmt.Sprintf("%s: %s: %s (%s)", prob.Category, prob.Id, prob.Description, prob.Expression))
				}
				t.Errorf("Check for %s failed. Unexpected problems:\n%v", fi.Name(), strings.Join(p, "\n"))
			}

		}
	}

}

func TestValidateMany(t *testing.T) {

	file, err := os.Open("testdata/mcda-porsche - Kopie.yml")
	defer file.Close()
	if err != nil {
		e := fmt.Errorf("%s, %s", file.Name(), err)
		t.Errorf(e.Error())
	}
	ag, err := yaml.Import(file)
	if err != nil {
		e := fmt.Errorf("%s, %s", file.Name(), err)
		t.Errorf(e.Error())
	}
	problems := validation.Validate(ag)
	numExpectedProblems := 48
	if len(problems) != numExpectedProblems {
		var p []string
		for _, prob := range problems {
			p = append(p, fmt.Sprintf("%s: %s %s: %s", prob.Category, prob.Description, prob.Id, prob.Expression))
		}
		t.Errorf("Check for %s failed. Expected %d problems, got %d problems:\n%v", file.Name(), numExpectedProblems, len(problems), strings.Join(p, "\n"))
	}

}

//TODO: Add test for wrong assumed text.
func TestValidateStatements(t *testing.T) {
	//defer recoverTesting("statements", t)
	//one test for each problem: "key not a term", "key not a ground atomic formula", "label wrong"(not in validation.go)
	possibleProblems := 2

	problems := getProblems("testdata/validateStatements.yml", t)

	probs := make(map[string][]validation.Problem)

	if len(problems) == possibleProblems {

		for _, prob := range problems {

			if prob.Category != 1 {
				t.Errorf("Expected category STATEMENT but got %s for expression: %s", prob.Category, prob.Expression)
			}
			switch prob.Description {
			case "key not a term":
				probs["noTerm"] = append(probs["noTerm"], prob)
			case "key not a ground atomic formula":
				probs["noGround"] = append(probs["noGround"], prob)
			case "label wrong": //TODO add in error throwing later
				probs["label"] = append(probs["label"], prob)
			default:
				probs["default"] = append(probs["default"], prob)
			}

		}
		if len(probs["default"]) > 0 {
			for _, prob := range probs["default"] {
				t.Errorf("Got validation result with unexpected description for Statement %s: %s", prob.Expression, prob.Description)

			}
		}
		if len(probs["noTerm"]) > 1 {
			var p []string
			for _, prob := range probs["noTerm"] {
				p = append(p, fmt.Sprintf("Category: %s, ID: %s, Espression: %s", prob.Category, prob.Id, prob.Expression))
			}
			t.Errorf("Got more validation results with description 'key not a term' than expected:\n%s", strings.Join(p, " \n"))

		}
		if len(probs["noGround"]) > 1 {
			var p []string
			for _, prob := range probs["noGround"] {
				p = append(p, fmt.Sprintf("Category: %s, ID: %s, Espression: %s", prob.Category, prob.Id, prob.Expression))
			}
			t.Errorf("Got more validation results with description 'key not a ground atomic formula' than expected:\n%s", strings.Join(p, " \n"))

		}

	} else if len(problems) == 0 {
		t.Errorf("Check for Statement Validator failed. Expected to get %d problems but got none", possibleProblems)
	} else {
		var p []string
		for _, prob := range problems {
			p = append(p, fmt.Sprintf("Category: %s, Expression: %s, Id: %s, Description: %s", prob.Category, prob.Expression, prob.Id, prob.Description))
		}
		t.Errorf("Check for Statement Validator failed. Expected to get %d problems but got %d:\n%v", possibleProblems, len(problems), strings.Join(p, "\n"))

	}

}

//Validation throws panic
func TestValidateIssues(t *testing.T) {
	defer recoverTesting("issues", t)
	problems := getProblems("testdata/validateIssues.yml", t)
	possibleProblems := 0 //2 but parser breaks for two

	if len(problems) != possibleProblems {
		var p []string
		for _, prob := range problems {
			p = append(p, fmt.Sprintf("Category: %s, Expression: %s, Id: %s, Description: %s", prob.Category, prob.Expression, prob.Id, prob.Description))
		}
		t.Errorf("Check for Argument Validator failed. Expected to get %d problems but got %d:\n%v", possibleProblems, len(problems), strings.Join(p, "\n"))

	}

	//TODO: check errors
}
func TestValidateArguments(t *testing.T) {
	defer recoverTesting("Arguments", t)
	problems := getProblems("testdata/validateArguments.yml", t)
	possibleProblems := 0 //7 but parser breaks for two
	if len(problems) != possibleProblems {
		var p []string
		for _, prob := range problems {
			p = append(p, fmt.Sprintf("Category: %s, Expression: %s, Id: %s, Description: %s", prob.Category, prob.Expression, prob.Id, prob.Description))
		}
		t.Errorf("Check for Argument Validator failed. Expected to get %d problems but got %d:\n%v", possibleProblems, len(problems), strings.Join(p, "\n"))

	}

}

func TestValidateLabels(t *testing.T) {
	defer recoverTesting("Labels", t)
	problems := getProblems("testdata/validateLabels.yml", t)
	possibleProblems := 4
	if len(problems) != possibleProblems {
		var p []string
		for _, prob := range problems {
			p = append(p, fmt.Sprintf("Category: %s, Expression: %s, Id: %s, Description: %s", prob.Category, prob.Expression, prob.Id, prob.Description))
		}
		t.Errorf("Check for Label Validator failed. Expected to get %d problems but got %d:\n%v", possibleProblems, len(problems), strings.Join(p, "\n"))

	}

}
func TestValidateLanguage(t *testing.T) {
	defer recoverTesting("Language", t)
	problems := getProblems("testdata/validateLanguage.yml", t)
	possibleProblems := 8

	probs := make(map[string][]validation.Problem)

	if len(problems) == possibleProblems {

		for _, prob := range problems {

			if prob.Category != 6 {
				t.Errorf("Expected category LANGUAGE but got %s for expression: %s", prob.Category, prob.Expression)
			}
			switch prob.Description {
			case "does not have the form predicate/arity":
				probs["form"] = append(probs["form"], prob)
			case "not a predicate symbol":
				probs["notPredicate"] = append(probs["notPredicate"], prob)
			case "non-integer arity":
				probs["arity"] = append(probs["arity"], prob)
			case "format string has incorrect number of placeholders (verbs)":
				probs["placeholder"] = append(probs["placeholder"], prob)
			default:
				probs["default"] = append(probs["default"], prob)
			}

		}
		if len(probs["default"]) > 0 {
			for _, prob := range probs["default"] {
				t.Errorf("Got validation result with unexpected description for Statement %s: %s", prob.Expression, prob.Description)

			}
		}
		if len(probs["form"]) > 1 {
			var p []string
			for _, prob := range probs["form"] {
				p = append(p, fmt.Sprintf("Category: %s, ID: %s, Espression: %s", prob.Category, prob.Id, prob.Expression))
			}
			t.Errorf("Got more validation results with description 'does not have the form predicate/arity' than expected:\n%s", strings.Join(p, " \n"))

		}
		if len(probs["notPredicate"]) > 2 {
			var p []string
			for _, prob := range probs["notPredicate"] {
				p = append(p, fmt.Sprintf("Category: %s, ID: %s, Espression: %s", prob.Category, prob.Id, prob.Expression))
			}
			t.Errorf("Got more validation results with description 'not a predicate symbol' than expected:\n%s", strings.Join(p, " \n"))

		}
		if len(probs["arity"]) > 2 {
			var p []string
			for _, prob := range probs["arity"] {
				p = append(p, fmt.Sprintf("Category: %s, ID: %s, Espression: %s", prob.Category, prob.Id, prob.Expression))
			}
			t.Errorf("Got more validation results with description 'non-integer arity' than expected:\n%s", strings.Join(p, " \n"))

		}
		if len(probs["placeholder"]) > 3 {
			var p []string
			for _, prob := range probs["placeholder"] {
				p = append(p, fmt.Sprintf("Category: %s, ID: %s, Espression: %s", prob.Category, prob.Id, prob.Expression))
			}
			t.Errorf("Got more validation results with description 'format string has incorrect number of placeholders (verbs)' than expected:\n%s", strings.Join(p, " \n"))

		}

	} else if len(problems) == 0 {
		t.Errorf("Check for Statement Validator failed. Expected to get %d problems but got none", possibleProblems)
	} else {
		var p []string
		for _, prob := range problems {
			p = append(p, fmt.Sprintf("Category: %s, Expression: %s, Id: %s, Description: %s", prob.Category, prob.Expression, prob.Id, prob.Description))
		}
		t.Errorf("Check for Statement Validator failed. Expected to get %d problems but got %d:\n%v", possibleProblems, len(problems), strings.Join(p, "\n"))

	}

}
func TestValidateAssumptions(t *testing.T) {
	defer recoverTesting("Assumptions", t)

	possibleProblems := 5
	problems := getProblems("testdata/validateAssumptions.yml", t)

	probs := make(map[string][]validation.Problem)

	if len(problems) == possibleProblems {

		for _, prob := range problems {

			if prob.Category != 4 {
				t.Errorf("Expected category ASSUMPTION but got %s for expression: %s", prob.Category, prob.Expression)
			}
			switch prob.Description {
			case "not a term":
				probs["noTerm"] = append(probs["noTerm"], prob)
			case "not a ground atomic formula":
				probs["noGround"] = append(probs["noGround"], prob)
			default:
				probs["default"] = append(probs["default"], prob)
			}

		}
		if len(probs["default"]) > 0 {
			for _, prob := range probs["default"] {
				t.Errorf("Got validation result with unexpected description for Statement %s: %s", prob.Expression, prob.Description)

			}
		}
		if len(probs["noTerm"]) > 2 {
			var p []string
			for _, prob := range probs["noTerm"] {
				p = append(p, fmt.Sprintf("Category: %s, ID: %s, Espression: %s", prob.Category, prob.Id, prob.Expression))
			}
			t.Errorf("Got more validation results with description 'not a term' than expected:\n%s", strings.Join(p, " \n"))

		}
		if len(probs["noGround"]) > 3 {
			var p []string
			for _, prob := range probs["noGround"] {
				p = append(p, fmt.Sprintf("Category: %s, ID: %s, Espression: %s", prob.Category, prob.Id, prob.Expression))
			}
			t.Errorf("Got more validation results with description 'not a ground atomic formula' than expected:\n%s", strings.Join(p, " \n"))

		}

	} else if len(problems) == 0 {
		t.Errorf("Check for Assumption Validator failed. Expected to get %d problems but got none", possibleProblems)
	} else {
		var p []string
		for _, prob := range problems {
			p = append(p, fmt.Sprintf("Category: %s, Expression: %s, Id: %s, Description: %s", prob.Category, prob.Expression, prob.Id, prob.Description))
		}
		t.Errorf("Check for Assumption Validator failed. Expected to get %d problems but got %d:\n%v", possibleProblems, len(problems), strings.Join(p, "\n"))

	}

}

func TestArgumentSchemes(t *testing.T) {
	defer recoverTesting("ArgumentSchemes", t)
	problems := getProblems("testdata/validateArgumentSchemes.yml", t)

	possibleProblems := 11
	probs := make(map[string][]validation.Problem)

	if len(problems) == possibleProblems {

		for _, prob := range problems {

			if prob.Category != 7 {
				t.Errorf("Expected category \"argument scheme\" but got %s for expression: %s", prob.Category, prob.Expression)
			}
			switch prob.Description {
			case "duplicate scheme id":
				probs["duplicate"] = append(probs["duplicate"], prob)
			case "not a variable":
				probs["noVar"] = append(probs["noVar"], prob)
			case "not a term":
				probs["noTerm"] = append(probs["noTerm"], prob)
			case "variable not used in premise":
				probs["varNotPremise"] = append(probs["varNotPremise"], prob)
			case "predicate not declared in the language":
				probs["predicateNotDeclared"] = append(probs["predicateNotDeclared"], prob)
			case "variable not declared in the scheme":
				probs["varNotDeclared"] = append(probs["varNotDeclared"], prob)
			default:
				probs["default"] = append(probs["default"], prob)
			}

		}
		if len(probs["default"]) > 0 {
			for _, prob := range probs["default"] {
				t.Errorf("Got validation result with unexpected description for Statement %s: %s", prob.Expression, prob.Description)

			}
		}
		if len(probs["duplicate"]) > 1 {
			var p []string
			for _, prob := range probs["duplicate"] {
				p = append(p, fmt.Sprintf("Category: %s, ID: %s, Espression: %s", prob.Category, prob.Id, prob.Expression))
			}
			t.Errorf("Got more validation results with description 'duplicate scheme id' than expected:\n%s", strings.Join(p, " \n"))

		}
		if len(probs["noVar"]) > 1 {
			var p []string
			for _, prob := range probs["noVar"] {
				p = append(p, fmt.Sprintf("Category: %s, ID: %s, Espression: %s", prob.Category, prob.Id, prob.Expression))
			}
			t.Errorf("Got more validation results with description 'not a variable' than expected:\n%s", strings.Join(p, " \n"))

		}
		if len(probs["noTerm"]) > 1 {
			var p []string
			for _, prob := range probs["noTerm"] {
				p = append(p, fmt.Sprintf("Category: %s, ID: %s, Espression: %s", prob.Category, prob.Id, prob.Expression))
			}
			t.Errorf("Got more validation results with description 'not a term' than expected:\n%s", strings.Join(p, " \n"))

		}
		if len(probs["varNotPremise"]) > 1 {
			var p []string
			for _, prob := range probs["varNotPremise"] {
				p = append(p, fmt.Sprintf("Category: %s, ID: %s, Espression: %s", prob.Category, prob.Id, prob.Expression))
			}
			t.Errorf("Got more validation results with description 'variable not used in premise' than expected:\n%s", strings.Join(p, " \n"))

		}
		if len(probs["predicateNotDeclared"]) > 3 {
			var p []string
			for _, prob := range probs["predicateNotDeclared"] {
				p = append(p, fmt.Sprintf("Category: %s, ID: %s, Espression: %s", prob.Category, prob.Id, prob.Expression))
			}
			t.Errorf("Got more validation results with description 'predicate not declared in the language' than expected:\n%s", strings.Join(p, " \n"))

		}
		if len(probs["varNotDeclared"]) > 5 {
			var p []string
			for _, prob := range probs["varNotDeclared"] {
				p = append(p, fmt.Sprintf("Category: %s, ID: %s, Espression: %s", prob.Category, prob.Id, prob.Expression))
			}
			t.Errorf("Got more validation results with description 'variable not declared in the scheme' than expected:\n%s", strings.Join(p, " \n"))

		}

	} else if len(problems) == 0 {
		t.Errorf("Check for Argument Scheme Validator failed. Expected to get %d problems but got none", possibleProblems)
	} else {
		var p []string
		for _, prob := range problems {
			p = append(p, fmt.Sprintf("Category: %s, Expression: %s, Id: %s, Description: %s", prob.Category, prob.Expression, prob.Id, prob.Description))
		}
		t.Errorf("Check for Argument Scheme Validator failed. Expected to get %d problems but got %d:\n%v", possibleProblems, len(problems), strings.Join(p, "\n"))

	}

}

func TestValidateIssueSchemes(t *testing.T) {
	defer recoverTesting("IssueSchemes", t)
	problems := getProblems("testdata/validateIssueSchemes.yml", t)

	possibleProblems := 3
	probs := make(map[string][]validation.Problem)

	if len(problems) == possibleProblems {

		for _, prob := range problems {

			if prob.Category != 8 {
				t.Errorf("Expected category \"issue scheme\" but got %s for expression: %s", prob.Category, prob.Expression)
			}
			switch prob.Description {
			case "predicate of pattern not declared in the language":
				probs["predicateNotDeclared"] = append(probs["predicateNotDeclared"], prob)
			case "pattern is not a term":
				probs["notTerm"] = append(probs["notTerm"], prob)
			case "fewer than two patterns":
				probs["fewPatterns"] = append(probs["fewPatterns"], prob)
			default:
				probs["default"] = append(probs["default"], prob)
			}

		}
		if len(probs["default"]) > 0 {
			for _, prob := range probs["default"] {
				t.Errorf("Got validation result with unexpected description for Statement %s: %s", prob.Expression, prob.Description)

			}
		}
		if len(probs["duplicate"]) > 1 {
			var p []string
			for _, prob := range probs["duplicate"] {
				p = append(p, fmt.Sprintf("Category: %s, ID: %s, Espression: %s", prob.Category, prob.Id, prob.Expression))
			}
			t.Errorf("Got more validation results with description 'duplicate scheme id' than expected:\n%s", strings.Join(p, " \n"))

		}
		if len(probs["noVar"]) > 1 {
			var p []string
			for _, prob := range probs["noVar"] {
				p = append(p, fmt.Sprintf("Category: %s, ID: %s, Espression: %s", prob.Category, prob.Id, prob.Expression))
			}
			t.Errorf("Got more validation results with description 'not a variable' than expected:\n%s", strings.Join(p, " \n"))

		}
		if len(probs["noTerm"]) > 1 {
			var p []string
			for _, prob := range probs["noTerm"] {
				p = append(p, fmt.Sprintf("Category: %s, ID: %s, Espression: %s", prob.Category, prob.Id, prob.Expression))
			}
			t.Errorf("Got more validation results with description 'not a term' than expected:\n%s", strings.Join(p, " \n"))

		}
		if len(probs["predicateNotDeclared"]) > 3 {
			var p []string
			for _, prob := range probs["predicateNotDeclared"] {
				p = append(p, fmt.Sprintf("Category: %s, ID: %s, Espression: %s", prob.Category, prob.Id, prob.Expression))
			}
			t.Errorf("Got more validation results with description 'predicate not declared in the language' than expected:\n%s", strings.Join(p, " \n"))

		}
		if len(probs["varNotDeclared"]) > 3 {
			var p []string
			for _, prob := range probs["varNotDeclared"] {
				p = append(p, fmt.Sprintf("Category: %s, ID: %s, Espression: %s", prob.Category, prob.Id, prob.Expression))
			}
			t.Errorf("Got more validation results with description 'variable not declared in the scheme' than expected:\n%s", strings.Join(p, " \n"))

		}

	} else if len(problems) == 0 {
		t.Errorf("Check for Argument Scheme Validator failed. Expected to get %d problems but got none", possibleProblems)
	} else {
		var p []string
		for _, prob := range problems {
			p = append(p, fmt.Sprintf("Category: %s, Expression: %s, Id: %s, Description: %s", prob.Category, prob.Expression, prob.Id, prob.Description))
		}
		t.Errorf("Check for Argument Scheme Validator failed. Expected to get %d problems but got %d:\n%v", possibleProblems, len(problems), strings.Join(p, "\n"))

	}

}

func getProblems(testfile string, t *testing.T) []validation.Problem {
	file, err := os.Open(testfile)
	defer file.Close()
	if err != nil {
		e := fmt.Errorf("%s, %s", file.Name(), err)
		t.Errorf(e.Error())
	}
	ag, err := yaml.Import(file)
	if err != nil {
		e := fmt.Errorf("%s, %s", file.Name(), err)
		t.Errorf(e.Error())
		t.Skip()
	}
	return validation.Validate(ag)

}

func recoverTesting(testcase string, t *testing.T) {
	if r := recover(); r != nil {

		t.Errorf("Testing validation for %s threw an unexpected panic", testcase)
	}
}
