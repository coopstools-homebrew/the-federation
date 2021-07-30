FROM golang:1.16-alpine AS builder
WORKDIR /go/src
COPY . .
RUN go build -o api

FROM dtzar/helm-kubectl:3.6.3
WORKDIR /home
COPY --from=builder /go/src/api .

CMD ./api $PORT $URL_PATH_PREFIX
