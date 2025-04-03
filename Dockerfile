FROM golang:1.24-bookworm as build
RUN apt update && apt install -y libtesseract-dev
ADD . /app
WORKDIR /app
RUN go build -o /usr/bin/rok-server ./cmd/rok-server

FROM debian:bookworm
RUN apt update && apt install -y libtesseract5
COPY --from=build /usr/bin/rok-server /usr/bin/rok-server
EXPOSE 8080
ENV GOMEMLIMIT=100MiB
ENTRYPOINT [ "/usr/bin/rok-server" ]
