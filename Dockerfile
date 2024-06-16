# syntax=docker/dockerfile:1

FROM golang:1.21 AS build-stage

WORKDIR /app

COPY . .
COPY go.mod go.sum ./
RUN go mod download
RUN go test ./...

RUN CGO_ENABLED=0 GOOS=linux go build -o /benchmark

FROM build-stage AS run-test-stage
RUN go test -race ./...

FROM gcr.io/distroless/base-debian11 AS release-stage

WORKDIR /

COPY --from=build-stage /benchmark /benchmark

USER nonroot:nonroot

ENTRYPOINT ["/benchmark"]