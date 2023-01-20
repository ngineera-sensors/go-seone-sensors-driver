FROM golang:1.19.4-alpine3.16
WORKDIR /var/lib/app

COPY ./go.mod ./go.sum ./
RUN go mod download

COPY ./sensors ./sensors
COPY ./main.go ./

RUN go build -o app

CMD ./app
