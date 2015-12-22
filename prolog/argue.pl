#!/usr/bin/swipl -f -q

% Reads a CHR theory and query from the standard input and
% writes records in JSON format describing the arguments
% constructed to the standard output.

:- use_module(library(chr)).
:- use_module(arg2json).
:- initialization main.

main :-
    % todo
    consult(user),  % load from stdin
    facts,
    halt(0).

