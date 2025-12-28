FROM golang:1.25.5-alpine3.23 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o r_place .

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/r_place .

COPY --from=builder /app/static ./static

EXPOSE 8080

CMD ["./r_place"]
