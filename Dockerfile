FROM golang:1.22-alpine

# Set up environment and install necessary packages
RUN apk add --no-cache git netcat-openbsd gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application
COPY . .

# Set the working directory to cmd where the main.go is located
WORKDIR /app/cmd

# Build the Go app
RUN go build -o /app/main .

# Command to run the executable
CMD ["/app/main"]