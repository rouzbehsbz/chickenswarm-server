FROM golang:1.23.6-alpine3.21 AS builder
RUN apk update && apk add --no-cache make
WORKDIR /app
COPY go.mod go.sum ./
ENV GOPROXY=https://goproxy.io,direct
ENV GO111MODULE=on CGO_ENABLED=0
RUN go mod download
COPY . .
RUN make build

FROM golang:1.23.6-alpine3.21
RUN apk update && apk add --no-cache make
RUN adduser -D guard
USER guard
WORKDIR /app
COPY --from=builder /app/.bin/ ./.bin
COPY --from=builder /app/Makefile .
CMD ["make", "run-prod"]