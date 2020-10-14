FROM golang:alpine

WORKDIR /go/src/reversetls
COPY . .
RUN go install

ENTRYPOINT ["reversetls"]
