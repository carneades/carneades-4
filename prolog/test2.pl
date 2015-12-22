:- use_module(library(http/json)).
:- use_module(library(http/json_convert)).
:- json_object argument(scheme:text, parameters:list(text)).
:- initialization main.

terms_strings([],[]).
terms_strings([H|T],[SH|ST]) :-
    term_string(H,SH),
    terms_strings(T,ST).

main :-
  terms_strings([rain(berlin)], L),
  prolog_to_json(argument(expert_witness, L), J),
  json_write(current_output, J), 
  nl,
  halt(0).


