FROM alpine/git:2.36.2
RUN apk upgrade --no-cache --ignore alpine-baselayout
COPY ./build/linux /
ENTRYPOINT ["./peacock"]
