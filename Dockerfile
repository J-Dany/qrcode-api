FROM golang:1.22.0-alpine AS builder
ARG PORT=8000

WORKDIR /app

COPY . .

RUN go build -o app

FROM alpine:latest
ARG PORT=8000

WORKDIR /app

COPY --from=builder /app/app /app/app

ENV PORT=${PORT}
EXPOSE 8000

CMD ["/app/app"]
