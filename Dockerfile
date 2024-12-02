FROM golang:1.20-alpine

WORKDIR /app

COPY go.mod ./
RUN go mod download && go mod verify

COPY . .

RUN go build -v -o ./opengraph-image-creator

EXPOSE 8080

CMD ["./opengraph-image-creator"]