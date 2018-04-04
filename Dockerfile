FROM alpine:latest

MAINTAINER John Weldon <johnweldon4@gmail.com>

RUN apk update && \
    apk upgrade && \
    apk add \
        bind-tools \
        ca-certificates \
        openssl \
    && (update-ca-certificates || true) \
    && rm -rf /var/cache/apk/*

ADD location location
ADD public public

ENTRYPOINT ["/location"]
