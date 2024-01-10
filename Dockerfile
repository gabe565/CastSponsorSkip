FROM golang:1.21.6-alpine as go-builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
COPY cmd/ cmd/
COPY internal/ internal/

RUN --mount=type=cache,target=/root/.cache \
    go build -ldflags="-w -s" -trimpath


FROM alpine:3.19
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
