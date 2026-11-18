# ---- build ----
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/bibliotheca ./cmd/server

# ---- run ----
FROM alpine:3.20
RUN adduser -D -u 10001 app
USER app
COPY --from=build /out/bibliotheca /bibliotheca
EXPOSE 8080
ENTRYPOINT ["/bibliotheca"]
