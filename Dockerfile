
ARG GO_VERSION=1.11

FROM golang:${GO_VERSION}-alpine AS builder
ENV GOPROXY=https://goproxy.io
WORKDIR /src
COPY ./ ./
RUN CGO_ENABLED=0 go build \
    -installsuffix 'static' \
    -o /tv_server .

FROM scratch AS final
COPY --from=builder /tv_server /tv_server

ENTRYPOINT ["/tv_server"]
