FROM golang:1.24 as builder
WORKDIR /app/build
COPY . .
RUN apt-get update
RUN apt-get install -y build-essential gcc sqlite3
RUN go mod download
RUN CGO_ENABLED=1 GOOS=linux go build --ldflags '-linkmode external -extldflags "-static"' -o eventloop-bin .

FROM scratch
WORKDIR /app
COPY --from=builder /app/build/eventloop-bin .

EXPOSE 8080
CMD ["/app/eventloop-bin"]

