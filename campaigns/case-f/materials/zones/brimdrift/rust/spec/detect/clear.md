# detect/clear — No margin on the free-content picture

run: 2

A picture whose edges carry content and no even strip answers zeros on all four sides.


scope: src/**, Cargo.toml, Cargo.lock, build, materials/clear.jpg

## row 0

command: ./brim detect materials/clear.jpg
exit: 0
stdout: 0 0 0 0

proof: green (a29b4a2b56b4)

