# Use an official Golang runtime as a parent image
FROM golang:1.16-alpine

# Set the working directory to /app
WORKDIR /app

# Copy the Go module files
COPY go.mod go.sum ./

# Install the dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the Go app
RUN go build -o server server.go

# Expose port 5000
EXPOSE 5000

# Run the app
CMD ["./server"]