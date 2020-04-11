# Dockerfile References: https://docs.docker.com/engine/reference/builder/

# Start from the latest golang base image
FROM golang:latest

# Add Maintainer Info
LABEL maintainer="Katchoua <webmaster@chnm.gmu.edu>"

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the source from the current directory to the Working Directory inside the container
COPY . /app

# Build the Go app
RUN cd cmd/dataapi && go build -o main .

# Expose port 8090 to the outside world
EXPOSE 8090

# Command to run the executable
CMD ["/app/md/dataapi/main"]
