"""enimda.cli — print the border offsets of an image from the command line."""
import sys

from . import ENIMDA


def main():
    if len(sys.argv) != 2:
        print("usage: python3 -m enimda.cli <image>", file=sys.stderr)
        return 2
    borders = ENIMDA(fp=sys.argv[1]).scan(fast=True)
    print(" ".join(str(value) for value in borders))
    return 0


if __name__ == "__main__":
    sys.exit(main())
