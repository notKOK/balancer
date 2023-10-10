FROM golang:1.19

WORKDIR /

COPY go.mod go.sum ./

RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o app

EXPOSE 8000

CMD ["chmod a+x ./app"]


CMD ["/app"]
