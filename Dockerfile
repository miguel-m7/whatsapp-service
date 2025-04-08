# Use the official Go image as a base image
FROM golang:1.24.1

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go application
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/server.bin ./server/server.go && \
    chmod +x /app/server.bin

# Expose the port the server listens on
EXPOSE 50051

# Command to run the application
CMD ["/app/server.bin"]