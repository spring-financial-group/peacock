FROM alpine:3.16
RUN apk upgrade --no-cache

COPY ./build/linux /

ENTRYPOINT ["peacock", "version"]
