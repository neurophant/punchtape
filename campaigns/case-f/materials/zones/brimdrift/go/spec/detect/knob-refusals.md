# detect/knob-refusals — a clumsy invocation is an honest refusal naming the knob and its requirement

run: 2

A bad knob value, a knob without its value, an unknown option, an extra file argument, or no picture at all is refused with exit code 2 and a brief English message on stderr; each invalid knob value is named along with what the knob requires: --band wants a fraction of the side greater than 0 and at most 0.5, --threshold wants a fraction from 0 to 1.

scope: main.go, picture.go, testdata/bordered.jpg, testdata/clear.jpg

## row 0

command: ./brim
exit: 2
proof: green (7847365a60c8)

## row 1

command: ./brim detect
exit: 2
proof: green (81f6541edeec)

## row 2

command: ./brim detect --band abc testdata/bordered.jpg
exit: 2
proof: green (eac8af9bec2e)

## row 3

command: ./brim detect --band 0 testdata/bordered.jpg
exit: 2
proof: green (ce8c9bce1e3a)

## row 4

command: ./brim detect --band 0.6 testdata/bordered.jpg
exit: 2
proof: green (14ac2b7b072b)

## row 5

command: ./brim detect --threshold 1.5 testdata/bordered.jpg
exit: 2
proof: green (a12f364ca170)

## row 6

command: ./brim detect --wibble testdata/bordered.jpg
exit: 2
proof: green (fc13c8bd1384)

## row 7

command: ./brim detect testdata/bordered.jpg testdata/clear.jpg
exit: 2
proof: green (3aab27623271)

