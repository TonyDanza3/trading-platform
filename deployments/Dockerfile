FROM golang:1.23-alpine AS build
WORKDIR /src
ENV GOTOOLCHAIN=local
ENV GOPROXY=https://proxy.golang.org,direct
ENV GOSUMDB=off
RUN apk add --no-cache ca-certificates
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/gateway ./services/api-gateway/cmd \
 && CGO_ENABLED=0 GOOS=linux go build -o /out/auth ./services/auth-service/cmd \
 && CGO_ENABLED=0 GOOS=linux go build -o /out/user ./services/user-service/cmd \
 && CGO_ENABLED=0 GOOS=linux go build -o /out/market ./services/market-data-service/cmd

FROM alpine:3.21
RUN apk add --no-cache wget \
 && adduser -D -H -u 10001 app
WORKDIR /app
COPY --from=build /out/gateway /out/auth /out/user /out/market /app/
COPY api/openapi.yaml /app/openapi.yaml
USER app
EXPOSE 8080
ENV HTTP_ADDR=:8080
ENV OPENAPI_PATH=/app/openapi.yaml
ENTRYPOINT ["/app/gateway"]
