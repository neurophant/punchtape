# detect/picture-troubles — a missing, foreign, or too small picture is reported briefly and clearly

run: 1

A missing or unreadable file, a file that is not a JPEG picture, and a picture with a side shorter than two pixels — too small to search for a margin in — are each reported in one brief English line on stderr with exit code 1.

scope: main.go, picture.go, testdata/not-jpeg.txt, testdata/tiny.jpg

## row 0

command: ./brim detect testdata/missing.jpg
exit: 1
proof: green (84ea45bb3d17)

## row 1

command: ./brim detect testdata/not-jpeg.txt
exit: 1
proof: green (86be8e9f9227)

## row 2

command: ./brim detect testdata/tiny.jpg
exit: 1
proof: green (20cd28adcfca)

