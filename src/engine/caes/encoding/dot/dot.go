// Copyright © 2015 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

// Visualizing CAES AG using [Dot](http://www.graphviz.org/Documentation.php)
package dot

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/carneades/carneades-4/src/engine/caes"
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

type gmlEdge struct {
	id         string
	source     string
	target     string
	color      string
	lineType   string
	width      string
	lineTarget string
	edgeLabel  string
}

type gmlNode struct {
	id              string
	color           string
	borderLine      string
	borderWidth     string
	nodeLabel       string
	underlinedLabel bool
	shapeType       string
}

var graphNr, nodeNr, edgeNr int

func newGmlNode() gmlNode {
	nodeNr++
	return gmlNode{
		id: fmt.Sprintf("n%v", nodeNr),
		// color: "",
		borderLine:  line,
		borderWidth: thinLine,
		// nodeLabel: "",
		underlinedLabel: false,
		shapeType:       rectangle,
	}
}

func newGmlEdge() gmlEdge {
	edgeNr++
	return gmlEdge{
		id:         fmt.Sprintf("e%v", edgeNr),
		source:     "n1",
		target:     "n2",
		color:      black,
		lineType:   line,
		width:      thinLine,
		lineTarget: noArrow,
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

func pNodes(w io.Writer, nodes []gmlNode) {

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

		p(w, "node_"+node.id+
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

func pEdges(w io.Writer, edges []gmlEdge) {

	for _, edge := range edges {

		color := black
		if edge.color != "" {
			color = edge.color
		}

		p(w, "node_"+edge.source+" -> node_"+edge.target+
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

func mkNodesAndEdges(ag *caes.ArgGraph) (nodes []gmlNode, edges []gmlEdge, err error) {
	assums := caes.SliceToMap(ag.Assumptions)
	stat2Node := make(map[string]string)
	firstNode := true
	firstEdge := true
	// Statements []
	for _, stat := range ag.Statements {
		// ignore undercutters unless they have supporting arguments
		if stat.IsUndercutter && stat.Args == nil { continue }
		nNode := newGmlNode() // shapeType [] (rectangle)
		stat2Node[stat.Id] = nNode.id
		nNode.nodeLabel = stat.Text
		if assums[stat.Id] {
			nNode.underlinedLabel = true
		}
		switch stat.Label {
		case caes.Out:
			nNode.color = red
		case caes.In:
			nNode.color = green
		case caes.Undecided:
			nNode.color = yellow
		}
		if firstNode {
			nodes = []gmlNode{nNode}
			firstNode = false
		} else {
			nodes = append(nodes, nNode)
		}
	}
	// Arguments O
	for _, arg := range ag.Arguments {
		nNode := newGmlNode()
		nNode.shapeType = roundrectangle // O
		if arg.Scheme != nil {
			nNode.nodeLabel = fmt.Sprintf("%s: %s", arg.Id, arg.Scheme.Id)
		} else {
			nNode.nodeLabel = arg.Id
		}
		if firstNode {
			nodes = []gmlNode{nNode}
			firstNode = false
		} else {
			nodes = append(nodes, nNode)
		}
		// argument --- premises
		for _, prem := range arg.Premises {
			nodeId, found := stat2Node[prem.Stmt.Id]
			if found == false {
				err = errors.New(" *** Premises-Statement: " + prem.Stmt.Id + "from Argument: " + arg.Id + "not found")
				return
			}
			edge := newGmlEdge()
			edge.source = nodeId
			edge.target = nNode.id
			edge.edgeLabel = prem.Role
			if firstEdge {
				edges = []gmlEdge{edge}
				firstEdge = false
			} else {
				edges = append(edges, edge)
			}
		}
		// conclusion <---- argument
		// conclusion <-- value -- argument
		if arg.Conclusion != nil {
			edge := newGmlEdge()
			edge.lineTarget = withArrow
			edge.source = nNode.id
			nodeId, found := stat2Node[arg.Conclusion.Id]
			if found == false {
				err = errors.New(" *** Conclusion-Statement: " + arg.Conclusion.Id + "from Argument: " + arg.Id + "not found")
				return
			}
			edge.target = nodeId
			// if arg.Weight != nil {
			edge.edgeLabel = fmt.Sprintf("%.2f", arg.Weight)
			if len(edge.edgeLabel) > 4 {
				edge.edgeLabel = edge.edgeLabel[0:3]
			}
			// }
			if firstEdge {
				edges = []gmlEdge{edge}
				firstEdge = false
			} else {
				edges = append(edges, edge)
			}
		}
		// argument - - - - undercut
		// display undercutters only if they have supporting arguments
		if arg.Undercutter != nil && ag.Statements[arg.Undercutter.Id].Args != nil {
			nodeId, found := stat2Node[arg.Undercutter.Id]
			if found == false {
				err = errors.New(" *** Undercut-Statement: " + arg.Undercutter.Id + "from Argument: " + arg.Id + "not found")
				return
			}
			edge := newGmlEdge()
			edge.source = nodeId
			edge.target = nNode.id
			edge.lineType = dashed
			edge.width = mediumLine
			if firstEdge {
				edges = []gmlEdge{edge}
				firstEdge = false
			} else {
				edges = append(edges, edge)
			}
		}
	}
	// Issue <>
	for _, issue := range ag.Issues {
		nNode := newGmlNode()
		nNode.shapeType = hexagon
		s := ""
		switch issue.Standard {
		case 0:
			s = "PE"
		case 1:
			s = "CCE"
		case 2:
			s = "BRD"
		}
		nNode.nodeLabel = fmt.Sprintf("%s: %s", issue.Id, s)
		if firstNode {
			nodes = []gmlNode{nNode}
			firstNode = false
		} else {
			nodes = append(nodes, nNode)
		}
		// issue ---- positions
		for _, pos := range issue.Positions {
			nodeId, found := stat2Node[pos.Id]
			if found == false {
				err = errors.New(" *** Position-Statement: " + pos.Id + "from Issue: " + issue.Id + "not found")
				return
			}
			edge := newGmlEdge()
			edge.source = nodeId
			edge.target = nNode.id
			if firstEdge {
				edges = []gmlEdge{edge}
				firstEdge = false
			} else {
				edges = append(edges, edge)
			}
		}
	}

	return
}

func Export(w io.Writer, ag *caes.ArgGraph) error {
	nodes, edges, err := mkNodesAndEdges(ag)
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
