# refusal/bad-file — File trouble reported briefly and clearly

run: 1

A missing or unreadable file, a file that is not a picture at all, a picture in another format than JPEG, and a picture too small to search (a side shorter than two pixels) are each refused with a one-line English error and a nonzero exit.


scope: src/**, Cargo.toml, Cargo.lock, build, materials/notpic.jpg, materials/picture.png, materials/tiny.jpg

## row 0

command: ./brim detect materials/missing.jpg
exit: 1
proof: green (8d6a538b4564)

## row 1

command: ./brim detect materials/notpic.jpg
exit: 1
proof: green (e99c28537d03)

## row 2

command: ./brim detect materials/picture.png
exit: 1
proof: green (27bd6b5406cc)

## row 3

command: ./brim detect materials/tiny.jpg
exit: 1
proof: green (5e8a776fdd4e)

