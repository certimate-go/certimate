FROM node:24-alpine AS webui-builder
WORKDIR /app
COPY . /app/
RUN \
  cd /app/ui && \
  npm install && \
  npm run build



FROM golang:1.24-alpine AS server-builder
WORKDIR /app
COPY ../. /app/
RUN rm -rf /app/ui/dist
COPY --from=webui-builder /app/ui/dist /app/ui/dist
ENV CGO_ENABLED=0
RUN go build -trimpath -ldflags="-s -w" -o certimate



FROM alpine:latest
WORKDIR /app
COPY --from=server-builder /app/certimate .
ENTRYPOINT ["./certimate", "serve", "--http", "0.0.0.0:8090"]
