# brim/errors — brief clear trouble reports

run: 1

Missing or unreadable files, bytes that are not a picture at all, a real picture in a format other than JPEG, and pictures too small to search are each refused in one English line on a failing exit.


scope: brim, package.json, src/**, fixtures/**, node_modules/**

## row 0

command: ./brim detect missing.jpg
exit: 1
stdout: contains: cannot read 'missing.jpg': no such file or directory
proof: green (0f23e0a11cb6)

## row 1

command: ./brim detect fixtures/not-picture.txt
exit: 1
stdout: contains: is not a picture at all
proof: green (24d35453c6ec)

## row 2

command: ./brim detect fixtures/photo.png
exit: 1
stdout: contains: is not my format
proof: green (6c98c7eed0ca)

## row 3

command: ./brim detect fixtures/broken.jpg
exit: 1
stdout: contains: is not a picture at all
proof: green (8589af659a31)

## row 4

command: ./brim detect fixtures/tiny.jpg
exit: 1
stdout: contains: too small to search for a margin
proof: green (e484d9d67f97)

