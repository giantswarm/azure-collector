FROM alpine:3.15.2

RUN apk add --no-cache ca-certificates

ADD ./azure-collector /azure-collector

ENTRYPOINT ["/azure-collector"]
