FROM golang:1.6

RUN apt-get update
RUN apt-get install -qq -y npm

RUN npm install -g swagger-ui

RUN mkdir -p /go/src/app
WORKDIR /go/src/app

RUN mkdir -p /go/src/app/Godeps
COPY ./Godeps /go/src/app/Godeps

# Get dependencies
RUN go get github.com/tools/godep
RUN godep restore

COPY . /go/src/app
RUN godep go build .

#CMD ["go-wrapper", "run"]
CMD ["go", "run", "main.go"]

EXPOSE 8080
