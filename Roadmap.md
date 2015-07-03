
# v4.0

- Dung abstract argumentation frameworks
- Semantics: grounded, complete, preferred, stable
- import: Trivial Graph Format (TGF)
- export: GraphML

# v4.1

- Structured argumentation
- Generalization of Carneades to handle:
    * Cycles, via a mapping to Dung AFs
    * Cumulative arguments
    * Issue-based Information Systems (IBIS)
    * Multicriteria Decision-Making (MCDM)
- Import
    * Carneades Argument Format (CAF), extended to handle the new model
    * Argument Interchange Format (AIF), JSON serialization
    * Simple, plain text format, based on YAML
- Export
    * GraphML
    * YAML/Markdown

# v4.2

- Structured survey tool, similar to Parmenides
- Based on design in: Gordon, T. F. (2013). Structured Consultation
  with Argument Graphs. In K. Atkinson, H. Prakken & A. Wyner (ed.),
  From Knowledge Representation to Argumentation in AI, Law and Policy
  Making (pp. 115–134). College Publications.

# v4.3

- Static website generator, for browsing argument graphs
- Similar to Jeykll <http://jekyllrb.com/> and Hugo <http://gohugo.io/>
- Views: hypertext and argument maps (diagrams)
- No Carneades Web service required
- Static pages may be uploaded to any Web server, e.g. via FTP

# v4.4

- Argumentation schemes, represented in Datalog
- Argument construction, via a Datalog inference engine
- Argument validation, by matching arguments to schemes

