# brim/band — the search band knob

run: 2

The --band knob bounds how deep from the edge the margin is hunted: a band shallower than the frame cuts the found numbers down (0.01 answers 2), a deeper one leaves them (0.05 answers 4).

scope: src/**, Makefile, materials/bordered.jpg

## row 0

command: ./brim detect materials/bordered.jpg --band 0.01
exit: 0
stdout: 2 2 2 2

proof: green (6ca64e3e05bd)

## row 1

command: ./brim detect materials/bordered.jpg --band 0.05
exit: 0
stdout: 4 4 4 4

proof: green (31034e074103)

