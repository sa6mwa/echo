FROM golang:1.19-alpine AS builder
ARG VERSION
ENV VERSION=${VERSION}
WORKDIR /app
COPY vendor vendor
COPY go.mod ./
COPY go.sum ./
COPY cmd cmd
RUN CGO_ENABLED=0 go build -v -ldflags="-s -X main.versionOverride=${VERSION}"  -o echo ./...

FROM alpine
EXPOSE 8080
WORKDIR /app
COPY --from=builder /app/echo echo
RUN chmod 0755 echo
ENTRYPOINT [ "/app/echo" ]
