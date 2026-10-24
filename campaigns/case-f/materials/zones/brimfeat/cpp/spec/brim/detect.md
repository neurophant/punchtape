# brim/detect — the four-number answer on the materials

run: 1

brim detect on a picture prints four numbers — top, bottom, left, right — the even margin found from each edge. On the framed material all four sides answer 4, on the clear material all four answer 0, and a mid-range sensitivity keeps the framed answer.

scope: src/**, Makefile, materials/bordered.jpg, materials/clear.jpg

## row 0

command: ./brim detect materials/bordered.jpg
exit: 0
stdout: 4 4 4 4

proof: green (124cb3e5d6a6)

## row 1

command: ./brim detect materials/clear.jpg
exit: 0
stdout: 0 0 0 0

proof: green (e21a5986e6bd)

## row 2

command: ./brim detect materials/bordered.jpg --threshold 0.3
exit: 0
stdout: 4 4 4 4

proof: green (50bd92341daa)

## row 3

command: ./brim detect materials/bordered.jpg --threshold 0.7
exit: 0
stdout: 4 4 4 4

proof: green (828346efa8e2)

