#!/bin/sh

# Get the arguments
CCR_DICTIONARY=$1
CCR_HASH=$2
CCR_OUTPUT=$3

# Set up the environment variables to make the CLI work
export AWS_ACCESS_KEY_ID=$4
export AWS_SECRET_ACCESS_KEY=$5
export AWS_DEFAULT_REGION=$6

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