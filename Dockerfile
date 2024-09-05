
FROM golang:1.22 AS builder


WORKDIR /app


COPY go.mod go.sum ./
RUN go mod tidy


COPY . .


RUN go build -o main ./cmd


FROM golang:1.22


WORKDIR /app


COPY --from=builder /app/main .


COPY ./migrations /app/migrations


EXPOSE 8080


CMD ["./main"]
