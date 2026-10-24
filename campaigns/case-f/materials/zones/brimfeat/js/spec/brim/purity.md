# brim/purity — reads the picture, writes nothing

run: 1

The program only reads the picture and prints the answer: after a detect run the tree holds exactly the files it held before.


scope: brim, package.json, src/**, material/**, node_modules/**

## row 0

command: before=$(find . -type f -not -path "./.punchtape/*" | sort); ./brim detect material/clear.jpg > /dev/null; after=$(find . -type f -not -path "./.punchtape/*" | sort); [ "$before" = "$after" ]
exit: 0
proof: green (90f99903cfc4)

