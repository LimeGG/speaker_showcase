FROM golang:latest

ENV GOPATH=/

WORKDIR /go/src/github.com/go-mysql

COPY . .

RUN go mod download
RUN go get
RUN go build -o cms ./main.go

CMD ["./cms"]