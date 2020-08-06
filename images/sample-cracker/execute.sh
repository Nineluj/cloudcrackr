#!/bin/sh

# Essential bash file
aws s3 cp "$CCR_DICTIONARY" /app/main/dictionary.txt
aws s3 cp "$CCR_HASH" /app/main/hash.txt

/app/main

aws s3 cp /app/main/output "$CCR_OUTPUT"