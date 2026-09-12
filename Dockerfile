FROM golang:1.25.10-alpine AS builder

WORKDIR /announcements

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o announcements ./cmd/announcements

FROM alpine:latest

WORKDIR /announcements

COPY --from=builder /announcements/announcements .

EXPOSE 8086

CMD ["./announcements"]