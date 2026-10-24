# brim/usage — honest refusal of clumsy invocation

run: 1

A bad knob value, an extra file argument, an unknown option, an unknown command, and a missing picture are refused with the offending value named along with its requirement, plus the usage line, on exit code 2.


scope: brim, package.json, src/**, material/**, node_modules/**

## row 0

command: ./brim
exit: 2
stdout: contains: missing command
proof: green (caa2182c52d0)

## row 1

command: ./brim detect
exit: 2
stdout: contains: no picture given
proof: green (41d308bbd6c1)

## row 2

command: ./brim detect material/bordered.jpg extra.jpg
exit: 2
stdout: contains: extra file argument 'extra.jpg'
proof: green (868cac94accc)

## row 3

command: ./brim detect material/bordered.jpg --band 0.7
exit: 2
stdout: contains: bad --band value '0.7': a fraction of the side, greater than 0 and at most 0.5
proof: green (94e79164dfd8)

## row 4

command: ./brim detect material/bordered.jpg --band abc
exit: 2
stdout: contains: bad --band value 'abc': a fraction of the side, greater than 0 and at most 0.5
proof: green (e1b8edc0942e)

## row 5

command: ./brim detect material/bordered.jpg --threshold 1.5
exit: 2
stdout: contains: bad --threshold value '1.5': a fraction from 0 to 1
proof: green (ef675f6c097f)

## row 6

command: ./brim detect material/bordered.jpg --wat
exit: 2
stdout: contains: unknown option '--wat'
proof: green (a19a141096a2)

## row 7

command: ./brim frobnicate x.jpg
exit: 2
stdout: contains: unknown command 'frobnicate'
proof: green (ad8501e88d9e)

