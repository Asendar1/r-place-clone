FROM golang:1.25.5-alpine3.23

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o r_place .

EXPOSE 8080

CMD [ "./r_place" ]
