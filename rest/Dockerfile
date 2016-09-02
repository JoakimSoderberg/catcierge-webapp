FROM golang:1.6

RUN apt-get update
RUN apt-get install -qq -y npm

RUN npm install -g swagger-ui

RUN mkdir -p /go/src/app
WORKDIR /go/src/app

# Tool for dependencies.
RUN go get github.com/tools/godep

RUN mkdir -p /go/src/app/Godeps
COPY ./Godeps /go/src/app/Godeps

# Get dependencies.
RUN godep restore

EXPOSE 8080

COPY . /go/src/app
RUN godep go build -o $GOPATH/bin/catcierge-rest .

CMD ["catcierge-rest"]