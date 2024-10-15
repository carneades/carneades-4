# Carneades User Guide

## YAML Input Format

The YAML based input format is the preferred input for Carneades 4.x.
See also the numerous examples in [[./examples/AGs/YAML/]].
The file specifies the argumentation system as an object at the top level.

### Metainformation

At the toplevel the format has an optional `meta` object which has no semantic meaning but is commonly used to hold structured meta information like a `title`, `note`, `description`, or references to `source` documents. 

### Language

For readability of the outputs of Carneades you can specify verbalisations of the predicates used in the statements.
This is done by adding an optional `language` object at the top level.
Each field in this language object should have its name following the scheme `PRED/ARITY` where `PRED` is the name of the predicate and `ARITY` specified its arity.
The value is then a Go formatting string accepting the number of arguments matching the `ARITY` which inserts the predicate arguments into a human readable verbalisation used in the output graph.

For example a predicate expressing explanation could be specified by a field `explanation/2: "Theory %s explains %s."`.

To give names for ground atomic formulas the optional field `statements` can be given an object where each field is a ground atomic formula and the value is a string representing its verbalisation.

For example such a field could be `forbidden(cannabis): Cannabis consumption is illegal.`.

### Arguments

A list `arguments` can be used to specify concrete arguments in the graph however Carneades can also generate arguments from schemes.

Each argument is an object with the following fields:

  - `conclusion`: a ground atomic formula.
  - `premises`: a list of ground atomic formulas.
  
An argument can optionally have the following fields:
  - `scheme`: reference to a scheme via the scheme's `id` to inherit properties specified by the scheme. When a scheme is specified it can also be used to infer conclusion or premises.
  - `undercutter`: a ground atomic formula.

#### Generating Arguments

Apart from specifying concrete arguments you can also specify a list `argument_schemes` which then generate concrete instantiated arguments.
Each scheme needs to specify the following:

  - `id`: a unique identifier of the argument scheme.
  - `variables`: a list of variables occuring in the atomic formulas of scheme. 
    Only required when the premises and conclusions contain variables.
    Keeping with Prolog syntax they should be starting with an upper case letter. 
  - `premises`: a list of atomic formulas acting as premises for the scheme.
  - `conclusions`: a list of atomic formulas acting as the conclusion of the scheme.

Optionally a scheme can also contain:
  
  - `meta`: a metadata object without semantic relevance.
  - `weight`: a weighing function (see below for details).
  - `roles`: a list of role names for the corresponding `premises`.
  - `assumptions`: a list of atomic formulas.
  - `exceptions`: a list of atomic formulas.

Note that argument schemes in Carneades can go beyond their argumentation theoretic capabilities exposing the full capability of the underlying CHR solver through use of the `deletions` and `guards` fields.

#### Weighing Functions

Weighing functions enable many interesting specification capabilities:

  - `linked`: The default weighing function putting a weight of `0.0` if any premise is not labelled `in`, `1.0` otherwise.
  - `convergent`: Puts weight `1.0` if any premise is labelled `in`, `0.0` otherwise.
  - `cumulative`: Puts the fraction of `in`-labelled premises as weight.
  - `factorized`: Similar to `linked` but compares the number of premises of other arguments participating in the same issue and ranking those higher with most premises. 
  
Additionaly Carneades supports means of specifying custom weighing functions:
  
  - `constant: FLOAT`: Specifying a floating point number represents a constant weighing function that behaves like `linked` but outputs the given factor instead of `1.0`. 
  - `criteria`: A criteria weighing function is an object with a list of `hard` role names and an object `soft` specifying how the soft constraints schould be accumulated. 
    Hard constraints are treated like `linked` premises and soft contraints weaken the acceptance according to specification.
    TODO: document soft constraint specification
  - `preference`: Preference weighing function (like linked but orders arguments in the same issue via preference ordering)

TODO: fully document custom functions, named custom functions
  
### Issues

In the field `issues` contains an object where each field is the id of an issue object.
An issue object has the following fields:

  - `positions`: a list of ground atomic formulas participating in the issue.
  - `standard`: optionally a proof standard can be specified.

#### Generating Issues

Like the arguments you can also specify an object `issue_schemes` to automatically instantiate issues.
Each field in this object is the id of an issue scheme and contains a list of atomic formulas that will instantiate the positions of the resulting issues.
There is also an enumeration like feature for specifying e.g. that a property can only hold for one individual by putting `...` between two list items sharing similar structure.
You can see this e.g. in the car buying examle where it is used to specify that only a single purchase option can be chosen.
This can be especially useful for axiomatising certain predicates

#### Proof Standards

Proof standards are used to decide the winning statement of an issue.
Note that in cases where there is no clear winner, none of the statements will get labelled `in`.

Built in proof standards are:
 
  - `PE` (Preponderance of the Evidence): The default proof standard picking the strongest incoming argument by weight.
  - `CCE` (Clear and Convincing Evidence): Like `PE` but the weight of the winning argument needs to be sufficiently larger than the other argument.
  - `BRD` (Beyond Reasonable Doubt): Like `CCE` but in addition the losing arguments need to have sufficiently low weights.

### Assumptions

In the field `assumptions` a list of ground atomic formulas can be specified that are to be taken as true.

### Tests

To specify the expected result of an argumentation system the field `tests` can contain an object with two fields `in` and `out` each containing a list of ground atomic formulas that are supposed to get labeled `in` or `out`, respectively. 

