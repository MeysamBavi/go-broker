## Build
FROM golang:1.19 AS build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 go build -o ./go-broker

## Deploy
FROM gcr.io/distroless/base-debian11

WORKDIR /app

COPY ./config.json .
COPY --from=build /app/go-broker .

ENTRYPOINT ["./go-broker"]