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

COPY . /go/src/app
#RUN godep go build .

# TODO: Make this nicer so we don't have to list all files...
CMD ["go", "run", "main.go", "cat_event.go", "unzip.go"]

EXPOSE 8080
