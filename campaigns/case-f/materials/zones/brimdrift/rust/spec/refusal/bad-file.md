# refusal/bad-file — File trouble reported briefly and clearly

run: 2

A missing or unreadable file, a file that is not a picture at all, a picture in another format than JPEG, and a picture too small to search (a side shorter than two pixels) are each refused with a one-line English error and a nonzero exit.


scope: src/**, Cargo.toml, Cargo.lock, build, materials/notpic.jpg, materials/picture.png, materials/tiny.jpg

## row 0

command: ./brim detect materials/missing.jpg
exit: 1
proof: green (2e3d84573eb7)

## row 1

command: ./brim detect materials/notpic.jpg
exit: 1
proof: green (bf056daf8cd8)

## row 2

command: ./brim detect materials/picture.png
exit: 1
proof: green (856f79f48ecb)

## row 3

command: ./brim detect materials/tiny.jpg
exit: 1
proof: green (de0fc190ef5f)

