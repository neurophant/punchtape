# cli/errors — refuse trouble honestly

run: 1

Every refusal is one brief English line on stderr and a nonzero exit: usage trouble (a bad knob value, an extra file argument, an unknown option) exits 2; picture trouble (a missing or unreadable file, not a picture at all, not JPEG, a side shorter than two pixels) exits 1. Each knob's invalid value is named together with its requirement.


scope: brim, material/**

## row 0

command: ./brim detect material/nosuch.jpg 2>&1
exit: 1
stdout: contains: cannot read material/nosuch.jpg: no such file
proof: green (bde028918d78)

## row 1

command: ./brim detect material/notjpeg.txt 2>&1
exit: 1
stdout: contains: not a picture at all
proof: green (3ce57b71aebc)

## row 2

command: ./brim detect material/photo.png 2>&1
exit: 1
stdout: contains: not a JPEG picture
proof: green (655177863894)

## row 3

command: ./brim detect material/tiny.jpg 2>&1
exit: 1
stdout: contains: picture too small to search for a margin
proof: green (2321f42e7598)

## row 4

command: ./brim detect --band 0.6 material/bordered.jpg 2>&1
exit: 2
stdout: contains: invalid --band 0.6: want a fraction of the side, greater than 0 and at most 0.5
proof: green (f537b5a18583)

## row 5

command: ./brim detect --band 0 material/bordered.jpg 2>&1
exit: 2
stdout: contains: invalid --band 0: want a fraction of the side, greater than 0 and at most 0.5
proof: green (d49b4e44219c)

## row 6

command: ./brim detect --band abc material/bordered.jpg 2>&1
exit: 2
stdout: contains: invalid --band abc: want a fraction of the side, greater than 0 and at most 0.5
proof: green (1c2cfdce9fad)

## row 7

command: ./brim detect --threshold 1.5 material/bordered.jpg 2>&1
exit: 2
stdout: contains: invalid --threshold 1.5: want a fraction from 0 to 1
proof: green (66c0b1614d70)

## row 8

command: ./brim detect --threshold -0.1 material/bordered.jpg 2>&1
exit: 2
stdout: contains: invalid --threshold -0.1: want a fraction from 0 to 1
proof: green (2bb9f0dd2b1d)

## row 9

command: ./brim detect material/bordered.jpg material/clear.jpg 2>&1
exit: 2
stdout: contains: unrecognized arguments: material/clear.jpg
proof: green (152a46e81935)

## row 10

command: ./brim --bogus detect material/bordered.jpg 2>&1
exit: 2
stdout: contains: unrecognized arguments: --bogus
proof: green (5de4fd8be475)

