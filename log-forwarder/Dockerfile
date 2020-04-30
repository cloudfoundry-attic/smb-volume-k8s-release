FROM golang:stretch as builder
RUN mkdir /app
ADD . /app/
WORKDIR /app
RUN go build -o main .

FROM ubuntu
RUN mkdir /app
COPY --from=builder /app/main /app/main
WORKDIR /app
CMD ["./main"]