# Use the official Golang image as a base
FROM golang:1.23

# Set the working directory
WORKDIR /app

# Copy all files
COPY . .

# Download dependencies
RUN go mod download && go mod verify

# Build the application
RUN CGO_ENABLED=0 go build -v -o app .

FROM scratch
COPY --from=0 /app/app /app
EXPOSE 8080
CMD ["/app"]
