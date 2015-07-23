package web

import (
	"bytes"
	"fmt"
	"github.com/carneades/carneades-4/internal/engine/caes"
	"github.com/carneades/carneades-4/internal/engine/caes/encoding/agxml"
	"github.com/carneades/carneades-4/internal/engine/caes/encoding/aif"
	"github.com/carneades/carneades-4/internal/engine/caes/encoding/dot"
	"github.com/carneades/carneades-4/internal/engine/caes/encoding/graphml"
	"github.com/carneades/carneades-4/internal/engine/caes/encoding/lkif"
	"github.com/carneades/carneades-4/internal/engine/caes/encoding/yaml"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"path/filepath"
	"sync"
)

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
	// root
	http.Handle("/", &templateHandler{filename: "carneades.html", templatesDir: templatesDir})
	http.Handle("/eval-form", &templateHandler{filename: "eval.html", templatesDir: templatesDir})
	http.Handle("/dung-form", &templateHandler{filename: "dung.html", templatesDir: templatesDir})
	http.HandleFunc("/eval", evalHandler)
	// start the web server
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("CarneadesServer:", err)
	}
}
