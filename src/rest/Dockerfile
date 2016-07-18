FROM golang:1.6-onbuild

RUN apt-get update
RUN apt-get install -qq -y npm

RUN npm install -g swagger-ui

EXPOSE 8080
