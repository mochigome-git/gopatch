# Stage 1: Build the Go program
FROM golang:1.23.3-alpine AS build
WORKDIR /opt/go/patch

# Copy the project files and build the program
COPY . .
RUN apk --no-cache add gcc musl-dev
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o patch main.go

# Stage 2: Copy the built Go program into a minimal container
FROM alpine:3.18
RUN apk --no-cache add ca-certificates

# Copy the Go binary from the first stage
COPY --from=build /opt/go/patch/patch /app/patch

RUN chmod +x /app/patch

CMD ["/app/patch"]


# Build Image with command
# docker build -t patch:${version} .
# docker tag patch:${version} mochigome/patch:${version}
# docker push mochigome/patch:tagname