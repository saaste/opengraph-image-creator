FROM golang:1.23-alpine AS build

WORKDIR /app

COPY go.mod ./
RUN go mod download && go mod verify

COPY . .

RUN go build -v -o ./opengraph-image-creator


FROM chromedp/headless-shell:latest AS final

WORKDIR /app
COPY --from=build /app/ ./

EXPOSE 8080

ENTRYPOINT ["/app/opengraph-image-creator"]



