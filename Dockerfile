# Compile stage
FROM golang:1.13.4 AS build-env

ADD . /dockerdev
WORKDIR /dockerdev

RUN go build -o /server

# Final stage
FROM debian:buster

EXPOSE 8000

WORKDIR /
COPY --from=build-env /server /

CMD ["/server"]
