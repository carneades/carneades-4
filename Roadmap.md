
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
    * Multi-criteria Decision Analysis (MCDA)
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
- Language for representing argumentation schemes, based on Constraint Handling Rules (CHR)
- Argument construction, via a CHR inference engine

# v4.3
- CRUD (Create, Read, Update and Delete) operations for managing argument graphs using the Open Source CoachDB document database system 
- Create and upload argument graphs to a CoachDB database
- Search for, retreive and read (view) argument graphs in the database
- Update (edit) argument graphs in the database
- Delete argument graphs from the database
- Web user interfaces will be provided for all these database operations

# v4.4
- Argument validation, by matching arguments to schemes
- Goal selection, using abduction, continuing our prior work: Ballnat, S. and Gordon, T.F. Goal Selection in Argumentation Processes — A Formal Model of Abduction in Argument Evaluation Structures. Computational Models of Argument – Proceedings of COMMA 2010, IOS Press (2010), 51–62.

# v4.5

- Interactive, web-based argument browser and structured survey tool
- The browser will provide "guided tours" of argument graphs, to help users to understand the arguments, taking into    consideration their interests and assumptions or beliefs.
- The structured survey tool is inspired by Parmenides: K Atkinson, T Bench-Capon, P McBurney, 
  "PARMENIDES: facilitating deliberation in democracies", Artificial Intelligence and Law, 2006,
  and will be based on the design in: Gordon, T. F. (2013). Structured Consultation
  with Argument Graphs. In K. Atkinson, H. Prakken & A. Wyner (ed.),
  From Knowledge Representation to Argumentation in AI, Law and Policy
  Making (pp. 115–134). College Publications.






