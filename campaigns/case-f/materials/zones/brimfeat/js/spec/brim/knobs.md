# brim/knobs — the two knobs with valid values

run: 1

--band sets how deep from the edge to search, a fraction of each side no greater than half; --threshold sets the sensitivity, a fraction from 0 to 1. Evenness stretching through the whole band makes the whole band the answer (uniform.jpg: bands of 15 rows and 20 columns at the default quarter), and each side is measured on its own (asym.jpg: frame of 4 top, 7 bottom, 3 left, 5 right).


scope: brim, package.json, src/**, material/**, fixtures/**, node_modules/**

## row 0

command: ./brim detect material/bordered.jpg --band 0.5
exit: 0
stdout: 4 4 4 4

proof: green (a6a56a306966)

## row 1

command: ./brim detect material/bordered.jpg --threshold 0.3
exit: 0
stdout: 4 4 4 4

proof: green (d508a6495e1e)

## row 2

command: ./brim detect material/bordered.jpg --threshold 0.7
exit: 0
stdout: 4 4 4 4

proof: green (c5c33faa3f51)

## row 3

command: ./brim detect fixtures/uniform.jpg
exit: 0
stdout: 15 15 20 20

proof: green (8028f9d32b3c)

## row 4

command: ./brim detect fixtures/uniform.jpg --band 0.5
exit: 0
stdout: 30 30 40 40

proof: green (74b5d378c40d)

## row 5

command: ./brim detect fixtures/asym.jpg
exit: 0
stdout: 4 7 3 5

proof: green (cd384f69970a)

