FROM alpine:3.16
RUN apk upgrade --no-cache --ignore alpine-baselayout
RUN apk add git
COPY ./build .
ENV PATH "$PATH:/cli"
CMD ["./api/peacock"]
