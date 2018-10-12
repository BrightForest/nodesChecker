FROM golang:1.10
RUN mkdir /spidernode
ADD . /spidernode/
WORKDIR /spidernode
RUN go get "github.com/gorilla/websocket"
RUN go build -o spidernode .
RUN groupadd -r spidernode && useradd -r -g spidernode spidernode
USER spidernode
EXPOSE 8080
CMD ["/spidernode/spidernode"]