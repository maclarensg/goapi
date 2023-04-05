# Use an official Golang runtime as a parent image
FROM golang:1.17-alpine as build

# Set the working directory to /app
WORKDIR /app

# Copy the current directory contents into the container at /app
COPY . /app

# Build the Go application
RUN go build -o app

# Test the Go application when 
RUN go test .

# Use a minimal Alpine image to run the application
FROM alpine:3.14

# Set the working directory to /app
WORKDIR /app

# Copy the binary from the build image to the final image
COPY --from=build /app/app .

# Expose port 3000
EXPOSE 3000

# Start the application
CMD ["./app"]
