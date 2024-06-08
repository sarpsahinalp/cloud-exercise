# Use the official golang image as a base image
FROM golang:1.22-bullseye

# Set the working directory inside the container
WORKDIR /app

# Copy the main.go file from your local machine into the container at /app
COPY . .

# Expose the port if your Go program listens on a specific port
EXPOSE 3030

# Command to run the main.go file with 'go run'
CMD ["go", "run", "cmd/main.go"]