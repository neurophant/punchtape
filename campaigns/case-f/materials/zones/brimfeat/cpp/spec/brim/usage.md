# brim/usage — honest refusals

run: 1

Clumsy calls and unusable files are refused with rc 1, a brief English message on stderr and nothing on stdout: a missing file, not a picture at all, another format than JPEG, sides too short to search, an extra file argument, an unknown option or command, a missing command, and each knob's invalid value named along with its requirement.

scope: src/**, Makefile, testdata/**, materials/bordered.jpg

## row 0

command: ./brim detect materials/absent.jpg 2>&1
exit: 1
stdout: contains: cannot read
proof: green (5fd014ed93f3)

## row 1

command: ./brim detect testdata/garbage.bin 2>&1
exit: 1
stdout: contains: not a picture at all
proof: green (6eba65c019a9)

## row 2

command: ./brim detect testdata/broken.jpg 2>&1
exit: 1
stdout: contains: not a picture at all
proof: green (143ca9e33526)

## row 3

command: ./brim detect testdata/notjpeg.png 2>&1
exit: 1
stdout: contains: not my format: PNG
proof: green (4f0a993105d4)

## row 4

command: ./brim detect testdata/tiny.jpg 2>&1
exit: 1
stdout: contains: too small
proof: green (5a07b64075ba)

## row 5

command: ./brim detect materials/bordered.jpg extra.jpg 2>&1
exit: 1
stdout: contains: unexpected extra argument
proof: green (2c59ce2e39d1)

## row 6

command: ./brim detect materials/bordered.jpg --bogus 1 2>&1
exit: 1
stdout: contains: unknown option
proof: green (4c934c7688f5)

## row 7

command: ./brim detect 2>&1
exit: 1
stdout: contains: no picture file given
proof: green (3bc01735e4a2)

## row 8

command: ./brim 2>&1
exit: 1
stdout: contains: missing command
proof: green (46c8edf773dd)

## row 9

command: ./brim detect materials/bordered.jpg --band 0.9 2>&1
exit: 1
stdout: contains: --band: invalid value '0.9': a fraction of the side
proof: green (475f8ab3e435)

## row 10

command: ./brim detect materials/bordered.jpg --band abc 2>&1
exit: 1
stdout: contains: --band: invalid value 'abc'
proof: green (e930953c21a4)

## row 11

command: ./brim detect materials/bordered.jpg --threshold 1.5 2>&1
exit: 1
stdout: contains: --threshold: invalid value '1.5': a fraction from 0 to 1
proof: green (e94a708c2278)

