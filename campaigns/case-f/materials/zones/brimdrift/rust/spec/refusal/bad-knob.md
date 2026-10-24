# refusal/bad-knob — Invalid knob values named with their requirement

run: 2

Each invalid --band or --threshold value is refused with the value and its requirement named on the line.


scope: src/**, Cargo.toml, Cargo.lock, build, materials/bordered.jpg

## row 0

command: ./brim detect materials/bordered.jpg --band 0.7
exit: 2
proof: green (256d6773286c)

## row 1

command: ./brim detect materials/bordered.jpg --band 0
exit: 2
proof: green (db3dda81c1ee)

## row 2

command: ./brim detect materials/bordered.jpg --band abc
exit: 2
proof: green (4803b5ff7094)

## row 3

command: ./brim detect materials/bordered.jpg --threshold 1.5
exit: 2
proof: green (345479c56176)

## row 4

command: ./brim detect materials/bordered.jpg --threshold -0.5
exit: 2
proof: green (6ad06cafc091)

## row 5

command: ./brim detect materials/bordered.jpg --threshold xyz
exit: 2
proof: green (59d762a2e441)

