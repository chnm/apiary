# Dockerfile References: https://docs.docker.com/engine/reference/builder/

# Start from the latest golang base image
FROM golang:latest AS compiler

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

# Build the Go app, making sure it is a static binary with no debugging symbols
RUN cd cmd/dataapi && CGO_ENABLED=0 go build -a -ldflags="-w -s"

# Create non-root user information
RUN echo "dataapi:x:65534:65534:Data API:/:" > /etc_passwd

# Start over with a completely empty image
FROM scratch

# Copy over just the static binary to root
COPY --from=compiler /app/cmd/dataapi/dataapi /dataapi

# Copy over non-root user information
COPY --from=0 /etc_passwd /etc/passwd

# Run as non-root user in container
USER dataapi

# Expose port 8090 to the outside world
EXPOSE 8090

# Command to run the executable
CMD ["/dataapi"]
