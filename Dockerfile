FROM golang:1.27.2-bookworm AS builder

WORKDIR /src

# Сначала зависимости: слой кешируется, пока не меняются go.mod и go.sum.
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

ENV CGO_ENABLED=0

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -ldflags="-s -w" -o /out/iam ./cmd/iam && \
    go build -trimpath -ldflags="-s -w" -o /out/catalog ./cmd/catalog && \
    go build -trimpath -ldflags="-s -w" -o /out/appointment ./cmd/appointment && \
    go build -trimpath -ldflags="-s -w" -o /out/payment ./cmd/payment && \
    go build -trimpath -ldflags="-s -w" -o /out/notification ./cmd/notification && \
    go build -trimpath -ldflags="-s -w" -o /out/gateway ./cmd/gateway

FROM gcr.io/distroless/static-debian12:nonroot AS ouchfix

WORKDIR /apps

COPY --from=builder --chown=nonroot:nonroot /out/iam iam
COPY --from=builder --chown=nonroot:nonroot /out/catalog catalog
COPY --from=builder --chown=nonroot:nonroot /out/appointment appointment
COPY --from=builder --chown=nonroot:nonroot /out/payment payment
COPY --from=builder --chown=nonroot:nonroot /out/notification notification
COPY --from=builder --chown=nonroot:nonroot /out/gateway gateway

USER nonroot:nonroot
