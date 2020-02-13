FROM golang as builder
RUN mkdir /app
ADD . /app/
WORKDIR /app
RUN go build -o main -mod=vendor .

RUN apt update && apt -y install cifs-utils

FROM golang
RUN mkdir /app
COPY --from=builder /app/main /app/main
WORKDIR /app
CMD ["./main"]
