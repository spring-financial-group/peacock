FROM alpine:3.16
RUN apk upgrade --no-cache --ignore alpine-baselayout
COPY ./build/linux /
ENTRYPOINT ["./peacock"]
