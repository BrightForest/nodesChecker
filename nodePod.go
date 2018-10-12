package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"github.com/gorilla/websocket"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

var NodePodStatesMap = make(map[string]int)

type NodePod struct {
	IP string
	NodeName string
	kubernetesApiAddr string
	clusterAddr string
	authToken string
	clusterApiConnTimeout int
	curatorServiceAddr string
	PodNetworkChannel chan *networkMessage
	CuratorIsAvailable bool
	RegistrationIsOK bool
	CuratorConnectIsActive bool
	CuratorWSConn *websocket.Conn
	WSStates *WSReaderWriterState
	CuratorNodeIPMap map[string]string
	CuratorNodeIPMapLocked bool
	NodesEnvMap map[string]int
	NodesEnvMapIsLocked bool
	NodesStateChannel chan *NodeState
	Alerter *AlertService
}

type NodeState struct {
	NodeName string
	NodeState int
}

type NodeClientConnection struct {
	ConnectionName string
	ConnectionSocket *websocket.Conn
	RWState *WSReaderWriterState
	ConnectionIsActive bool
}

type WSReaderWriterState struct {
	ReaderReady bool
	WriterReady bool
}

type NodeMessage struct {
	IP string `json:"ip"`
	NodeName string `json:"nodename"`
	NodesAvailableMap map[string]int `json:"nodesAvailableMap"`
}

type NodePodSettings struct {
	IP string
	NodeName string
}

type NodePodConnection struct {
	ws *websocket.Conn
	PodConf *NodePodSettings
	PodIsAvailable bool
	PodNetworkChannel chan *networkMessage
}

func podLoader(config *Configuration, curatorServiceAddr string) *NodePod{
	var nodePod NodePod
	nodePod.kubernetesApiAddr = config.KubernetesApiAddr
	nodePod.clusterAddr = config.ClusterAddr
	nodePod.authToken = config.AuthToken
	nodePod.clusterApiConnTimeout = config.ClusterApiConnTimeout
	nodePod.curatorServiceAddr = curatorServiceAddr
	nodePod.PodNetworkChannel = make(chan *networkMessage)
	nodePod.CuratorIsAvailable = false
	nodePod.RegistrationIsOK = false
	nodePod.CuratorConnectIsActive = false
	nodePod.WSStates = &WSReaderWriterState{
		false,
		false,
	}
	nodePod.CuratorNodeIPMap = make(map[string]string)
	nodePod.CuratorNodeIPMapLocked = false
	nodePod.NodesEnvMap = make(map[string]int)
	nodePod.NodesStateChannel = make(chan *NodeState)
	nodePod.NodesEnvMapIsLocked = false
	nodePod.getPodIPEnv()
	nodePod.getPodNode()
	nodePod.Alerter = &AlertService{
		make(chan *AlertMessage),
	}
	go nodePod.Alerter.load(&nodePod)
	go nodePod.CuratorConnectionService()
	return &nodePod
}

func (nodePod *NodePod) getPodIPEnv() {
	conn, err := net.Dial("udp", nodePod.kubernetesApiAddr + ":443")
	checkErr(err)
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	nodePod.IP = localAddr.IP.String()
}

func (nodePod *NodePod) getPodNode() {
	message := map[string]interface{}{
		"pretty":"true",
	}

	bytesRepresentation, err := json.Marshal(message)
	if err != nil {
		Warning.Fatalln(err)
	}
	req, err := http.NewRequest("GET", nodePod.clusterAddr + "/api/v1/pods", bytes.NewBuffer(bytesRepresentation))
	req.Header.Add("Authorization", "Bearer " + nodePod.authToken)
	req.Header.Add("Accept", "application/json")
	tr := &http.Transport{
		IdleConnTimeout: 1000 * time.Millisecond * time.Duration(nodePod.clusterApiConnTimeout),
		TLSHandshakeTimeout: 1000 * time.Millisecond * time.Duration(nodePod.clusterApiConnTimeout),
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport:tr}
	resp, err := client.Do(req)
	if err != nil {
		Warning.Println("Error on response.\n[ERRO] -", err)
	}
	if resp != nil{
		body, _ := ioutil.ReadAll(resp.Body)
		var podsList PodList
		if err := json.Unmarshal(body, &podsList); err != nil {
			Warning.Println(err)
		}
		for _, pod := range podsList.Items{
			if pod.Metadata.Name == os.Getenv("HOSTNAME"){
				nodePod.NodeName = pod.Spec.NodeName
				break
			}
		}
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}
}

