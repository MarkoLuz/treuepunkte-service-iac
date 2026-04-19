FROM golang:1.24-alpine AS builder

WORKDIR /src/treuepunkte-function

COPY treuepunkte-function/go.mod treuepunkte-function/go.sum ./
RUN go mod download

COPY treuepunkte-function/ ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/treuepunkte-service ./main.go


FROM alpine:3.20

WORKDIR /app

RUN adduser -D appuser
COPY --from=builder /app/treuepunkte-service /app/treuepunkte-service

USER appuser

EXPOSE 8080

CMD ["/app/treuepunkte-service"]