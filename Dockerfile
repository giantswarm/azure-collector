FROM alpine:3.17.3

RUN apk add --no-cache ca-certificates

ADD ./azure-collector /azure-collector

ENTRYPOINT ["/azure-collector"]