func (nodePod *NodePod) CuratorConnectionService(){
	if nodePod.CuratorWSConn != nil{
		nodePod.CuratorWSConn.Close()
	}
	Info.Println("Started CuratorConnectionService")
	nodePod.CuratorIsAvailable = false
	nodePod.RegistrationIsOK = false
	nodePod.CuratorConnectIsActive = false

	for nodePod.RegistrationIsOK == false{
		regConditions := true
		if checkTCPAvailable(nodePod.curatorServiceAddr) == false{
			regConditions = false
		}
		if regConditions == true && nodePod.ClientWSConnect() == true{
			nodePod.CuratorIsAvailable = true
			go nodePod.CuratorWriter(nodePod.CuratorWSConn)
			go nodePod.CuratorReader(nodePod.CuratorWSConn)
			for !nodePod.WSStates.WriterReady && !nodePod.WSStates.ReaderReady{
				time.Sleep(1000 * time.Millisecond * time.Duration(1))
			}
			if nodePod.WSStates.WriterReady && nodePod.WSStates.ReaderReady{
				go nodePod.ConnectController()
				go nodePod.RegistrationOnCurator(nodePod.CuratorWSConn)
				go nodePod.NodeScheduler()
			}
			break
		}
		time.Sleep(1000 * time.Millisecond * time.Duration(10))
	}
}

func (nodePod *NodePod) ClientWSConnect() bool{
	u := url.URL{Scheme: "ws", Host: nodePod.curatorServiceAddr, Path: "/ws"}
	clientConnect, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err == nil{
		nodePod.CuratorWSConn = clientConnect
		nodePod.CuratorConnectIsActive = true
		return true
	} else {
		Warning.Println("Unable to connect WS Socket to Curator.")
		return false
	}
}

func (nodePod *NodePod) ConnectController(){
	Info.Println("Start ConnectController")
	for nodePod.CuratorConnectIsActive == true{
		time.Sleep(1000 * time.Millisecond * time.Duration(3))
	}
	Info.Println("Registration service was restarted.")
	go nodePod.CuratorConnectionService()
}

func (nodePod *NodePod) CuratorWriter(curatorConnect *websocket.Conn){
	Info.Println("Start CuratorWriter")
	nodePod.WSStates.WriterReady = true
	for nodePod.CuratorConnectIsActive == true{
		if nodePod.CuratorIsAvailable == true {
			nodePodMessage := nodePod.checkClientMessageQueue()
			if nodePodMessage != nil{
				err := curatorConnect.WriteJSON(nodePodMessage)
				if err != nil{
					Warning.Println("Unable to send data to Curator pod.")
					nodePod.CuratorIsAvailable = false
					nodePod.CuratorConnectIsActive = false
				}
			}
		} else {
			Warning.Println("Unable to send data to Curator pod.")
			nodePod.CuratorIsAvailable = false
			nodePod.CuratorConnectIsActive = false
		}
	}
}

func (nodePod *NodePod) CuratorReader(curatorConnect *websocket.Conn){
	Info.Println("Start CuratorReader")
	requestsRouter := netRouter{};
	nodePod.WSStates.ReaderReady = true
	for nodePod.CuratorConnectIsActive == true{
		if nodePod.CuratorIsAvailable == true{
			getMessage := networkMessage{}
			err := curatorConnect.ReadJSON(&getMessage)
			if err != nil {
				Warning.Println("Unable to read curator data")
				nodePod.CuratorIsAvailable = false
				nodePod.CuratorConnectIsActive = false
			}
			requestsRouter.request = &getMessage
			requestsRouter.ClientRouter(nodePod, requestsRouter.request)
		}
	}
}


