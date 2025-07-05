# Go Builder Stage
ARG GO_VERSION=1.24.4
FROM golang:${GO_VERSION}-alpine AS builder

RUN apk add --no-cache git alpine-sdk
WORKDIR /api

COPY ./server/go.mod ./server/go.sum ./
RUN go mod download

COPY ./server ./
RUN go build -o app ./main.go

# Node Build Stage
FROM node:16-alpine AS build-stage

WORKDIR /app

COPY ./ui/package*.json ./
RUN apk add --no-cache \
    autoconf \
    automake \
    build-base \
    nasm \
    libc6-compat \
    python3 \
    make \
    g++ \
    libpng-dev \
    zlib-dev \
    pngquant

RUN npm install
COPY ./ui ./
RUN npm run build

# Final Runtime Stage
FROM alpine:latest

LABEL org.opencontainers.image.source="https://github.com/alfhou/hammond"

ENV CONFIG=/config \
    DATA=/assets \
    UID=998 \
    PID=100 \
    GIN_MODE=release

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /api

COPY --from=builder /api/app ./
COPY --from=build-stage /app/dist ./dist

EXPOSE 3000
ENTRYPOINT ["./app"]
