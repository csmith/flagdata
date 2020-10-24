#!/bin/sh

set -e

mkdir -p images/webp images/webp-resized images/jpg-resized

for file in images/original/*.jpg; do
  basename=${file##*/}
  extless=${basename%.*}

  # Flags have a 1 pixel border around them, shave it off
  convert "$file" -shave 1x1 "images/jpg/$basename"

  # Resize and convert to webp
  convert "images/jpg/$basename" -resize 250x250 "images/jpg-resized/$basename"
  cwebp "images/jpg/$basename" -mt -m 6 -o "images/webp/$extless.webp"
  cwebp "images/jpg/$basename" -resize 250 0 -mt -m 6 -o "images/webp-resized/$extless.webp"
done
