FROM golang:alpine as builder
RUN mkdir /app
ADD . /app/
WORKDIR /app
RUN go build -o main .

FROM alpine
RUN mkdir /app
COPY --from=builder /app/main /app/main
WORKDIR /app
CMD ["./main"]