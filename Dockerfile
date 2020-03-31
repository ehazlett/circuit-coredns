FROM golang:1.12-alpine as build
ENV GO111MODULE=on
RUN apk add -U --no-cache git mercurial gcc build-base
COPY . /go/src/circuit-coredns
RUN git clone https://github.com/coredns/coredns /go/src/coredns && \
	cd /go/src/coredns && \
	git checkout 1766568398e3120c85d44f5c6237a724248b652e
WORKDIR /go/src/coredns
RUN echo "circuit:github.com/ehazlett/circuit-coredns" >> plugin.cfg
RUN echo "replace github.com/ehazlett/circuit-coredns => /go/src/circuit-coredns" >> go.mod
RUN go generate
RUN go build -v

FROM scratch as binary
COPY --from=build /go/src/coredns/coredns /

FROM alpine:latest
COPY --from=build /go/src/circuit-coredns/corefile /etc/
COPY --from=build /go/src/coredns/coredns /bin/
CMD ["/bin/coredns", "-conf", "/etc/corefile"]
