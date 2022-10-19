FROM alpine:3.16
RUN apk upgrade --no-cache --ignore alpine-baselayout
RUN apk add git make
COPY ./build/linux /peacock
ENV PATH "$PATH:/peacock"
