# Use golang base image to get the Go environment
FROM golang:1.20 as builder

# Install zip utility and bash which is already there in official golang image
RUN apt-get update && \
    apt-get install -y zip && \
    apt-get clean

# Set the shell to bash
SHELL ["/bin/bash", "-c"]

WORKDIR /app

# Copy our application code into the container's app directory
COPY . .

# Run unit tests
RUN make unit_test

# Build the application binary named 'bootstrap' so it can be run in Lambda
RUN make package

