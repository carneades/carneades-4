
# v4.0

- Dung abstract argumentation frameworks
- Semantics: grounded, complete, preferred, stable
- Import: Trivial Graph Format (TGF)
- Export: GraphML

# v4.1

- Structured argumentation
- Generalization of Carneades to handle:
    * Cycles
    * Cumulative arguments
    * Issue-based Information Systems (IBIS)
    * Multicriteria Decision-Making (MCDM)
- Import
    * Simple, plain text format, based on YAML
    * The Legal Knowledge Interchange Format (LKIF), used by Carneades 2
    * The Carneades Argument Format (CAF), used by Carneades 3
    * The JSON serialization of the Argument Interchange Format (AIF)
- Export
    * GraphML
    * YAML
    * DOT (GraphViz)

# v4.2

- user-definable argument evaluation functions, for deriving relative argument weights from properties of the arguments (4.1 includes some builtin argument evaluation functions)
- Argument graph upload to CoachDB repositories
- Argument graph search and retrieval

# v4.3

- Language for representing argumentation schemes, based on Datalog
- Argument construction, via a Datalog inference engine
- Argument validation, by matching arguments to schemes
- Goal selection, based on: Ballnat, S. and Gordon, T.F. Goal Selection in Argumentation Processes — A Formal Model of Abduction in Argument Evaluation Structures. Computational Models of Argument – Proceedings of COMMA 2010, IOS Press (2010), 51–62.

# v4.4

- Structured survey tool, similar to Parmenides
- Based on design in: Gordon, T. F. (2013). Structured Consultation
  with Argument Graphs. In K. Atkinson, H. Prakken & A. Wyner (ed.),
  From Knowledge Representation to Argumentation in AI, Law and Policy
  Making (pp. 115–134). College Publications.

# v4.5

- Static website generator, for browsing argument graphs
- Views: hypertext and argument maps in SVG
- No Carneades web services will be required
- Static pages may be uploaded to any Web server, e.g. via FTP




