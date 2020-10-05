FROM alpine:3.12
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY azure-collector /app
ENTRYPOINT ["/app/azure-collector"]
