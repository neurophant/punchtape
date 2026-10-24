# brim/errors — brief clear trouble reports

run: 1

Missing or unreadable files, bytes that are not a picture at all, a real picture in a format other than JPEG, and pictures too small to search are each refused in one English line on a failing exit.


scope: brim, package.json, src/**, fixtures/**, node_modules/**

## row 0

command: ./brim detect missing.jpg
exit: 1
stdout: contains: cannot read 'missing.jpg': no such file or directory
proof: green (3316f8a20829)

## row 1

command: ./brim detect fixtures/not-picture.txt
exit: 1
stdout: contains: is not a picture at all
proof: green (4542b9522173)

## row 2

command: ./brim detect fixtures/photo.png
exit: 1
stdout: contains: is not my format
proof: green (2e11139a4831)

## row 3

command: ./brim detect fixtures/broken.jpg
exit: 1
stdout: contains: is not a picture at all
proof: green (be5b7b9eb533)

## row 4

command: ./brim detect fixtures/tiny.jpg
exit: 1
stdout: contains: too small to search for a margin
proof: green (95e5b88d820a)

