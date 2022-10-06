FROM golang:stretch AS build-env
ADD . /go/src/example-project
RUN cd /go/src/example-project \
  && go get ./... \
  && CGO_ENABLED=0 GOOS=linux go build main.go

FROM busybox:1.29.3
LABEL maintainer="thycotic.com"
ENV CONTROLLER_SERVICE="localhost"
COPY --from=build-env /go/src/example-project /
CMD /main
