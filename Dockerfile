# Multi-stage build for guardmesh single binary
FROM node:22-alpine AS web-builder
WORKDIR /web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM golang:1.24-alpine AS go-builder
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web-builder /web/dist ./web/dist
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o guardmesh ./cmd/guardmesh

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=go-builder /app/guardmesh /usr/local/bin/guardmesh
COPY guardmesh.yaml /app/guardmesh.yaml
EXPOSE 8080
ENTRYPOINT ["guardmesh"]
CMD ["start", "--config", "/app/guardmesh.yaml"]
