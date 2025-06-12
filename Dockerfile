FROM golang:1.24-alpine AS builder

WORKDIR /src

COPY go.mod .
COPY go.sum .
COPY . .

RUN --mount=type=ssh --mount=type=cache,target=/go/pkg/mod go mod tidy
RUN --mount=type=ssh --mount=type=cache,target=/go/pkg/mod mkdir -p ./bin && go build -o ./bin ./cmd/playwallet 


FROM debian:stable-slim

ENV TZ=Asia/Singapore
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

WORKDIR /app

COPY --from=builder /src/bin/playwallet .
COPY --from=builder /src/config config

ENTRYPOINT ["/app/playwallet"]

