#!/bin/sh
# Essential bash file

echo "Getting the files from S3"
aws s3 cp "$CCR_DICTIONARY" /app/main/dictionary.txt
aws s3 cp "$CCR_HASH" /app/main/hash.txt

echo "Running main executable"
/app/main

echo "Uploading output to S3"
aws s3 cp /app/main/output "$CCR_OUTPUT"

echo "Shutting off"
poweroff