FROM golang:1.17 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./

ENV GOOS=linux
ENV CGO_ENABLED=0

RUN go build -o main ./cmd/main.go

FROM alpine:latest

WORKDIR /app

ENV TZ=Asia/Bangkok

COPY --from=builder /app/main .

CMD ["/app/main"]
