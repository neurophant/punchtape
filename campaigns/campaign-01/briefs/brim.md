# Brief: brim — measuring margins on scans

I scan and photograph documents, then crop the margins by hand.
I want a program that measures the margins itself and tells me how much
to cut off.

## Stack and tools — my requirement, not the executor's choice

C++17, built with g++. Reading and decoding pictures is done by the
GraphicsMagick++ library (a system package, with the
GraphicsMagick++-config wrapper). The build is one builder command run
from the project root (Makefile or CMake — the executor's choice).

## The idea: evenness is measured with entropy

A margin strip differs from the content by its disorder. An even strip
has no disorder — its entropy (a measure of chaos) is near zero; with
content the picture is alive — entropy is high. The program finds the
boundary like this: it walks from the edge inward and watches where the
band's entropy stops being low — there the margin ends and the content
begins. The boundary is where the difference in chaos between
neighboring chunks is at its largest.

## What it does

I give it a picture, and it answers with four numbers: how many pixels
of margin there are on the top, bottom, left, and right. The margin is
an even strip running from the edge: white, black, anything — the main
thing is that it's even. Each side is measured on its own; evenness can
stretch all the way to the end of the search band — then the whole band
is the answer; no margin means zero.

## How I use it

    brim detect image.jpg

Three handy knobs:

- how deep from the edge to search for the margin — a fraction of the
  side, no more than half allowed, a quarter of the side by default:
  `--band 0.25`;
- the sensitivity threshold — a fraction from 0 to 1, half by default:
  `--threshold 0.5`. The higher it is, the more aggressively the
  boundary is hunted;
- strict mode `--thorough`: the boundary is refined over several
  passes; the numbers on each side come out no smaller than in fast
  mode.

## How it must work

- My pictures come in two kinds: JPEG (ordinary) and GIF. An animated
  GIF is measured across ALL frames; the answer for each side is the
  most cautious (smallest) one among the frames.
- A repeat run on the same picture gives the same answer, to the
  letter.

## Test material — attached, the same on every run

Three pictures sit next to the task in the `materials/` folder:

- `bordered.jpg` — still, 250×250, a thin black even frame. I expect
  four identical numbers — this material has four pixels on each side;
- `bordered.gif` — animated, 480×265, thirty-eight frames, a wide
  margin. Across frames I take the most cautious (smallest) value of
  each side. Fast mode gives on the order of 12 top, 72 right, 9
  bottom, 80 left; strict mode refines it to on the order of
  22 / 85 / 16 / 85 — no side smaller than in fast mode;
- `clear.jpg` — still, 510×350, no margin: zeros on all sides.

A small search band (say 0.05) cuts down the found numbers on pictures
that have a margin.

## Limits and errors

Let it report trouble briefly and clearly, in English: the file is
missing or unreadable; this is not a picture at all; the picture is so
small there is nothing to search for a margin in (a side shorter than
two pixels). A clumsy invocation — a bad knob value, an extra file
argument, an unknown option — is an honest refusal, and each knob's
invalid value is named along with its requirement.

## What must not happen

The program only reads the picture and prints the answer: it writes
nothing to disk, goes to no network, keeps no settings in files.
