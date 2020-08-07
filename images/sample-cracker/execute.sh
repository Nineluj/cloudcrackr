#!/bin/sh

# Essential bash file
echo "Getting the files from S3"
aws s3 cp "$CCR_DICTIONARY" /app/main/dictionary.txt
aws s3 cp "$CCR_HASH" /app/main/hash.txt

# Sync the output of the program to the S3 folder every 15 seconds
watch -n 15 aws sync /app/main/output/ "$CCR_OUTPUT/" &

echo "Running main executable"
/app/main > "$CCR_OUTPUT/stdout"

echo "Uploading output to S3"
aws sync /app/main/output/ "$CCR_OUTPUT/"

echo "Shutting off"
poweroff