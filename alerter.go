package main

import "encoding/json"

type AlertMessage struct {
	NodeName string `json:"nodename"`
	DisconnectedNodeName string `json:"disconnectednode"`
	DateTime string `json:"datetime"`
}

type AlertService struct {
	AlertChannel chan *AlertMessage
}

func (alertService *AlertService) load(curatorPod *NodePod){
	for{
		select {
			case message := <- alertService.AlertChannel:
				alertJsonMessage, err := json.Marshal(message)
				checkErr(err)
				sendDataToClient(curatorPod.PodNetworkChannel, &networkMessage{
					"alert",
					string(alertJsonMessage),
				})
				Warning.Println(message.NodeName, "was lost contact with", message.DisconnectedNodeName, "Time:", message.DateTime)
		}
	}
}