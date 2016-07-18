FROM golang:1.6

RUN apt-get update
RUN apt-get install -qq -y npm

RUN npm install -g swagger-ui

RUN mkdir -p /go/src/app
WORKDIR /go/src/app

COPY . /go/src/app
RUN go-wrapper download
RUN go-wrapper install

CMD ["go-wrapper", "run"]

EXPOSE 8080
