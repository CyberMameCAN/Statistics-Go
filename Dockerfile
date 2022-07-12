FROM golang:1.18-alpine

WORKDIR /go/src/app
COPY ./src .

RUN go get -u github.com/cosmtrek/air && \
    go build -o /go/bin/air github.com/cosmtrek/air && \
    go get github.com/gin-gonic/gin && \
    go get github.com/go-playground/validator

RUN go mod tidy
#RUN go mod download

RUN apk upgrade --update && \
    apk --no-cache add git

CMD ["air", "-c", ".air.toml"]