func (nodePod *NodePod) CuratorPinger(curatorConnect *websocket.Conn){
	Info.Println("Start CuratorPinger")
	for nodePod.CuratorConnectIsActive == true {
		time.Sleep(1000 * time.Millisecond * time.Duration(5))
		if nodePod.CuratorIsAvailable == true{
			nodePod.sendDataToCurator(&networkMessage{
				"ping",
				"",
			})
		} else {
			Warning.Println("Unable to ping curator.")
			nodePod.CuratorIsAvailable = false
			nodePod.CuratorConnectIsActive = false
		}
	}
}

func (nodePod *NodePod) RegistrationOnCurator(curatorConnect *websocket.Conn){
	Info.Println("Start Registration")
	var localNodeReg NodePodSettings
	localNodeReg.IP = nodePod.IP
	localNodeReg.NodeName = nodePod.NodeName
	podSettingsToCurator, err := json.Marshal(&localNodeReg)
	if err != nil{
		Warning.Println(err)
	}
	nodePod.sendDataToCurator(&networkMessage{"registration", string(podSettingsToCurator)})
}

func (nodePod *NodePod) NodeStateUpdater(){
	//Info.Println("Start NodeStateUpdater")
	if nodePod.RegistrationIsOK{
		var nodeMessage NodeMessage
		nodeMessage.IP = nodePod.IP
		nodeMessage.NodeName = nodePod.NodeName
		nodeMessage.NodesAvailableMap = make(map[string]int)
		for node, state := range nodePod.NodesEnvMap{
			nodeMessage.NodesAvailableMap[node] = state
		}
		message, err := json.Marshal(nodeMessage)
		checkErr(err)
		sendDataToClient(nodePod.PodNetworkChannel, &networkMessage{
			"nodeUpdate",
			string(message),
		})
	}
}

func (nodePod *NodePod) NodesNetServiceLoad(){
	Info.Println("Start NodesNetServiceLoad")
	for nodePod.CuratorConnectIsActive && nodePod.RegistrationIsOK {
		time.Sleep(5 * time.Second)
		if !nodePod.CheckNodesConnectionsCount(){
			nodePod.NodesNetServiceConnStarter()
		}
	}
}

func (nodePod *NodePod) CheckNodesConnectionsCount() bool{
	for nodePod.NodesEnvMapIsLocked == true{
		time.Sleep(20*time.Millisecond)
	}
	nodePod.NodesEnvMapIsLocked = true
	curatorNodesCount := len(nodePod.CuratorNodeIPMap)
	nodePod.NodesEnvMapIsLocked = false
	if curatorNodesCount != len(nodePod.NodesEnvMap){
		return false
	} else {
		return true
	}
}

func (nodePod *NodePod) NodesNetServiceConnStarter(){
	Info.Println("Start NodesNetServiceConnStarter")
	var CuratorNodesList = make(map[string]string)
	for nodePod.CuratorNodeIPMapLocked == true {
		time.Sleep(20*time.Millisecond)
	}
	nodePod.CuratorNodeIPMapLocked = true
	for key, value := range nodePod.CuratorNodeIPMap{
		CuratorNodesList[key] = value
	}
	nodePod.CuratorNodeIPMapLocked = false
	var AvailableConnectionsList = make(map[string]int)
	for nodePod.NodesEnvMapIsLocked == true{
		time.Sleep(20*time.Millisecond)
	}
	nodePod.NodesEnvMapIsLocked = true
	for key, value := range nodePod.NodesEnvMap{
		AvailableConnectionsList[key] = value
	}
	nodePod.NodesEnvMapIsLocked = false
	for nodeName, nodeIp := range CuratorNodesList{
		if _, ok := AvailableConnectionsList[nodeName]; !ok {
			Info.Println("Initiate connection to node", nodeName, "with IP", nodeIp)
			go InitiateNodeNetConnection(nodeName, nodeIp, nodePod)
		}
	}
}

