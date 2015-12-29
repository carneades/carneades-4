:- use_module(library(chr)).
:- chr_constraint bar/0, foo/0, go/0.

r1 @ go ==> foo.
r2 @ go ==> bar.

