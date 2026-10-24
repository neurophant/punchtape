# tool/stack — a Go program on the standard library alone, no third-party dependencies

run: 1

The stack is the customer's demand, not the executor's choice: the tool is a Go program and it uses the standard library only. The module declares no requirements — go.mod carries the single module line and the go directive, nothing else; go list -m all names the main module and nothing but it; and the built binary's own build info contains no dependency line, so no third-party package rides inside the delivered program.

scope: go.mod, build, main.go, picture.go, entropy.go, margin.go

## row 0

command: cat go.mod
exit: 0
stdout: module brim

go 1.27

proof: green (7f6a371159cf)

## row 1

command: go list -m all
exit: 0
stdout: brim

proof: green (2bda2b844ed4)

## row 2

command: test -z "$(go version -m ./brim | grep "^dep")" && echo "no third-party dependencies"
exit: 0
stdout: no third-party dependencies

proof: green (17f0a8d8013a)

