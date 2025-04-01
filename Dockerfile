FROM --platform=$BUILDPLATFORM tonistiigi/xx:1.6.1 AS xx

FROM --platform=$BUILDPLATFORM golang:1.24.2-alpine AS go-builder
WORKDIR /app

COPY --from=xx / /

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
COPY cmd/ cmd/
COPY internal/ internal/

ARG TARGETPLATFORM
RUN --mount=type=cache,target=/root/.cache \
  CGO_ENABLED=0 xx-go build -ldflags="-w -s" -trimpath -tags grpcnotrace


FROM alpine:3.21.3
WORKDIR /app

RUN apk add --no-cache tzdata

ARG USERNAME=castsponsorskip
ARG UID=1000
ARG GID=$UID
RUN addgroup -g "$GID" "$USERNAME" \
    && adduser -S -u "$UID" -G "$USERNAME" "$USERNAME"

COPY --from=go-builder /app/castsponsorskip ./

USER $UID
CMD ["./castsponsorskip"]
