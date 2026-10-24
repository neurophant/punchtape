# detect/frame — Frame margins found on the bordered test picture

run: 1

brim detect on a JPEG with a thin even frame answers the four margin depths in pixels (top bottom left right); the band knob caps the search depth and shallower bands cut the answer down, the threshold knob steers how much disorder still counts as even.


scope: src/**, Cargo.toml, Cargo.lock, build, materials/bordered.jpg

## row 0

command: ./brim detect materials/bordered.jpg
exit: 0
stdout: 4 4 4 4

proof: green (600a9c772296)

## row 1

command: ./brim detect materials/bordered.jpg --band 0.01
exit: 0
stdout: 2 2 2 2

proof: green (d6faa512f86c)

## row 2

command: ./brim detect materials/bordered.jpg --band 0.05
exit: 0
stdout: 4 4 4 4

proof: green (656db091697e)

## row 3

command: ./brim detect materials/bordered.jpg --band 0.5
exit: 0
stdout: 4 4 4 4

proof: green (a2749a7d389e)

## row 4

command: ./brim detect materials/bordered.jpg --threshold 0.7
exit: 0
stdout: 4 4 4 4

proof: green (5d9fb870a046)

## row 5

command: ./brim detect materials/bordered.jpg --threshold 0.9
exit: 0
stdout: 62 62 62 62

proof: green (75172c9f53f8)

