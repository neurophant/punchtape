# brim/knobs — the two knobs with valid values

run: 1

--band sets how deep from the edge to search, a fraction of each side no greater than half; --threshold sets the sensitivity, a fraction from 0 to 1. Evenness stretching through the whole band makes the whole band the answer (uniform.jpg: bands of 15 rows and 20 columns at the default quarter), and each side is measured on its own (asym.jpg: frame of 4 top, 7 bottom, 3 left, 5 right).


scope: brim, package.json, src/**, material/**, fixtures/**, node_modules/**

## row 0

command: ./brim detect material/bordered.jpg --band 0.5
exit: 0
stdout: 4 4 4 4

proof: green (c012e4ffcd43)

## row 1

command: ./brim detect material/bordered.jpg --threshold 0.3
exit: 0
stdout: 4 4 4 4

proof: green (186fd55b584d)

## row 2

command: ./brim detect material/bordered.jpg --threshold 0.7
exit: 0
stdout: 4 4 4 4

proof: green (da960392e38c)

## row 3

command: ./brim detect fixtures/uniform.jpg
exit: 0
stdout: 15 15 20 20

proof: green (42bd83d9a2e3)

## row 4

command: ./brim detect fixtures/uniform.jpg --band 0.5
exit: 0
stdout: 30 30 40 40

proof: green (2fc0f5fbcc81)

## row 5

command: ./brim detect fixtures/asym.jpg
exit: 0
stdout: 4 7 3 5

proof: green (7f90849e0260)

