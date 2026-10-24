# brim/detect — the four-number answer on the materials

run: 2

brim detect on a picture prints four numbers — top, bottom, left, right — the even margin found from each edge. On the framed material all four sides answer 4, on the clear material all four answer 0, and a mid-range sensitivity keeps the framed answer.

scope: src/**, Makefile, materials/bordered.jpg, materials/clear.jpg

## row 0

command: ./brim detect materials/bordered.jpg
exit: 0
stdout: 4 4 4 4

proof: green (bb695818a34a)

## row 1

command: ./brim detect materials/clear.jpg
exit: 0
stdout: 0 0 0 0

proof: green (8f459d78466a)

## row 2

command: ./brim detect materials/bordered.jpg --threshold 0.3
exit: 0
stdout: 4 4 4 4

proof: green (96f1a6ee8847)

## row 3

command: ./brim detect materials/bordered.jpg --threshold 0.7
exit: 0
stdout: 4 4 4 4

proof: green (79363a23ba9e)

