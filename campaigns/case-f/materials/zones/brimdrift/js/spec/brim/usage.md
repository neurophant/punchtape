# brim/usage — honest refusal of clumsy invocation

run: 1

A bad knob value, an extra file argument, an unknown option, an unknown command, and a missing picture are refused with the offending value named along with its requirement, plus the usage line, on exit code 2.


scope: brim, package.json, src/**, material/**, node_modules/**

## row 0

command: ./brim
exit: 2
stdout: contains: missing command
proof: green (432b612e3cc6)

## row 1

command: ./brim detect
exit: 2
stdout: contains: no picture given
proof: green (256dd5d0af41)

## row 2

command: ./brim detect material/bordered.jpg extra.jpg
exit: 2
stdout: contains: extra file argument 'extra.jpg'
proof: green (21537f0252a3)

## row 3

command: ./brim detect material/bordered.jpg --band 0.7
exit: 2
stdout: contains: bad --band value '0.7': a fraction of the side, greater than 0 and at most 0.5
proof: green (591d0a35ee45)

## row 4

command: ./brim detect material/bordered.jpg --band abc
exit: 2
stdout: contains: bad --band value 'abc': a fraction of the side, greater than 0 and at most 0.5
proof: green (ede1c6274706)

## row 5

command: ./brim detect material/bordered.jpg --threshold 1.5
exit: 2
stdout: contains: bad --threshold value '1.5': a fraction from 0 to 1
proof: green (8be97dff0ef6)

## row 6

command: ./brim detect material/bordered.jpg --wat
exit: 2
stdout: contains: unknown option '--wat'
proof: green (02b349e073c6)

## row 7

command: ./brim frobnicate x.jpg
exit: 2
stdout: contains: unknown command 'frobnicate'
proof: green (609f3506b908)

