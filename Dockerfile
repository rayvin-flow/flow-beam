FROM golang:alpine as base

# Set necessary environmet variables needed for our image
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Move to working directory /build
WORKDIR /build

# Copy the code into the container
COPY src/ .

# Copy and download dependency using go mod
RUN go mod download

# Build the application
RUN go build -o beam-server ./main/main.go

# Move to /dist directory as the place for resulting binary folder
WORKDIR /dist

# Copy the access-nodes.json file into the container
COPY access-nodes.json .

# Copy binary from build to main folder
RUN cp /build/beam-server .

# Command to run when starting the container
CMD ["/dist/beam-server"]