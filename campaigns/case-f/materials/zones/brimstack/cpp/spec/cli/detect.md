# cli/detect — measure the four margins

run: 1

brim detect picture.jpg prints four numbers — margin pixels on the top, bottom, left, and right. Each side is walked from the edge inward one pixel line at a time; a line is a chunk, and its histogram entropy normalized to [0,1] is the chunk's chaos. The margin ends where the entropy first stops being low (above --threshold); the boundary sits at the largest entropy jump between neighboring chunks. Evenness stretching to the end of the search band answers the whole band; no even strip at the edge answers zero. The program only reads the picture and prints the answer: same picture, same answer, every run.


scope: brim, material/**

## row 0

command: ./brim detect material/bordered.jpg
exit: 0
stdout: contains: 4 4 4 4
proof: green (e9e55f1be3ca)

## row 1

command: ./brim detect material/clear.jpg
exit: 0
stdout: contains: 0 0 0 0
proof: green (8e4715c3c13e)

## row 2

command: ./brim detect --band 0.01 material/bordered.jpg
exit: 0
stdout: contains: 2 2 2 2
proof: green (7b4e2712167e)

## row 3

command: ./brim detect --band 0.05 material/bordered.jpg
exit: 0
stdout: contains: 4 4 4 4
proof: green (ccba7e28c02b)

## row 4

command: [ "$(./brim detect material/bordered.jpg)" = "$(./brim detect material/bordered.jpg)" ] && [ "$(./brim detect material/clear.jpg)" = "$(./brim detect material/clear.jpg)" ]
exit: 0
proof: green (df7dc4077679)

