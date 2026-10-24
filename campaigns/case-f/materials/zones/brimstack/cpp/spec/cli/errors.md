# cli/errors — refuse trouble honestly

run: 1

Every refusal is one brief English line on stderr and a nonzero exit: usage trouble (a bad knob value, an extra file argument, an unknown option) exits 2; picture trouble (a missing or unreadable file, not a picture at all, not JPEG, a side shorter than two pixels) exits 1. Each knob's invalid value is named together with its requirement.


scope: brim, material/**

## row 0

command: ./brim detect material/nosuch.jpg 2>&1
exit: 1
stdout: contains: cannot read material/nosuch.jpg: no such file
proof: green (fcb47929f7e5)

## row 1

command: ./brim detect material/notjpeg.txt 2>&1
exit: 1
stdout: contains: not a picture at all
proof: green (9bff9882e260)

## row 2

command: ./brim detect material/photo.png 2>&1
exit: 1
stdout: contains: not a JPEG picture
proof: green (2b59d9bb5649)

## row 3

command: ./brim detect material/tiny.jpg 2>&1
exit: 1
stdout: contains: picture too small to search for a margin
proof: green (c8bbba9ed8d6)

## row 4

command: ./brim detect --band 0.6 material/bordered.jpg 2>&1
exit: 2
stdout: contains: invalid --band 0.6: want a fraction of the side, greater than 0 and at most 0.5
proof: green (220bd0aa2974)

## row 5

command: ./brim detect --band 0 material/bordered.jpg 2>&1
exit: 2
stdout: contains: invalid --band 0: want a fraction of the side, greater than 0 and at most 0.5
proof: green (1913f030ef92)

## row 6

command: ./brim detect --band abc material/bordered.jpg 2>&1
exit: 2
stdout: contains: invalid --band abc: want a fraction of the side, greater than 0 and at most 0.5
proof: green (18d270f1f28e)

## row 7

command: ./brim detect --threshold 1.5 material/bordered.jpg 2>&1
exit: 2
stdout: contains: invalid --threshold 1.5: want a fraction from 0 to 1
proof: green (9a9541f1525b)

## row 8

command: ./brim detect --threshold -0.1 material/bordered.jpg 2>&1
exit: 2
stdout: contains: invalid --threshold -0.1: want a fraction from 0 to 1
proof: green (426d85191670)

## row 9

command: ./brim detect material/bordered.jpg material/clear.jpg 2>&1
exit: 2
stdout: contains: unrecognized arguments: material/clear.jpg
proof: green (2f8ec714890c)

## row 10

command: ./brim --bogus detect material/bordered.jpg 2>&1
exit: 2
stdout: contains: unrecognized arguments: --bogus
proof: green (a9c6b64525ee)

