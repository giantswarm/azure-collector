FROM alpine:3.14.0

RUN apk add --no-cache ca-certificates

ADD ./azure-collector /azure-collector

ENTRYPOINT ["/azure-collector"]
