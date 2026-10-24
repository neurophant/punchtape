# refusal/bad-knob — Invalid knob values named with their requirement

run: 1

Each invalid --band or --threshold value is refused with the value and its requirement named on the line.


scope: src/**, Cargo.toml, Cargo.lock, build, materials/bordered.jpg

## row 0

command: ./brim detect materials/bordered.jpg --band 0.7
exit: 2
proof: green (4928f26464fa)

## row 1

command: ./brim detect materials/bordered.jpg --band 0
exit: 2
proof: green (aa7e1192e1fe)

## row 2

command: ./brim detect materials/bordered.jpg --band abc
exit: 2
proof: green (3c08e856903d)

## row 3

command: ./brim detect materials/bordered.jpg --threshold 1.5
exit: 2
proof: green (d853557265da)

## row 4

command: ./brim detect materials/bordered.jpg --threshold -0.5
exit: 2
proof: green (281d1f01d313)

## row 5

command: ./brim detect materials/bordered.jpg --threshold xyz
exit: 2
proof: green (3779ca76b1d0)