func InitiateNodeNetConnection(nodeName string, nodeIp string, nodePod *NodePod){
	Info.Println("Connect to", nodeIp + ":8080")
	if checkTCPAvailable(nodeIp + ":8080"){
		clientConnect := NodeClientConnection{ConnectionName:nodeName, RWState:&WSReaderWriterState{
			false,
			false,
		}}
		go clientConnect.ClientConnectionService(nodeIp, nodePod)
	} else {
		Warning.Println("Node ", nodeName, "not available by tcp.")
	}
}

func (nodeClientConnection *NodeClientConnection) ClientConnectionService(nodeIp string, nodePod *NodePod){
	Info.Println("Started ClientConnectionService.")
	connectionURL := nodeIp + ":8080"
	u := url.URL{Scheme: "ws", Host: connectionURL, Path: "/ws"}
	clientConnect, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err == nil{
		nodeClientConnection.ConnectionIsActive = true
		nodeClientConnection.ConnectionSocket = clientConnect
		go nodeClientConnection.ClientConnectionReader(nodePod)
	} else {
		Error.Println(err)
	}
}

func (nodeClientConnection *NodeClientConnection) ClientConnectionReader(nodePod *NodePod){
	Info.Println("Started ClientConnectionReader for node", nodeClientConnection.ConnectionName)
	nodeClientConnection.RWState.ReaderReady = true
	nodeConnMessageActive := &NodeState{
		nodeClientConnection.ConnectionName,
		1,
	}
	nodePod.NodesStateChannel <- nodeConnMessageActive
	alertMessage := &AlertMessage{
		nodePod.NodeName,
		nodeClientConnection.ConnectionName,
		"",
	}
	for nodeClientConnection.ConnectionIsActive == true {
		getMessage := networkMessage{}
		err := nodeClientConnection.ConnectionSocket.ReadJSON(&getMessage)
		if err != nil {
			alertMessage.DateTime = getDateTime()
			nodePod.Alerter.AlertChannel <- alertMessage
			nodeConnMessage := &NodeState{
				nodeClientConnection.ConnectionName,
				0,
			}
			nodePod.NodesStateChannel <- nodeConnMessage
			nodeClientConnection.RWState.ReaderReady = false
			nodeClientConnection.ConnectionIsActive = false
			break
		}
	}
}

func (nodePod *NodePod) NodesStateHandler(){
	for{
		select{
			case message := <- nodePod.NodesStateChannel:
				go nodePod.UpdateNodeEnvMap(message)
		}
	}
}

func (nodePod *NodePod) UpdateNodeEnvMap(nodeState *NodeState){
	for nodePod.NodesEnvMapIsLocked == true{
		time.Sleep(20*time.Millisecond)
	}
	nodePod.NodesEnvMapIsLocked = true
	if nodeState.NodeState == 1 {
		nodePod.NodesEnvMap[nodeState.NodeName] = nodeState.NodeState
	} else {
		delete(nodePod.NodesEnvMap, nodeState.NodeName)
	}
	nodePod.NodesEnvMapIsLocked = false
	nodePod.NodeStateUpdater()
}

func (nodePod *NodePod) checkClientMessageQueue() *networkMessage{
	select{
	case message := <-nodePod.PodNetworkChannel:
		return message
	}
}

func (nodePod *NodePod)sendDataToCurator(message *networkMessage){
	nodePod.PodNetworkChannel <- message
}

func (nodePod *NodePod) NodeScheduler(){
	for {
		time.Sleep(1000 * time.Millisecond * time.Duration(10))
		nodePod.NodeStateUpdater()
	}
}

func checkTCPAvailable(addr string) bool{
	conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
	if err != nil{
		Warning.Println(addr, "is not available by TCP.")
		if conn != nil{
			conn.Close()
		}
		return false
	} else {
		if conn != nil{
			conn.Close()
		}
		return true
	}
}

func getDateTime() string{
	Time := time.Now()
	dateTimeString := Time.Format("02.01.2006 15:04:05.") + strconv.Itoa(Time.Nanosecond())
	return dateTimeString
}