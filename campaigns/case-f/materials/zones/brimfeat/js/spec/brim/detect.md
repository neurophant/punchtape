# brim/detect — margins on the given materials

run: 1

The wish's own ground truth: bordered.jpg (250x250, a 4-pixel even frame) answers 4 4 4 4, clear.jpg (no margin) answers 0 0 0 0, a band shallower than the frame cuts the numbers down to the band itself, and a repeat run answers to the letter the same.


scope: brim, package.json, src/**, material/**, node_modules/**

## row 0

command: ./brim detect material/bordered.jpg
exit: 0
stdout: 4 4 4 4

proof: green (b66465b7610a)

## row 1

command: ./brim detect material/clear.jpg
exit: 0
stdout: 0 0 0 0

proof: green (3a7534fb5302)

## row 2

command: ./brim detect material/bordered.jpg --band 0.01
exit: 0
stdout: 2 2 2 2

proof: green (02a495c3e46f)

## row 3

command: ./brim detect material/bordered.jpg --band 0.05
exit: 0
stdout: 4 4 4 4

proof: green (a07fc6e61514)

## row 4

command: ./brim detect material/bordered.jpg && ./brim detect material/bordered.jpg
exit: 0
stdout: 4 4 4 4
4 4 4 4

proof: green (4370bdc9e22e)

