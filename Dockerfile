FROM golang:alpine3.22 AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
COPY main.go ./

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o ssh2telnet .

FROM gcr.io/distroless/static-debian12

COPY --from=builder /app/ssh2telnet /

ENTRYPOINT ["/ssh2telnet"]
