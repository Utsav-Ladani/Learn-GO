# Use the official Go image as a parent image
FROM golang:1.17

# Set the working directory inside the container
WORKDIR /app

# Copy the Go application source code into the container
COPY /main.go /app/main.go
COPY /tmpl /app/tmpl
COPY /asset /app/asset

# Build the Go application
RUN go build -o myapp main.go

# Create a data directory
RUN mkdir data

# Create a sample file
RUN echo "Welcome" > /app/data/welcome.txt

# Expose the port that the application will run on
EXPOSE 8080

# Define the command to run your application
CMD ["./myapp"]
