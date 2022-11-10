FROM golang:1.19-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY cmd cmd
RUN GOOS_ENABLED=0 go build -v -ldflags '-s' -o ./echo ./cmd/echo

FROM alpine
EXPOSE 8080
WORKDIR /app
COPY --from=builder /app/echo echo
RUN chmod 0755 echo
ENTRYPOINT [ "/app/echo" ]
