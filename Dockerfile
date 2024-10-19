# Build stage
FROM docker.arvancloud.ir/golang:1.23.1 AS build

RUN mkdir /app/distributed-file-system -p
RUN chmod +rwx /app/distributed-file-system
WORKDIR /app/distributed-file-system

# Copy the source code
COPY . .

# List contents to debug
RUN ls -la

# Build the application with static linking
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/fs driver/main.go

# List contents after build
RUN ls -la /app/distributed-file-system/bin

# Final stage
FROM docker.arvancloud.ir/alpine:3.18

RUN mkdir /app/distributed-file-system -p
RUN chmod +rwx /app/distributed-file-system

WORKDIR /app/distributed-file-system

# Copy the binary from the build stage
COPY --from=build /app/distributed-file-system/bin/fs /app/distributed-file-system/fs

# Ensure the binary is executable
RUN chmod +x /app/distributed-file-system/fs

# Debug: List contents of /app
RUN ls -la /app/distributed-file-system

# Expose necessary ports
EXPOSE 3000 7000 5000

# Run the binary
CMD ["/app/distributed-file-system/fs"]
