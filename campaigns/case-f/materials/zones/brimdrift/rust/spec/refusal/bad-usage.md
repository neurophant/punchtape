# refusal/bad-usage — Clumsy invocation refused honestly

run: 2

A missing image argument, an extra file argument, an unknown command, and an unknown option are each refused with a nonzero exit and the usage line.


scope: src/**, Cargo.toml, Cargo.lock, build, materials/bordered.jpg

## row 0

command: ./brim detect
exit: 2
proof: green (f0c68365f71d)

## row 1

command: ./brim detect materials/bordered.jpg extra.jpg
exit: 2
proof: green (fe50c14589cf)

## row 2

command: ./brim scan materials/bordered.jpg
exit: 2
proof: green (23258d76db53)

## row 3

command: ./brim detect materials/bordered.jpg --wat
exit: 2
proof: green (a2def1651bad)

## row 4

command: ./brim
exit: 2
proof: green (f5e9864b7af0)

