# safety/read-only-run — The run reads the picture and writes nothing anywhere

run: 1

The wish's "what must not happen": the program only reads the picture and prints the answer — it writes nothing to disk, goes to no network, keeps no settings in files. One row snapshots every file in the working tree (the machine's own .punchtape zone and the snapshot's own scratch excluded) with size and mtime, runs a detect, and snapshots again; the two snapshots must be identical, which catches any file created, removed, rewritten, or re-timed by the run. The other row runs a detect under a fresh empty HOME and demands that HOME stay empty — a settings file is a disk write, so the two checks carry both bans.

scope: src/**, Cargo.toml, Cargo.lock, build, materials/bordered.jpg

## row 0

command: mkdir -p .snap && find . \( -path ./.punchtape -o -path ./.snap \) -prune -o -type f -printf "%p %s %T@\n" | sort > .snap/before && ./brim detect materials/bordered.jpg && find . \( -path ./.punchtape -o -path ./.snap \) -prune -o -type f -printf "%p %s %T@\n" | sort > .snap/after && diff .snap/before .snap/after && rm -rf .snap && echo NO_WRITES
exit: 0
stdout: 4 4 4 4
NO_WRITES

proof: green (bcbd42d4ceaa)

## row 1

command: rm -rf .probe-home && mkdir .probe-home && HOME="$PWD/.probe-home" ./brim detect materials/bordered.jpg && [ -z "$(ls -A .probe-home)" ] && rm -rf .probe-home && echo "keeps no settings in files"
exit: 0
stdout: 4 4 4 4
keeps no settings in files

proof: green (2f3ca0a467f5)

