# brim/detect — margins on the given materials

run: 1

The wish's own ground truth: bordered.jpg (250x250, a 4-pixel even frame) answers 4 4 4 4, clear.jpg (no margin) answers 0 0 0 0, a band shallower than the frame cuts the numbers down to the band itself, and a repeat run answers to the letter the same.


scope: brim, package.json, src/**, material/**, node_modules/**

## row 0

command: ./brim detect material/bordered.jpg
exit: 0
stdout: 4 4 4 4

proof: green (aec476d0f982)

## row 1

command: ./brim detect material/clear.jpg
exit: 0
stdout: 0 0 0 0

proof: green (9168a71c9dc2)

## row 2

command: ./brim detect material/bordered.jpg --band 0.01
exit: 0
stdout: 2 2 2 2

proof: green (a15bd9fd3a30)

## row 3

command: ./brim detect material/bordered.jpg --band 0.05
exit: 0
stdout: 4 4 4 4

proof: green (79e37e524189)

## row 4

command: ./brim detect material/bordered.jpg && ./brim detect material/bordered.jpg
exit: 0
stdout: 4 4 4 4
4 4 4 4

proof: green (5d723a34c650)

