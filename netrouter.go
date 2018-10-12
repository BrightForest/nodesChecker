package main

import (
	"encoding/json"
	"time"
)

type netRouter struct {
	request *networkMessage
}

type networkMessage struct {
	Action string `json:"action"`
	Content string `json:"content"`
}

func (netRouter *netRouter) ClientRouter(podConnect *NodePod, request *networkMessage){
	switch request.Action {
	case "ping":
		sendDataToClient(podConnect.PodNetworkChannel, &networkMessage{
			"pong",
			"",
		})
	case "registrationOk":
		podConnect.RegistrationIsOK = true
		go podConnect.NodeStateUpdater()
		go podConnect.NodesNetServiceLoad()
		go podConnect.NodesStateHandler()
	case "nodesInfo":
		go netRouter.NodesInfoUpdate(podConnect, request)
	default:
		sendDataToClient(podConnect.PodNetworkChannel, &networkMessage{"error", "notRecognized"})
	}
}

func (netRouter *netRouter) NodesInfoUpdate(podConnect *NodePod, request *networkMessage){
	var nodesMap = make(map[string]string)
	err := json.Unmarshal([]byte(request.Content), &nodesMap)
	checkErr(err)
	if err == nil{
		for podConnect.CuratorNodeIPMapLocked == true {
			time.Sleep(20*time.Millisecond)
		}
		podConnect.CuratorNodeIPMapLocked = true
		podConnect.CuratorNodeIPMap = nodesMap
		podConnect.CuratorNodeIPMapLocked = false
	} else {
		Warning.Println("Cannot update nodes info from Curator.")
	}
}

func (netRouter *netRouter) RequestsRouter(podConnect *NodePodConnection, request *networkMessage){
	switch request.Action {
	case "ping":
		netRouter.Ping(podConnect, request)
	default:
		netRouter.NotRecognized(podConnect)
	}
}

func (netRouter *netRouter) Ping(podConnect *NodePodConnection, message *networkMessage){
	NodePodStatesMap[podConnect.PodConf.NodeName] = 1
	sendDataToClient(podConnect.PodNetworkChannel, &networkMessage{
		"pong",
		"",
	})
}


func (netRouter *netRouter) NotRecognized(podConnect *NodePodConnection){
	networkAnswer := &networkMessage{"error", "notRecognized"}
	sendDataToClient(podConnect.PodNetworkChannel, networkAnswer)
}

func sendDataToClient(podNetworkChannel chan *networkMessage, message *networkMessage){
	podNetworkChannel <- message
}