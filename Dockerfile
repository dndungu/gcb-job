FROM golang:1.11 as builder

LABEL maintainer "David Ndungu <dnjuguna@gmail.com>"

ENV GOPATH /go

WORKDIR /go/src/github.com/zendesk/gcb-stage

COPY . .

ARG SOURCE_COMMIT

RUN go get -u github.com/golang/dep/cmd/dep

RUN dep ensure

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
	-o /gcb-stage -ldflags "-X main.GitRevision=${SOURCE_COMMIT}" .

FROM scratch

LABEL maintainer "David Ndungu <dndungu@zendesk.com>"

ENV GRIFFON_PORT 80

WORKDIR /bin

ADD config.yml /etc/gcb-stage.yml

COPY --from=builder /gcb-stage .

EXPOSE ${GRIFFON_PORT}

ENTRYPOINT ["/bin/gcb-stage"]
