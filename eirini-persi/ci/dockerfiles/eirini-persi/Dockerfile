FROM golang:alpine as builder
RUN mkdir /app
ADD ../../.. /app/
WORKDIR /app
RUN CGO_ENABLED=0 go build -o "${BASEDIR}/binaries/eirini-persi" -ldflags="-X github.com/SUSE/eirini-persi/version.Version=dev" cmd/eirini-ext/main.go

FROM alpine
RUN mkdir /app
COPY --from=builder /app/binaries/eirini-persi /app/main
WORKDIR /app
CMD ["./main"]
