FROM golang:1.18.10 as builder
WORKDIR /root/
COPY . /root/
RUN CGO_ENABLED=0 GOOS=linux go build whois_server.go

FROM alpine:latest as prod
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /root/whois_server .
CMD ["./whois_server","-port=8088"]
