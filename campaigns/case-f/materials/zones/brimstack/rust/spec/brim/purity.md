# brim/purity — reads the picture, writes nothing

run: 1

The tool only reads the picture and prints the answer: run from an empty directory it leaves that directory empty — nothing written to disk, no settings kept in files.

scope: src/**, Makefile, materials/bordered.jpg

## row 0

command: mkdir .purity-check && cd .purity-check && ../brim detect ../materials/bordered.jpg > /dev/null && ls -A | wc -l; cd .. && rm -rf .purity-check
exit: 0
stdout: 0

proof: green (cf8e691fab0c)

