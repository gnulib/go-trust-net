# use this base image the first time tand tag it as p2p-pager image for subsequent builds
#FROM golang

# use this base image for incremental build on top of last p2p-pager image
FROM p2p-pager

# cleanup older code in base image (e.g. when directory layout changes)
RUN rm -rf /go/src/github.com/trust-net/go-trust-net

# copy current source codebase
ADD . /go/src/github.com/trust-net/go-trust-net

# install dependencies on base image (e.g. when using golang)
RUN go get github.com/ethereum/go-ethereum
RUN go get github.com/syndtr/goleveldb/leveldb
RUN go get github.com/satori/go.uuid

# build and install app
RUN go clean -i -x
RUN go install github.com/trust-net/go-trust-net/app

ENTRYPOINT /go/bin/app

EXPOSE 30303
