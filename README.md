
These are the source files of Version 4 of the
Carneades argumentation system, written in the Go programming language.

This source code is subject to the terms of the Mozilla Public
License, version 2.0 (MPL-2.0). If a copy of the MPL was not
distributed with this software, it is also available online at
<http://mozilla.org/MPL/2.0/>.  For futher information about the MPL see <http://www.mozilla.org/MPL/2.0/FAQ.html>.

This version of Carneades consists of:

- An implementation of a solver for Dung abstract argumentation frameworks,
using grounded, complete, preferred and stable semantics. Argumentation Frameworks
can be represented using the [Trivial Graph Format](https://en.wikipedia.org/wiki/Trivial_Graph_Format). The computed extensions can be exported to DOT, GraphML and plain text.
- An evaluator for structured arguments, based on a new version of the 
Carneades Argument Evaluation Structures (CAES) formal model of argument. 
- New features of version 4 of Carneades include:
    * Support for cyclic argument graphs, cumulative arguments, issue-based information systems (IBIS) and multicriteria decision analysis.
    * Argument graphs can be represented in AGXML, AIF, LKIF, CAF and YAML and exported to DOT, GraphML, PNG, SVG and YAML.
    * User-definable argument weighing functions, for computing the relative weights of arguments based on their properties, such as the authority or effective date of the argumentation scheme applied, or the labels (in, out, undecided) of premises.  The failure of a premise can weaken or even strengthen an argument without defeating it entirely.
    * Automatic argument construction by applying argumentation schemes to assumptions, via an inference engine implemented using [Constraint Handling Rules](https://dtai.cs.kuleuven.be/CHR/). 

Carneades can visualize both Dung abstract argumentation frameworks and CAES argument graphs using [DOT](http://www.graphviz.org/content/dot-language) and [GraphML](https://en.wikipedia.org/wiki/GraphML).  We recommend using the free [yEd](http://www.yworks.com/yed) GraphML editor to view the GraphML files.


Please see the Carneades blog at <https://carneades.github.io/> for
announcements about the further development of this version.

See the INSTALL.md file in the same directory as this README for
installation instructions.

