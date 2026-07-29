FROM golang:1.25.10-alpine AS builder

WORKDIR /announcements

RUN apk add --no-cache build-base libwebp-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o announcements .

FROM alpine:latest

WORKDIR /announcements

RUN apk add --no-cache libwebp

COPY --from=builder /announcements/announcements .

EXPOSE 8086

CMD ["./announcements"]