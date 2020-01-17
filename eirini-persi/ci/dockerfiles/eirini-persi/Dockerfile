FROM golang:alpine as builder
RUN mkdir /app
ADD https://github.com/SUSE/eirini-persi/archive/master.tar.gz /app/eirini-persi.tar.gz
RUN cd /app/ && tar xvfz eirini-persi.tar.gz
WORKDIR /app/eirini-persi-master
RUN CGO_ENABLED=0 go build -o "./binaries/eirini-persi" -ldflags="-X github.com/SUSE/eirini-persi/version.Version=dev" cmd/eirini-ext/main.go

FROM alpine
RUN mkdir /app
COPY --from=builder /app/eirini-persi-master/binaries/eirini-persi /app/main
WORKDIR /app
CMD ["./main"]
