FROM golang:1.19-alpine

WORKDIR /go/src/app
COPY ./src .

RUN apk upgrade --update && \
    apk --no-cache add git

RUN go get github.com/gin-gonic/gin
RUN go get github.com/go-playground/validator
RUN go get github.com/dgrijalva/jwt-go
RUN go get github.com/joho/godotenv
RUN go get -u github.com/cosmtrek/air && \
    go build -o /go/bin/air github.com/cosmtrek/air

# RUN go mod tidy
RUN go mod download

CMD ["air", "-c", ".air.toml"]