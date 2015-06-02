// Copyright © 2015 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

// Visualizing Dung AFs using [GraphML](http://graphml.graphdrawing.org/)
package graphml

import (
	"fmt"
	"github.com/carneades/carneades-go/internal/engine/dung"
	"io"
)

func p(w io.Writer, strs ...string) {
	for _, s := range strs {
		fmt.Fprintln(w, s)
	}
}

var graphNr int

func pHead(w io.Writer) {
	p(w, "<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"no\"?> ",
		"<graphml xmlns=\"http://graphml.graphdrawing.org/xmlns\" ",
		"     xmlns:java=\"http://www.yworks.com/xml/yfiles-common/1.0/java\" ",
		"     xmlns:sys=\"http://www.yworks.com/xml/yfiles-common/markup/primitives/2.0\" ",
		"     xmlns:x=\"http://www.yworks.com/xml/yfiles-common/markup/2.0\" ",
		"     xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\" ",
		"     xmlns:y=\"http://www.yworks.com/xml/graphml\" ",
		"     xmlns:yed=\"http://www.yworks.com/xml/yed/3\" ",
		"     xsi:schemaLocation=\"http://graphml.graphdrawing.org/xmlns http://www.yworks.com/xml/schema/graphml/1.1/ygraphml.xsd\"> ",
		"<key for=\"node\" id=\"d5\" yfiles.type=\"nodegraphics\"/>")
	graphNr = 1
}

func pFoot(w io.Writer) {
	p(w, "</graphml>")
}

func pNodes(w io.Writer, arg []dung.Arg, extension dung.ArgSet) {
	for _, node := range arg {
		p(w, "   <node id=\""+string(node)+"\">",
			"      <data key=\"d5\">",
			"      <y:ShapeNode>")
		if b, ok := extension[node]; ok && b {
			p(w, "      <y:Fill color=\"#99cc00\" transparent=\"false\"/>")
		} else {
			p(w, "      <y:Fill color=\"#ffcc00\" transparent=\"false\"/>")
		}
		p(w, "      <y:NodeLabel >"+string(node)+"</y:NodeLabel>",
			"       <y:Shape type=\"ellipse\"/>",
			"       </y:ShapeNode>",
			"       </data>",
			"   </node>")
	}
}

func pEdges(w io.Writer, atks map[dung.Arg][]dung.Arg) {
	for target, nodes := range atks {
		for i, source := range nodes {
			p(w, "   <edge id=\"e-"+
				string(target)+
				fmt.Sprintf("-%d", i)+
				"\" source=\""+
				string(source)+
				"\" target=\""+
				string(target)+"\"/>")
		}
	}
}

func Export(w io.Writer, af dung.AF, extension dung.ArgSet) {
	pHead(w)
	p(w, "<graph edgedefault=\"directed\" id=\"G"+
		fmt.Sprintf("%d", graphNr)+"\">")
	graphNr++
	pNodes(w, af.Args(), extension)
	pEdges(w, af.Atks())
	p(w, "</graph>")
	pFoot(w)
}
