package web

import (
	"bytes"
	"fmt"
	"github.com/carneades/carneades-4/src/engine/caes"
	"github.com/carneades/carneades-4/src/engine/caes/encoding/agxml"
	"github.com/carneades/carneades-4/src/engine/caes/encoding/aif"
	"github.com/carneades/carneades-4/src/engine/caes/encoding/caf"
	"github.com/carneades/carneades-4/src/engine/caes/encoding/dot"
	"github.com/carneades/carneades-4/src/engine/caes/encoding/graphml"
	"github.com/carneades/carneades-4/src/engine/caes/encoding/lkif"
	"github.com/carneades/carneades-4/src/engine/caes/encoding/yaml"
	"github.com/carneades/carneades-4/src/engine/dung"
	ddot "github.com/carneades/carneades-4/src/engine/dung/encoding/dot"
	dgml "github.com/carneades/carneades-4/src/engine/dung/encoding/graphml"
	"github.com/carneades/carneades-4/src/engine/dung/encoding/tgf"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

const afLimit = 20 // max number of arguments handled by the Dung solver

type templateHandler struct {
	once         sync.Once
	filename     string
	templatesDir string
	templ        *template.Template
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join(t.templatesDir, t.filename)))
	})
	t.templ.Execute(w, nil)
}

func CarneadesServer(port string, templatesDir string) {

	var errorTemplate = template.Must(template.ParseFiles(filepath.Join(templatesDir, "error.html")))

	evalHandler := func(w http.ResponseWriter, req *http.Request) {
		inputFormat := req.FormValue("input-format")
		outputFormat := req.FormValue("output-format")
		file, _, err := req.FormFile("datafile")
		if err != nil {
			errorTemplate.Execute(w, err.Error())
			return
		}
		data, err := ioutil.ReadAll(file)
		if err != nil {
			errorTemplate.Execute(w, err.Error())
			return
		}

		var ag *caes.ArgGraph
		rd := bytes.NewReader(data)

		switch inputFormat {
		case "yaml":
			ag, err = yaml.Import(rd)
			if err != nil {
				errorTemplate.Execute(w, err.Error())
				return
			}
		case "agxml":
			ag, err = agxml.Import(rd)
			if err != nil {
				errorTemplate.Execute(w, err.Error())
				return
			}
		case "aif":
			ag, err = aif.Import(rd)
			if err != nil {
				errorTemplate.Execute(w, err.Error())
				return
			}
		case "lkif":
			ag, err = lkif.Import(rd)
			if err != nil {
				errorTemplate.Execute(w, err.Error())
				return
			}
		case "caf":
			ag, err = caf.Import(rd)
			if err != nil {
				errorTemplate.Execute(w, err.Error())
				return
			}
		default:
			errorTemplate.Execute(w, fmt.Sprintf("unknown or unsupported input format: %s\n", inputFormat))
			return
		}

		// evaluate the argument graph, using grounded semantics
		// and update the labels of the statements in the argument graph
		l := ag.GroundedLabelling()
		// fmt.Printf("labelling=%v\n", l)
		ag.ApplyLabelling(l)

		switch outputFormat {
		case "yaml":
			yaml.Export(w, ag)
		case "graphml":
			err = graphml.Export(w, ag)
			if err != nil {
				errorTemplate.Execute(w, err.Error())
				return
			}
		case "dot":
			err = dot.Export(w, *ag)
			if err != nil {
				errorTemplate.Execute(w, err.Error())
				return
			}
		case "png", "svg":
			cmd := exec.Command("dot", "-T"+outputFormat)
			w2 := bytes.NewBuffer([]byte{})
			cmd.Stdin = w2
			cmd.Stdout = w
			err = dot.Export(w2, *ag)
			if err != nil {
				errorTemplate.Execute(w, err.Error())
				return
			}
			err = cmd.Run()
			if err != nil {
				errorTemplate.Execute(w, err.Error())
				return
			}
		default:
			errorTemplate.Execute(w, fmt.Sprintf("unknown or unsupported output format: %s\n", outputFormat))
			return
		}
	}

	dungHandler := func(w http.ResponseWriter, req *http.Request) {
		semantics := req.FormValue("semantics")
		outputFormat := req.FormValue("output-format")
		file, _, err := req.FormFile("datafile")
		if err != nil {
			errorTemplate.Execute(w, err.Error())
			return
		}
		data, err := ioutil.ReadAll(file)
		if err != nil {
			errorTemplate.Execute(w, err.Error())
			return
		}

		var af dung.AF
		rd := bytes.NewReader(data)

		af, err = tgf.Import(rd)
		if err != nil {
			errorTemplate.Execute(w, err.Error())
			return
		} else if len(af.Args()) > afLimit {
			errorTemplate.Execute(w, fmt.Sprintf("Argumentation frameworks with more than %v arguments are not supported by this server.\n", afLimit))
			return
		}

		// evaluate the argumentation framework, using the selected semantics
		var extensions []dung.ArgSet

		switch semantics {
		case "complete":
			if outputFormat == "text" {
				extensions = af.CompleteExtensions()
			} else {
				e, ok := af.SomeExtension(dung.Complete)
				if ok {
					extensions = []dung.ArgSet{e}
				} else {
					extensions = []dung.ArgSet{}
				}
			}
		case "preferred":
			if outputFormat == "text" {
				extensions = af.PreferredExtensions()
			} else {
				e, ok := af.SomeExtension(dung.Preferred)
				if ok {
					extensions = []dung.ArgSet{e}
				} else {
					extensions = []dung.ArgSet{}
				}
			}
		case "stable":
			if outputFormat == "text" {
				extensions = af.StableExtensions()
			} else {
				e, ok := af.SomeExtension(dung.Stable)
				if ok {
					extensions = []dung.ArgSet{e}
				} else {
					extensions = []dung.ArgSet{}
				}
			}
		default:
			extensions = []dung.ArgSet{af.GroundedExtension()}
		}

		printExtensions := func(extensions []dung.ArgSet) {
			s := []string{}
			for _, E := range extensions {
				s = append(s, E.String())
			}
			fmt.Fprintf(w, "[%s]\n", strings.Join(s, ","))
		}

		switch outputFormat {
		case "graphml":
			var as dung.ArgSet
			if len(extensions) == 0 {
				as = dung.NewArgSet()
			} else {
				as = extensions[0]
			}
			dgml.Export(w, af, as)
		case "dot":
			err = ddot.Export(w, af)
			if err != nil {
				errorTemplate.Execute(w, err.Error())
				return
			}
		case "png", "svg":
			cmd := exec.Command("dot", "-T"+outputFormat)
			w2 := bytes.NewBuffer([]byte{})
			cmd.Stdin = w2
			cmd.Stdout = w
			err = ddot.Export(w2, af)
			if err != nil {
				errorTemplate.Execute(w, err.Error())
				return
			}
			err = cmd.Run()
			if err != nil {
				errorTemplate.Execute(w, err.Error())
				return
			}
		default: // text
			printExtensions(extensions)
		}
	}

	http.Handle("/", &templateHandler{filename: "carneades.html", templatesDir: templatesDir})
	http.Handle("/help", &templateHandler{filename: "help.html", templatesDir: templatesDir})
	http.Handle("/eval-form", &templateHandler{filename: "eval-form.html", templatesDir: templatesDir})
	http.Handle("/eval-help", &templateHandler{filename: "eval-help.html", templatesDir: templatesDir})
	http.HandleFunc("/eval", evalHandler)
	http.Handle("/dung-form", &templateHandler{filename: "dung-form.html", templatesDir: templatesDir})
	http.Handle("/dung-help", &templateHandler{filename: "dung-help.html", templatesDir: templatesDir})
	http.HandleFunc("/dung", dungHandler)
	http.Handle("/imprint", &templateHandler{filename: "imprint.html", templatesDir: templatesDir})

	// start the web server
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("carneades:", err)
	}
}
