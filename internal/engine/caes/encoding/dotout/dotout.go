// Copyright © 2015 The Carneades Authors
// This Source Code Form is subject to the terms of the
// Mozilla Public License, v. 2.0. If a copy of the MPL
// was not distributed with this file, You can obtain one
// at http://mozilla.org/MPL/2.0/.

// Visualizing CAES AG using [Dot](http://www.graphviz.org/Documentation.php)
package dotout

import (
	"errors"
	"fmt"
	"github.com/carneades/carneades-4/internal/engine/caes"
	"io"
)

const (
	black  = "#000000"
	red    = "#FF0000"
	green  = "#3AB54A"
	yellow = "#FCEE21"

	white = "#008000"
	// line type
	line   = "line"
	dashed = "dashed"
	//
	thinLine   = "1.0"
	mediumLine = "2.0"
	boldLine   = "3.0"
	//
	noArrow   = "none"
	withArrow = "standard"
	// font
	font = "Dialog"
	// shapeType
	// diamond   = "diamond"
	rectangle = "rectangle"
	// ellipse   = "ellipse"
	roundrectangle = "roundrectangle"
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
		"<key attr.name=\"description\" attr.type=\"string\" for=\"node\" id=\"d5\"/>",
		"<key for=\"node\" id=\"d6\" yfiles.type=\"nodegraphics\"/>",
		"<key attr.name=\"description\" attr.type=\"string\" for=\"edge\" id=\"d9\"/>",
		"<key for=\"edge\" id=\"d10\" yfiles.type=\"edgegraphics\"/>")
	graphNr++
}

func pFoot(w io.Writer) {
	p(w, "</graphml>")
}

func calculateWidth(l int) string {
	width := "30.0"
	switch l {
	case 0, 1, 2, 3:
	case 4, 5, 6:
		width = "60.0"
		// dheight = "45.0"
	case 7, 8, 9:
		width = "80.0"
		// dheight = "67,5"
	case 10, 11, 12:
		width = "100.0"
		// dheight = "90"
	case 13, 14, 15:
		width = "120.0"
		// dheight = "112,5"
	case 16, 17, 18:
		width = "140"
	default:
		width = "200.0"
		// dheight = "150.0"
	}
	return width

}

func pNodes(w io.Writer, nodes []gmlNode) {

	for _, node := range nodes {
		p(w, "   <node id=\""+node.id+"\">",
			"      <data key=\"d5\"/>",
			"      <data key=\"d6\">",
			"      <y:ShapeNode>")

		height := "30.0"
		width := calculateWidth(len(node.nodeLabel))

		p(w, "         <y:Geometry height=\""+height+"\" width= \""+width+"\" />")
		if node.color == "" {
			p(w, "         <y:Fill hasColor=\"false\" transparent=\"false\"/>")
		} else {
			p(w, "         <y:Fill color=\""+node.color+"\" transparent=\"false\"/>")
		}

		p(w, "         <y:BorderStyle color=\"#000000\" type=\""+
			node.borderLine+
			"\" width=\""+
			node.borderWidth+
			"\"/>")

		if node.underlinedLabel {
			p(w, "      <y:NodeLabel fontFamily=\""+font+"\" underlinedText=\"true\">"+node.nodeLabel+"</y:NodeLabel>")
		} else {
			p(w, "      <y:NodeLabel fontFamily=\""+font+"\" >"+node.nodeLabel+"</y:NodeLabel>")
		}
		p(w, "       <y:Shape type=\""+node.shapeType+"\"/>",
			"       </y:ShapeNode>",
			"       </data>",
			"   </node>")
	}
}

func pEdges(w io.Writer, edges []gmlEdge) {

	for _, edge := range edges {

		p(w, "   <edge id=\""+
			edge.id+
			"\" source=\""+
			edge.source+
			"\" target=\""+
			edge.target+"\">")

		p(w, "      <data key=\"d9\"/>",
			"      <data key=\"d10\">",
			"         <y:PolyLineEdge>")
		p(w, "            <y:LineStyle color=\"#000000\" type=\""+
			edge.lineType+"\" width=\""+edge.width+"\"/>")
		p(w, "            <y:Arrows source=\"none\" target=\""+
			edge.lineTarget+"\"/>")
		if edge.edgeLabel != "" {
			p(w, "		        <y:EdgeLabel>"+edge.edgeLabel+"</y:EdgeLabel>")
		}

		p(w, "         </y:PolyLineEdge>",
			"      </data>",
			"   </edge>")
	}
}

func mkNodesAndEdges(ag caes.ArgGraph) (nodes []gmlNode, edges []gmlEdge, err error) {
	stat2Node := make(map[string]string)
	firstNode := true
	firstEdge := true
	// Statements []
	for _, stat := range ag.Statements {
		nNode := newGmlNode() // shapeType [] (rectangle)
		stat2Node[stat.Id] = nNode.id
		nNode.nodeLabel = stat.Id
		if stat.Assumed {
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
			edge.edgeLabel = fmt.Sprintf("%v", arg.Weight)
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
		if arg.Undercutter != nil {
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
			s = "DV"
		case 1:
			s = "PE"
		case 2:
			s = "CCE"
		case 3:
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

func Export(w io.Writer, ag caes.ArgGraph) error {
	pHead(w)
	p(w, "<graph id=\"G"+
		fmt.Sprintf("%d", graphNr)+"\">")
	graphNr++
	nodes, edges, err := mkNodesAndEdges(ag)
	if err != nil {
		return err
	}
	pNodes(w, nodes)
	pEdges(w, edges)
	p(w, "</graph>")
	pFoot(w)
	return nil
}
