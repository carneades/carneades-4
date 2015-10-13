// Copyright © 2015 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

// Visualizing Dung AFs using [Dot](http://www.graphviz.org/Documentation.php)
package dot

import (
	"fmt"
	"github.com/carneades/carneades-4/src/engine/dung"
	"io"
	"strings"
)

const (
	maxCharsPerRow = 30
	maxRows        = 3
	//
	black  = "#000000"
	red    = "#FF0000"
	green  = "#3AB54A"
	yellow = "#FCEE21"
	white  = "#FFFFFF"
	// line types
	line   = "filled"
	dashed = "dashed"
	//
	thinLine   = "1.0"
	mediumLine = "2.0"
	boldLine   = "3.0"
	//
	noArrow   = "none"
	withArrow = "normal"
	// font
	font     = "Arial"
	fontsize = "16"
	// shapeType
	// diamond   = "diamond"
	rectangle = "box"
	// ellipse   = "ellipse"
	roundrectangle = "Mrecord"
	hexagon        = "hexagon"
)

type Edge struct {
	id         string
	source     string
	target     string
	color      string
	lineType   string
	width      string
	lineTarget string
	edgeLabel  string
}

type Node struct {
	id              string
	color           string
	borderLine      string
	borderWidth     string
	nodeLabel       string
	underlinedLabel bool
	shapeType       string
}

var graphNr, nodeNr, edgeNr int

func newNode() Node {
	nodeNr++
	return Node{
		id: fmt.Sprintf("n%v", nodeNr),
		// color: "",
		borderLine:  line,
		borderWidth: thinLine,
		// nodeLabel: "",
		underlinedLabel: false,
		shapeType:       roundrectangle,
	}
}

func newEdge() Edge {
	edgeNr++
	return Edge{
		id:         fmt.Sprintf("e%v", edgeNr),
		source:     "n1",
		target:     "n2",
		color:      black,
		lineType:   line,
		width:      thinLine,
		lineTarget: withArrow,
		// edgeLabel: "",
	}
}

func p(w io.Writer, strs ...string) {
	for _, s := range strs {
		fmt.Fprintln(w, s)
	}
}

func pHead(w io.Writer, number int) {
	p(w, "digraph G"+fmt.Sprintf("%d", number)+" {", "rankdir=RL")
}

func pFoot(w io.Writer) {
	p(w, "}")
}

func trimmString(inStr string, underlined bool) string {
	outStr := ""
	cChars := 0
	cRows := 0
	cliStr := strings.Split(inStr, " ")
	//	lenCliStr := len(cliStr)
	for _, str := range cliStr {
		//p1(w, fmt.Sprintf("%v", cChars))
		if cChars+len(str) > maxCharsPerRow {
			//p1(w, "[GtC]")
			cRows += 1
			if cRows >= maxRows {
				// p1(w, "[GtR]")
				// rest String
				if maxCharsPerRow-cChars-1 > 0 {
					outStr += str[:maxCharsPerRow-cChars-1] + "\u2026"
				} else {
					outStr += "\u2026"
				}
				return outStr
			} else {
				if (float32(len(str)) > 0.5*float32(maxCharsPerRow)) &&
					(float32(maxCharsPerRow-cChars) > 0.7*float32(len(str))) {
					// rest String drucken,
					if underlined {
						outStr += str[:maxCharsPerRow-cChars-1] + "\u2025<br/>"
					} else {
						outStr += str[:maxCharsPerRow-cChars-1] + "\u2025\\n"
					}
					cChars = 0
				} else {
					// p1(w, "*")
					if underlined {
						outStr += "<br/>" + str + " "
					} else {
						outStr += "\\n" + str + " "
					}

					cChars = len(str) + 1
				}
			}
		} else {
			outStr += str + " "
			cChars += len(str) + 1
		}
	}
	return outStr
}

func pNodes(w io.Writer, nodes []Node) {

	p(w, "node [shape=box, style=filled, penwidth=1, fontname="+font+", fontsize="+fontsize+"]")
	p(w, "edge [fontsize="+fontsize+", color=black]")

	for _, node := range nodes {

		fillcolor := white
		if node.color != "" {
			fillcolor = node.color
		}
		nodeLabel := trimmString(node.nodeLabel, node.underlinedLabel)
		label := ""
		if node.underlinedLabel {
			label = "<<u>" + nodeLabel + "</u>>"
		} else {
			label = "\"" + nodeLabel + "\""
		}

		// height := "30.0"
		// width := calculateWidth(len(node.nodeLabel))

		p(w, node.id+
			" [label="+label+
			", penwidth="+node.borderWidth+
			", fillcolor=\""+fillcolor+"\""+
			", shape=\""+node.shapeType+"\""+
			", style=\""+node.borderLine+"\""+
			//			", height="+height+
			//			", width="+width+
			" ]")
	}
}

func pEdges(w io.Writer, edges []Edge) {

	for _, edge := range edges {

		color := black
		if edge.color != "" {
			color = edge.color
		}

		p(w, edge.source+" -> "+edge.target+
			" [label=\""+edge.edgeLabel+"\""+
			", penwidth="+edge.width+
			", style="+edge.lineType+
			", color=\""+color+"\""+
			", arrowhead="+edge.lineTarget+
			" ]")

		/*
			p(w, "   <edge id=\""+edge.id+
			p(w, "      <data key=\"d9\"/>",
				"      <data key=\"d10\">",
		*/
	}
}

func mkNodesAndEdges(af dung.AF) (nodes []Node, edges []Edge, err error) {
	firstNode := true
	firstEdge := true
	// Arguments
	for _, arg := range af.Args() {
		nNode := newNode()
		nNode.nodeLabel = arg.String()
		nNode.id = arg.String()
		if firstNode {
			nodes = []Node{nNode}
			firstNode = false
		} else {
			nodes = append(nodes, nNode)
		}
	}
	// Attacks
	for arg, attackers := range af.Atks() {
		for _, attacker := range attackers {
			edge := newEdge()
			edge.source = attacker.String()
			edge.target = arg.String()
			if firstEdge {
				edges = []Edge{edge}
				firstEdge = false
			} else {
				edges = append(edges, edge)
			}
		}
	}
	return
}

func Export(w io.Writer, af dung.AF) error {
	nodes, edges, err := mkNodesAndEdges(af)
	if err != nil {
		return err
	}
	graphNr++
	pHead(w, graphNr)
	pNodes(w, nodes)
	pEdges(w, edges)
	pFoot(w)
	return nil
}
