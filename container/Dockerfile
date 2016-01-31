FROM golang

RUN go get github.com/jdiez17/masterrr
RUN go build github.com/jdiez17/masterrr

ENTRYPOINT /go/bin/masterrr
EXPOSE 8082
