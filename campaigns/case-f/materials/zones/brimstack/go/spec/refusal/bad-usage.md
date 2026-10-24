# refusal/bad-usage — Clumsy invocation refused honestly

run: 1

A missing image argument, an extra file argument, an unknown command, and an unknown option are each refused with a nonzero exit and the usage line.


scope: src/**, Cargo.toml, Cargo.lock, build, materials/bordered.jpg

## row 0

command: ./brim detect
exit: 2
proof: green (8e44725f9199)

## row 1

command: ./brim detect materials/bordered.jpg extra.jpg
exit: 2
proof: green (f8480e3240e7)

## row 2

command: ./brim scan materials/bordered.jpg
exit: 2
proof: green (ff87eabe47db)

## row 3

command: ./brim detect materials/bordered.jpg --wat
exit: 2
proof: green (e2fb05cf5241)

## row 4

command: ./brim
exit: 2
proof: green (c133946420e4)

