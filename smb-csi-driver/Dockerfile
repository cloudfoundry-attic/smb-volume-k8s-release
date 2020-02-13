FROM golang as builder
RUN mkdir /app
ADD . /app/
WORKDIR /app
RUN go build -o main -mod=vendor .

FROM golang
RUN apt update && apt -y install cifs-utils
RUN mkdir /app
COPY --from=builder /app/main /app/main
WORKDIR /app
CMD ["./main"]
