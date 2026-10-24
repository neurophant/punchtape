# detect/repeatable — Repeat runs answer the same to the letter

run: 2

Running detect twice on each test picture prints byte-identical answers, as the wish demands of a repeat run.


scope: src/**, Cargo.toml, Cargo.lock, build, materials/bordered.jpg, materials/clear.jpg

## row 0

command: ./brim detect materials/bordered.jpg && ./brim detect materials/bordered.jpg && ./brim detect materials/clear.jpg && ./brim detect materials/clear.jpg
exit: 0
stdout: 4 4 4 4
4 4 4 4
0 0 0 0
0 0 0 0

proof: green (b89204eb1e9c)

