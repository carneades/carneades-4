notes.md
========

- All test_cases1/tgf/rdm20_*.tgf tests terminate quickly with correct results!

- Most of the tests have a single extension, often empty

Exception: test_cases1/solutions/rdm200_35.EE-PR has two preferred extensions:

[[a89,a141,a64,a142,a101,a50,a71,a176,a182,a167,a34,a2,a125,a124,a111,a58,a29,a66,a57,a83,a129,a54],[a89,a141,a64,a142,a101,a50,a71,a176,a154,a152,a182,a167,a34,a2,a125,a124,a111,a58,a29,a66,a57,a129,a54]]

# Performance (iMac, dual-core)

$ time ./carneades -p EE-PR3 -f ../test_cases1/tgf/rdm20_0.tgf 
[[]]

real	0m8.470s
user	0m8.935s
sys	0m0.359s

$ time ../Tweety/solvers/tweetysolver-v1.1.1.sh -p EE-PR -fo tgf  -f ../test_cases1/tgf/rdm20_0.tgf 
[[]]

real	0m0.419s
user	0m0.528s
sys	0m0.063s
