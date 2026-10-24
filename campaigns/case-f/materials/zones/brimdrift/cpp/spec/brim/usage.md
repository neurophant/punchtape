# brim/usage — honest refusals

run: 2

Clumsy calls and unusable files are refused with rc 1, a brief English message on stderr and nothing on stdout: a missing file, not a picture at all, another format than JPEG, sides too short to search, an extra file argument, an unknown option or command, a missing command, and each knob's invalid value named along with its requirement.

scope: src/**, Makefile, testdata/**, materials/bordered.jpg

## row 0

command: ./brim detect materials/absent.jpg 2>&1
exit: 1
stdout: contains: cannot read
proof: green (746ee819e0ec)

## row 1

command: ./brim detect testdata/garbage.bin 2>&1
exit: 1
stdout: contains: not a picture at all
proof: green (7b71cd2bae59)

## row 2

command: ./brim detect testdata/broken.jpg 2>&1
exit: 1
stdout: contains: not a picture at all
proof: green (017baedf7285)

## row 3

command: ./brim detect testdata/notjpeg.png 2>&1
exit: 1
stdout: contains: not my format: PNG
proof: green (30e6003ea5bf)

## row 4

command: ./brim detect testdata/tiny.jpg 2>&1
exit: 1
stdout: contains: too small
proof: green (37341eea38d0)

## row 5

command: ./brim detect materials/bordered.jpg extra.jpg 2>&1
exit: 1
stdout: contains: unexpected extra argument
proof: green (37e400c28720)

## row 6

command: ./brim detect materials/bordered.jpg --bogus 1 2>&1
exit: 1
stdout: contains: unknown option
proof: green (0d389fc55906)

## row 7

command: ./brim detect 2>&1
exit: 1
stdout: contains: no picture file given
proof: green (c34a9fd90bd6)

## row 8

command: ./brim 2>&1
exit: 1
stdout: contains: missing command
proof: green (eee47906cb33)

## row 9

command: ./brim detect materials/bordered.jpg --band 0.9 2>&1
exit: 1
stdout: contains: --band: invalid value '0.9': a fraction of the side
proof: green (d76f8021cb2e)

## row 10

command: ./brim detect materials/bordered.jpg --band abc 2>&1
exit: 1
stdout: contains: --band: invalid value 'abc'
proof: green (e2124135263b)

## row 11

command: ./brim detect materials/bordered.jpg --threshold 1.5 2>&1
exit: 1
stdout: contains: --threshold: invalid value '1.5': a fraction from 0 to 1
proof: green (6815ebd10788)

