#!/usr/bin/swipl -f -q

:- use_module(library(chr)).
:- use_module(library(http/json)).
:- use_module(library(http/json_convert)).
:- chr_constraint argument/2.
:- json_object argument(scheme:text, parameters:list(text)).
:- initialization main.

terms_strings([],[]).
terms_strings([H|T],[SH|ST]) :-
    term_string(H,SH),
    terms_strings(T,ST).

argument(I,P) <=> 
  term_string(I,S),
  terms_strings(P,L),
  prolog_to_json(argument(S,L),J), 
  json_write(current_output, J), 
  nl | 
  true.

main :-
    assumptions,
    halt(0).

%%%  Example  

assumptions :- argument(expert_witness,[joe,made_of(moon,green_cheese)]).

