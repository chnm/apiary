# Dockerfile References: https://docs.docker.com/engine/reference/builder/

# Start from the latest golang base image
FROM golang:latest

# Add Maintainer Info
LABEL maintainer="G Katchoua <gkatchou@gmu.edu>"

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy dependencies prior to building so that this layer is cached unless
# specified dependencies change
COPY go.mod go.sum /app/
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . /app

# Build the Go app
RUN cd cmd/dataapi && go build -o main .

# Expose port 8090 to the outside world
EXPOSE 8090

# Command to run the executable
CMD ["/app/cmd/dataapi/main"]
