FROM golang:latest 

ENV p /go/src/github.com/bruno-nascimento/mock-url

RUN mkdir -p ${p}
ADD . ${p}
WORKDIR ${p}
RUN go get -v ./...

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /main

FROM scratch
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=0 /main /
CMD ["/main"]