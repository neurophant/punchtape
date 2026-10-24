# detect/answer — brim detect answers with the four margin numbers

run: 1

Given a JPEG picture the tool prints one line of four numbers — margin pixels on the top, bottom, left, and right, each side measured on its own. On the attached bordered.jpg (an even 4-pixel frame) it says 4 4 4 4; on clear.jpg (no margin) it says 0 0 0 0. A search band shallower than the frame cuts the numbers down to the band depth: --band 0.01 gives 2 2 2 2 on the 250-pixel side, because evenness stretches to the end of the band and the whole band is the answer. A deeper one leaves the numbers at the frame: --band 0.05 gives 4 4 4 4, and the stated ceiling --band 0.5 (half the side, no more is allowed) works the same. The threshold knob works across its whole valid range: 0.5 given explicitly matches the default; 0, the least aggressive bar, lets evenness stretch through the whole default band, so the answer is the whole band — 62 pixels, a quarter of the 250-pixel side, and on the non-square clear.jpg a fraction of each side on its own: 87 top and bottom of the 350-pixel height, 127 left and right of the 510-pixel width; 1, the most aggressive, leaves nothing even enough, so no margin is found and the answer is zeros. A repeat run on the same picture prints the same line to the letter.

scope: main.go, picture.go, entropy.go, margin.go, go.mod, testdata/bordered.jpg, testdata/clear.jpg

## row 0

command: ./brim detect testdata/bordered.jpg
exit: 0
stdout: 4 4 4 4

proof: green (7c6c73fdab04)

## row 1

command: ./brim detect testdata/clear.jpg
exit: 0
stdout: 0 0 0 0

proof: green (e022f4f8a324)

## row 2

command: ./brim detect testdata/bordered.jpg --band 0.01
exit: 0
stdout: 2 2 2 2

proof: green (c908512ee91c)

## row 3

command: ./brim detect testdata/bordered.jpg --band 0.05
exit: 0
stdout: 4 4 4 4

proof: green (b29086a5bde7)

## row 4

command: a=$(./brim detect testdata/bordered.jpg) && b=$(./brim detect testdata/bordered.jpg) && [ "$a" = "$b" ] && printf '%s\n' "$a"
exit: 0
stdout: 4 4 4 4

proof: green (7b0631bb8580)

## row 5

command: ./brim detect testdata/bordered.jpg --threshold 0.5
exit: 0
stdout: 4 4 4 4

proof: green (26625a03ef74)

## row 6

command: ./brim detect testdata/bordered.jpg --threshold 0
exit: 0
stdout: 62 62 62 62

proof: green (6875ec5fc357)

## row 7

command: ./brim detect testdata/bordered.jpg --threshold 1
exit: 0
stdout: 0 0 0 0

proof: green (e368271063f5)

## row 8

command: ./brim detect testdata/clear.jpg --threshold 0
exit: 0
stdout: 87 87 127 127

proof: green (6fd493de54cf)

## row 9

command: ./brim detect testdata/bordered.jpg --band 0.5
exit: 0
stdout: 4 4 4 4

proof: green (238539eb0717)

