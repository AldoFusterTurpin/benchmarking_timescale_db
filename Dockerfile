# syntax=docker/dockerfile:1

FROM golang:1.22 AS build-stage

WORKDIR /app

COPY . .
COPY go.mod go.sum ./
RUN go mod download
RUN go test -v -race ./...

RUN CGO_ENABLED=0 GOOS=linux go build -o /benchmark

FROM gcr.io/distroless/base-debian11 AS release-stage

WORKDIR /

COPY --from=build-stage /benchmark /benchmark

USER nonroot:nonroot

ENTRYPOINT ["/benchmark"]