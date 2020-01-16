FROM golang:alpine as builder
RUN mkdir /app
ADD . /app/
WORKDIR /app
RUN CGO_ENABLED=0 go build -o "${BASEDIR}/binaries/eirini-ext" -ldflags="-X github.com/SUSE/eirini-persi/version.Version=dev" cmd/eirini-ext/main.go

FROM ubuntu:xenial
RUN mkdir /app
COPY --from=builder /app/binaries/eirini-ext /app/main
WORKDIR /app
CMD ["./main"]
