FROM golang:1.14 as builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build -o main .


######## Start a new stage from scratch #######
FROM alpine:latest

# Get the required packages to run the executable script
RUN apk --no-cache add python3 py-pip
RUN pip install awscli

WORKDIR /app/main

# Copy the Pre-built binary file from the previous stage and the bash script
COPY --from=builder /app/main .
COPY execute.sh .

# Make the script executable
RUN chmod +x execute.sh

ENV CUSTOM="ON"

# Command to run the executable
ENTRYPOINT ["/app/main/execute.sh"]
