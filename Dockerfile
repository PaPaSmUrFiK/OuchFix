# Один образ на сервис: docker build --build-arg SERVICE=catalog .
FROM golang:1.27.2-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG SERVICE
RUN CGO_ENABLED=0 go build -o /out/app ./cmd/${SERVICE}

FROM alpine:3.22
COPY --from=build /out/app /app
ENTRYPOINT ["/app"]
