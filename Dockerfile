# Stage 1: Build the Go program
FROM golang:1.24.3-alpine AS build
WORKDIR /opt/go/patch

# Copy the project files and build the program
COPY . .
RUN apk --no-cache add gcc musl-dev
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o patch_app main.go

# Stage 2: Runtime image
FROM alpine:3.21
RUN apk --no-cache add ca-certificates

# Copy binary with unique name to avoid conflict
COPY --from=build /opt/go/patch/patch_app /app/patch_app
RUN chmod +x /app/patch_app

# Use the renamed binary as the command
CMD ["/app/patch_app"]



# Build Image with command
# docker build --no-cache -t patch:1.76v.ecs .
# docker tag patch:1.76v.ecs mochigome/patch:1.76v.ecs
# docker push mochigome/patch:1.76v.ecs

# current version : 1.76v.ecs