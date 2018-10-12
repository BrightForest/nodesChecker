package main

import (
	"net/http"
	"log"
	"github.com/gorilla/websocket"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func wsHandler(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize: 2048,
		WriteBufferSize: 2048,
	}
	ws, err := upgrader.Upgrade(w,r, nil)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		return
	}

	pConn := &NodePodConnection{ws: ws, PodIsAvailable: false, PodNetworkChannel: make(chan *networkMessage, 10)}
	go getPodData(pConn)
}

func getPodData(podConnect *NodePodConnection){
	requestsRouter := netRouter{};
	go writeServerDataLoop(podConnect)
	for{
		getMessage := networkMessage{}
		err := podConnect.ws.ReadJSON(&getMessage)
		if err != nil {
			Info.Println("Pod was disconnected.")
			if podConnect.PodIsAvailable == true{
				podConnect.PodIsAvailable = false
				NodePodStatesMap[podConnect.PodConf.NodeName] = 0
			}
			podConnect.ws.Close()
			break
		}
		requestsRouter.request = &getMessage
		requestsRouter.RequestsRouter(podConnect, requestsRouter.request)
	}
}

func writeServerDataLoop(podConnect *NodePodConnection){
	for{
		myPodMessage := checkMyPodConnections(podConnect)
		if myPodMessage != nil{
			//			Info.Println(myPodMessage.Action)
			//			Info.Println(myPodMessage.Content)
			err := podConnect.ws.WriteJSON(myPodMessage)
			if err != nil{
				podConnect.PodIsAvailable = false
				Info.Println("Unable to write in WS Socket.")
				break
			}
		}
	}
}

func checkMyPodConnections(podConnect *NodePodConnection) *networkMessage{
	select {
	case message := <- podConnect.PodNetworkChannel:
		return message
	}
}


func StartWSServer(nodeIP string)  {
	http.HandleFunc("/ws", wsHandler)
	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)
	go gracefulShutdownReciever(gracefulStop)
	if err := http.ListenAndServe(nodeIP + ":8080", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func gracefulShutdownReciever(osChannel chan os.Signal){
	sig := <-osChannel
	Info.Printf("caught sig: %+v", sig)
	Info.Println("Wait for 2 second to finish processing")
	time.Sleep(2*time.Second)
	os.Exit(0)
}