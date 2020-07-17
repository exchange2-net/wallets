package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"github.com/sendgrid/sendgrid-go"
	"sync"
	"wallets/walletsP/LocalConfig"
	"wallets/walletsP/Pairs"
	zmq "github.com/pebbe/zmq4"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var Fee uint64
var rcvTransactionAnswer = make(map[int]*zmq.Socket)
var rcvWalletsAuth *zmq.Socket
var rcvCheckBalance *zmq.Socket
var sndCheckBalance *zmq.Socket
var sndNewUsrAddr = make(map[int]*zmq.Socket)
var sndNewTransaction = make(map[int]*zmq.Socket)
var rcvTransactionHistory = make(map[int]*zmq.Socket)
var sendWalletsMail *zmq.Socket

var TMPMailData struct {
	data []byte
	mux sync.Mutex
}

var Wclients struct {
	data map[string]Session
	mux sync.Mutex //mutex for a concurrent map
}

var WclientsId struct {
	data map[uint64]Session
	mux sync.Mutex //mutex for a concurrent map
}

var WPorts[1000]Pairs.WalletType
var WTypeAddr[1000]Pairs.WalletType

var FreeCoinAdresses struct {
	data map[string]FreeAddresses
	mux sync.Mutex
}

var UsersHasAddr struct {
	data map[uint64]HasAddress
	mux sync.Mutex
}

var addrSendMap struct {
	data map[uint64]dataToSend
	mux sync.Mutex
}

// In/OUT Transactions history
var usersTrunsactionsIn struct {
	data map[string] UserTrunsaction
	mux sync.Mutex
}

var usersTrunsactionsOUT struct {
	data map[string] UserTrunsaction
	mux sync.Mutex
}

// In/OUT Transactions history By User Addr.
var WalletTrHistoryIn struct {
	data map[string]HistoryByAddr
	mux sync.Mutex
}

var WalletTrHistoryOUT struct {
	data map[string]HistoryByAddr
	mux sync.Mutex
}

//COLD WALLET HISTORY
var ColdWalletTrunsactionsIN = make(map[string] UserTrunsaction)
var ColdWalletTrunsactionsOUT = make(map[string] UserTrunsaction)

var AllHistoryBin struct {
	data []byte
	mux sync.Mutex
}

var HistoryBTC []byte
var HistoryETH []byte

//----------
type HistoryByAddr map[string]UserTrunsaction

type mailStruct struct {
	mail string
	sndFrom string
	sndTo string
	value uint64
	CoinID uint64
	InOut uint64
}

type dataToSend = map[uint64] struct {
	sndFrom string
	sndTo string
	CoinID uint64
}

type HasAddress map[string]struct {
	CoinID uint64
	PublicKey string
}
type TMPstruct struct {
	CoinID uint64
	PublicKey string
}
type uniqDataType struct {
	CoinID uint64
	TrTime uint64
	TrValue string
}
type UserTrunsaction struct {
	UserId uint64
	TransactionHash string
	TrTime uint64
	InOut uint64
	ConfirmedBlocks uint64
	TrFrom string
	TrTo string
	TrValue string
	CoinID uint64
}

type FreeAddresses  map[string]string

type Session struct {
	userId uint64 //make 10 words??
	userSession string //TODO protect from oversize
	time int64
	email string
}

var usersChannels  *UsersChannels

type UsersChannels struct {
	sendHTTPresponse chan []byte
}

var fileOL *os.File
var file200 *os.File
var file1000 *os.File
var file10000 *os.File
type Storage struct{
	file200 *os.File
	file1000 *os.File
	file10000 *os.File
	name string
	historyRegistry [1000000] Records //designed for one million users
}
type Records struct {
	pos200 int64
	pos1000 int64
	pos10000 int64
	records uint64
	file uint8
}
var TransactionsHistory Storage
var TransactionsHistoryETH Storage
var TransactionsHistoryLTC Storage

func main() {
	//make maps without any nil data
	Wclients.data = make(map[string]Session)
	WclientsId.data = make(map[uint64]Session)
	FreeCoinAdresses.data = make(map[string]FreeAddresses)
	UsersHasAddr.data = make(map[uint64]HasAddress)
	addrSendMap.data = make(map[uint64]dataToSend)
	usersTrunsactionsIn.data = make(map[string] UserTrunsaction)
	usersTrunsactionsOUT.data = make(map[string] UserTrunsaction)
	WalletTrHistoryIn.data = make(map[string]HistoryByAddr)
	WalletTrHistoryOUT.data = make(map[string]HistoryByAddr)
	//----
	globalCounter = 0
	Fee = 0
	usersChannels = &UsersChannels {
		sendHTTPresponse: make(chan []byte, 1000000),
	}
	WPorts = Pairs.WalletList
	WTypeAddr = Pairs.WalletList

	//loading users' coin address list
	loadUserAddr("BTC")
	loadUserAddr("ETH")

	//loading the list of free coins
	coinsBTC := [3]string{"Bitcoin", "Litecoin", "Ravecoin"}
	for _, dataToLoad := range coinsBTC {
		getFreeAddress(dataToLoad, "BTC")
	}
	coinsETH := [1]string{"Ethereum"}
	//coinsETH := [3]string{"Ethereum", "Ether1", "EthClassic"}
	for _, dataToLoad := range coinsETH {
		getFreeAddress(dataToLoad, "ETH")
	}

	//creating a connection with Authorization server (receive)
	//receive data from Authorization Server
	rcvWalletsAuth, _ = zmq.NewSocket(zmq.PULL)
	defer rcvWalletsAuth.Close()
	rcvWalletsAuth.SetRcvhwm(1100000)
	rcvWalletsAuthAddress := fmt.Sprintf("%v%v", LocalConfig.LocalTcpIpAddr, LocalConfig.WalletAuthPort)
	rcvWalletsAuth.Connect(rcvWalletsAuthAddress)

	//Creating a connection with "IN" server (receive)
	var rcvTransactionHistoryAddresses = make(map[int]string)
	ij := 1 // like Coin ID
	for i := LocalConfig.WalletrcvHistoryBegin; i <= LocalConfig.WalletrcvHistoryEnd; i++ {
		rcvTransactionHistory[ij],_ = zmq.NewSocket(zmq.PULL)
		defer rcvTransactionHistory[ij].Close()
		rcvTransactionHistory[ij].SetRcvhwm(1100000)
		rcvTransactionHistoryAddresses[ij] = fmt.Sprintf("%v%v", LocalConfig.LocalTcpIpAddr, i)
		rcvTransactionHistory[ij].Connect(rcvTransactionHistoryAddresses[ij])
		time.Sleep(10*time.Millisecond)
		ij++
	}

	//Creating a connection with "IN" servers (send)
	//sending data to "IN" servers. Info about new user address
	j := 1
	var addressMap = make(map[int]string)
	for i := LocalConfig.WalletaddressMapBegin; i <= LocalConfig.WalletaddressMapEnd; i++  {
		sndNewUsrAddr[j],_ = zmq.NewSocket(zmq.PUSH)
		defer sndNewUsrAddr[j].Close()
		sndNewUsrAddr[j].SetRcvhwm(1100000)
		addressMap[j] = fmt.Sprintf("%v%v", LocalConfig.LocalTcpIpAddr, i)
		sndNewUsrAddr[j].Bind(addressMap[j])
		j++
		time.Sleep(10*time.Millisecond)
	}

	//Creating a connection with "OUT" servers (send)
	//sending data to "OUT" servers.
	//Sending a request for making an  OUT transaction. Sending coins to other addresses
	j = 1
	var addressTrMap = make(map[int]string)
	for i := LocalConfig.WalletaddressTrMapBegin; i <= LocalConfig.WalletaddressTrMapEnd; i++  {
		sndNewTransaction[j],_ = zmq.NewSocket(zmq.PUSH)
		defer sndNewTransaction[j].Close()
		sndNewTransaction[j].SetRcvhwm(1100000)
		addressTrMap[j] = fmt.Sprintf("%v%v", LocalConfig.LocalTcpIpAddr, i)
		sndNewTransaction[j].Bind(addressTrMap[j])
		j++
		time.Sleep(10*time.Millisecond)
	}

	//Bind ports for http requests
	var Wservers = map[int] *http.ServeMux{}
	for i := 0; i <= 10; i++  {
		Wservers[i] = http.NewServeMux()
		Wservers[i].HandleFunc("/", handleWconn)
		port := 20000+i
		server_port := fmt.Sprintf(":%v", port)
		go func() {
			http.ListenAndServe(server_port, Wservers[i])
		}()
		time.Sleep(10*time.Millisecond)
	}

	//Creating a  connection with a balance server (receive)
	rcvCheckBalance, _ = zmq.NewSocket(zmq.PULL)
	defer rcvCheckBalance.Close()
	rcvCheckBalance.SetRcvhwm(1100000)
	rcvCheckBalance.Connect(fmt.Sprintf("%v%v", LocalConfig.LocalTcpIpAddr, LocalConfig.WalletrcvCheckBalancePort))

	//Creating a connection with balance server (send)
	sndCheckBalance, _ = zmq.NewSocket(zmq.PUSH)
	defer sndCheckBalance.Close()
	sndCheckBalance.SetRcvhwm(1100000)
	sndCheckBalance.Bind(fmt.Sprintf("%v%v", LocalConfig.LocalTcpIpAddr, LocalConfig.WalletsndCheckBalancePort))

	//Creating a connection with Mail server (send)
	sendWalletsMail ,_ = zmq.NewSocket(zmq.PUSH)
	defer sendWalletsMail.Close()
	sendWalletsMail.SetRcvhwm(1100000)
	sendWalletsMailAddress := fmt.Sprintf("%v%v", LocalConfig.LocalTcpIpAddr, LocalConfig.WalletsendMailPort)
	sendWalletsMail.Bind(sendWalletsMailAddress);

	//Creating a connection with OUT servers (receive)
	ij = 0
	for i := 7410; i <= 7419; i++  {
		rcvTransactionAnswer[ij], _ = zmq.NewSocket(zmq.PULL)
		defer rcvTransactionAnswer[ij].Close()
		rcvTransactionAnswer[ij].SetRcvhwm(1100000)
		rcvTransactionAnswer[ij].Connect(fmt.Sprintf("%v%v", LocalConfig.LocalTcpIpAddr, i))
		ij++
	}

	//Loading data from Bitcoin Transactions Storage
	var err error
	TransactionsHistory.file200, err = os.OpenFile("../Bitcoin/transactionsHistory200.db", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer TransactionsHistory.file200.Close()

	TransactionsHistory.file1000, err = os.OpenFile("../Bitcoin/transactionsHistory1000.db", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer TransactionsHistory.file1000.Close()

	TransactionsHistory.file10000, err = os.OpenFile("../Bitcoin/transactionsHistory10000.db", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer TransactionsHistory.file10000.Close()
	TransactionsHistory.name = "transactionsHistory"

	//Loading data from Ethereum Transactions Storage
	TransactionsHistoryETH.file200, err = os.OpenFile("../Ethereum/transactionsHistory200.db", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer TransactionsHistoryETH.file200.Close()

	TransactionsHistoryETH.file1000, err = os.OpenFile("../Ethereum/transactionsHistory1000.db", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer TransactionsHistoryETH.file1000.Close()

	TransactionsHistoryETH.file10000, err = os.OpenFile("../Ethereum/transactionsHistory10000.db", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer TransactionsHistoryETH.file10000.Close()
	TransactionsHistoryETH.name = "transactionsHistoryETH"

	//Loading data from Litecoin Transactions Storage
	TransactionsHistoryLTC.file200, err = os.OpenFile("../Litecoin/transactionsHistory200.db", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer TransactionsHistoryLTC.file200.Close()

	TransactionsHistoryLTC.file1000, err = os.OpenFile("../Litecoin/transactionsHistory1000.db", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer TransactionsHistoryLTC.file1000.Close()

	TransactionsHistoryLTC.file10000, err = os.OpenFile("../Litecoin/transactionsHistory10000.db", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer TransactionsHistoryLTC.file10000.Close()
	TransactionsHistoryLTC.name = "transactionsHistory"

	//Read data from Ethereum and Bitcoin Transactions Storages
	readMapData(&TransactionsHistoryETH, "Ethereum")
	readMapData(&TransactionsHistory, "Bitcoin")
	readMapData(&TransactionsHistoryLTC, "Litecoin")

	fmt.Println("Servers is loaded")
	go reciveTransactionAnswer()
	go sendToMailServer()
	go checkWalletsAuth()
	go reciveBalanceAnswer()
	reciveTrHistory()
}


func subscribe(mail string) int{
	apiKey := LocalConfig.SendGridPassw
	host := "https://api.sendgrid.com"
	request := sendgrid.GetRequest(apiKey, "/v3/marketing/contacts", host)
	request.Method = "PUT"
	data := fmt.Sprintf("{\"list_ids\": [\"%v\"],\"contacts\": [{\"email\": \"%v\"}]}",LocalConfig.SendGridListId, mail)
	request.Body = []byte(data)

	response, err := sendgrid.API(request)
	if err != nil {
		log.Println(err)
		return 0
	} else {
		fmt.Println(response.StatusCode)
		if response.StatusCode == 202 {
			return 1
		}
	}
	return 0
}
/*
function readMapData
args: storage, coin type
the function is to read data from the storage and to write it to binary array
 */
func readMapData(dataStorage *Storage, Coin string) {
	var userId uint64
	Path := fmt.Sprintf("../%v/%vMap.db",Coin, dataStorage.name)

	file, err := os.Open(Path)
	defer file.Close()

	if err != nil {
		return
	}

	fi, err := file.Stat()
	if err != nil {
		// Failed getting stat, handle error
		log.Fatal("Failed getting stat ", err)
	}

	var i uint64

	for i = 0; i < uint64(fi.Size()/48); i++ {
		data := readNextBytes(file, 48)

		if (len(data)<48){
			break
		}
		if (i == uint64(len(dataStorage.historyRegistry))) {
			//wrong size
			fmt.Println(len(dataStorage.historyRegistry), uint64(fi.Size()/48))
			break
		}

		userId = binary.LittleEndian.Uint64(data[40:48])
		fmt.Println(userId)
		if userId <= 2 {
			//Developers debug IDs
			continue
		}

		//writing data to history Registry
		dataStorage.historyRegistry[userId].pos200 = int64(binary.LittleEndian.Uint64(data[0:8]))
		dataStorage.historyRegistry[userId].pos1000 = int64(binary.LittleEndian.Uint64(data[8:16]))
		dataStorage.historyRegistry[userId].pos10000 = int64(binary.LittleEndian.Uint64(data[16:24]))
		dataStorage.historyRegistry[userId].records = binary.LittleEndian.Uint64(data[24:32])
		dataStorage.historyRegistry[userId].file = uint8(binary.LittleEndian.Uint64(data[32:40]))

		var new_read_offset uint64
		var read_strings uint64

		//how many records will be read from the  storage
		if dataStorage.historyRegistry[userId].records >= 200 {
			read_strings = 200
		} else {
			read_strings = dataStorage.historyRegistry[userId].records
		}

		if dataStorage.name == "transactionsHistoryETH" {
			//select read offset
			if dataStorage.historyRegistry[userId].file == 1 {
				new_read_offset = uint64(0)
			}
			if dataStorage.historyRegistry[userId].file == 2 {
				new_read_offset = uint64(200)
			}
			if dataStorage.historyRegistry[userId].file == 3 {
				new_read_offset = uint64(10000)
			}
			if dataStorage.historyRegistry[userId].file == 4 {
				new_read_offset = uint64(10000)
			}

			//reading data from the storage
			HistoryETH = readFileData(userId,read_strings, new_read_offset, dataStorage, Coin)

			var offset int
			offset = 0

			//converting data to the general format
			for n:=0; n < (len(HistoryETH)/192);n++ {

				Uid := binary.LittleEndian.Uint64(HistoryETH[0+offset:8+offset])
				InOut := binary.LittleEndian.Uint16(HistoryETH[8+offset:10+offset])
				CoinID := binary.LittleEndian.Uint16(HistoryETH[10+offset:12+offset])
				ConfirmedBlocks := binary.LittleEndian.Uint16(HistoryETH[12+offset:14+offset])
				TrTime := binary.LittleEndian.Uint64(HistoryETH[14+offset:22+offset])
				TrValue2 := binary.LittleEndian.Uint64(HistoryETH[22+offset:30+offset])
				TrValue := fmt.Sprintf("%v", TrValue2)
				TrTo := string(bytes.Trim(HistoryETH[30+offset:72+offset], "\x00"))
				TrFrom := string(bytes.Trim(HistoryETH[72+offset:114+offset], "\x00"))
				TransactionHash := string(bytes.Trim(HistoryETH[114+offset:192+offset], "\x00"))

				var tmpBuf= make([]byte, 192)

				binary.LittleEndian.PutUint64(tmpBuf[0:8], Uid)
				binary.LittleEndian.PutUint16(tmpBuf[8:10], uint16(InOut))
				binary.LittleEndian.PutUint16(tmpBuf[10:12], uint16(CoinID))
				binary.LittleEndian.PutUint16(tmpBuf[12:14], 0)
				binary.LittleEndian.PutUint16(tmpBuf[14:16], uint16(ConfirmedBlocks))
				binary.LittleEndian.PutUint64(tmpBuf[16:24], TrTime)

				copy(tmpBuf[24:66], []byte(TrTo))
				copy(tmpBuf[66:108], []byte(TrFrom))
				copy(tmpBuf[108:126], []byte(TrValue))
				copy(tmpBuf[126:192], []byte(TransactionHash))

				if len(AllHistoryBin.data) <= 1 {
					AllHistoryBin.mux.Lock()
					AllHistoryBin.data = tmpBuf
					AllHistoryBin.mux.Unlock()
				} else {
					AllHistoryBin.mux.Lock()
					AllHistoryBin.data = append(AllHistoryBin.data, tmpBuf...)
					AllHistoryBin.mux.Unlock()
				}
				offset = offset + 192
			}
		}

		if dataStorage.name == "transactionsHistory" {
			//select read offset
			if dataStorage.historyRegistry[userId].file == 1 {
				new_read_offset = uint64(0)
			}
			if dataStorage.historyRegistry[userId].file == 2 {
				new_read_offset = uint64(200)
			}
			if dataStorage.historyRegistry[userId].file == 3 {
				new_read_offset = uint64(10000)
			}
			if dataStorage.historyRegistry[userId].file == 4 {
				new_read_offset = uint64(10000)
			}

			HistoryBTC = []byte{}
			//reading data from the storage
			HistoryBTC = readFileData(userId,read_strings, new_read_offset, dataStorage, Coin)

			if len(AllHistoryBin.data) <= 1 {
				AllHistoryBin.data = HistoryBTC
			} else {
				AllHistoryBin.data = append(AllHistoryBin.data, HistoryBTC...)
			}
		}
	}
}

/*
function readFileDataOffset
args: User ID, how many records will be read from the storage, read offset, storage
function returns read offset
 */
func readFileDataOffset(userId uint64, amount uint64, offset uint64, dataFiles *Storage) (uint64, uint64, uint64, uint64, uint64, uint64, uint64, uint64) {
	recordData := &dataFiles.historyRegistry[userId]

	var begin uint64

	if offset == 0 {
		begin = 0;
	} else {
		begin = recordData.records - offset + 1
	}

	var readOffset200 uint64
	var readOffset1000 uint64
	var readOffset10000 uint64
	var readOffsetFile uint64
	var read200 uint64
	var read1000 uint64
	var read10000 uint64
	var readFile uint64


	if(begin < 200) {
		readOffset200 = begin
		if(begin + amount < 200) {
			read200 = amount
			return readOffset200, read200, readOffset1000, read1000, readOffset10000, read10000, readOffsetFile, readFile
		} else {
			read200 = 200 - begin
			readOffset1000 = 0
		}
	}

	if(begin < 1200) {
		if(begin > 200) {
			readOffset1000 = begin - 200
		}

		if(begin + amount < 1200) {
			read1000 = amount-read200
			return readOffset200, read200, readOffset1000, read1000, readOffset10000, read10000, readOffsetFile, readFile
		} else {
			read1000 = 1000 - readOffset1000
		}
	}

	if(begin < 11200) {
		if(begin > 1200) {
			readOffset10000 = begin - 1200
		}
		if(begin + amount < 11200) {
			read10000 = amount-read200-read1000
			return readOffset200, read200, readOffset1000, read1000, readOffset10000, read10000, readOffsetFile, readFile
		} else {
			read10000 = 10000 - readOffset10000
		}
	}

	if(begin > 11200) {
		readOffsetFile = offset //seek from the end to amount
		readFile = amount
	} else {
		readOffsetFile = recordData.records - 11200
		readFile = amount - read200 - read1000  - read10000
	}

	return readOffset200, read200, readOffset1000, read1000, readOffset10000, read10000, readOffsetFile, readFile
}

/*
function readFileData
args: User ID, how many records will be read from the storage, read offset, storage
function returns records from the storage
 */
func readFileData(userId uint64, amount uint64, offset uint64, dataFiles *Storage, Coin string) ([]byte) {
	buffer := make([]byte,1)
	recordData := &dataFiles.historyRegistry[userId]

	if (recordData.records < offset || recordData.records == 0 ) {
		return []byte{0}
	}

	ro200, r200, ro1000, r1000, ro10000, r10000, roFile, rFile := readFileDataOffset(userId, amount, offset, dataFiles)

	if (r200 > 0) {
		dataFiles.file200.Seek(recordData.pos200+int64(ro200)*192,0)
		data := readNextBytes(dataFiles.file200, 192*int(r200))

		if (len(buffer) == 1) {
			buffer = data
		} else {
			buffer = append(buffer, data...)
		}
	}

	if (r1000 > 0) {
		dataFiles.file1000.Seek(recordData.pos1000+int64(ro1000)*192,0)
		data := readNextBytes(dataFiles.file1000, 192*int(r1000))
		if (len(buffer) == 1) {
			buffer = data
		} else {
			buffer = append(buffer, data...)
		}
	}

	if (r10000 > 0) {
		dataFiles.file10000.Seek(recordData.pos10000+int64(ro10000)*192,0)
		data := readNextBytes(dataFiles.file10000, 192*int(r10000))
		if (len(buffer) == 1) {
			buffer = data
		} else {
			buffer = append(buffer, data...)
		}
	}

	if (rFile > 0) {
		Path := fmt.Sprintf("../%v/%vU%d.db", Coin, dataFiles.name, userId)
		fileX, err := os.Open(Path)

		if err != nil {
			log.Fatal(err)
		}

		defer fileX.Close()

		fi, err := fileX.Stat()
		if err != nil {
			log.Fatal(err)
		}

		_, err = fileX.Seek(-int64(roFile)*192,2)
		if err != nil {
			// Could not obtain stat, handle error
			log.Fatal(err, fi.Size(), int64(roFile)*192)
		}

		data := readNextBytes(fileX, 192*int(rFile))
		if (len(buffer) == 1) {
			buffer = data
		} else {
			buffer = append(buffer, data...)
		}
		fileX.Close()
	}
	return buffer
}

/*
function sendToMailServer
sending a request to create a new mail
 */
func sendToMailServer() {
	for {
		TMPMailData.mux.Lock()
		if len(TMPMailData.data) >= 158 {
			sendWalletsMail.SendBytes(TMPMailData.data, 0)
			TMPMailData.data = []byte{}
		}
		TMPMailData.mux.Unlock()
		time.Sleep(1*time.Minute)
	}
}

/*
function sendMail
creating a request with a binary data
 */
func sendMail(MailData mailStruct) {
	mailbuf := make([]byte, 158)

	binary.LittleEndian.PutUint16(mailbuf[0:2], uint16(MailData.InOut))
	binary.LittleEndian.PutUint64(mailbuf[2:10], MailData.value)
	binary.LittleEndian.PutUint64(mailbuf[10:18], MailData.CoinID)

	copy(mailbuf[18:74], []byte(MailData.mail))
	copy(mailbuf[74:116], []byte(MailData.sndFrom))
	copy(mailbuf[116:158], []byte(MailData.sndTo))

	TMPMailData.mux.Lock()
	if len(TMPMailData.data) <= 10 {
		TMPMailData.data = mailbuf
	} else {
		TMPMailData.data = append(TMPMailData.data, mailbuf...)
	}
	TMPMailData.mux.Unlock()
}

/*
function reciveTrHistory
function for listening sockets
listening data from "IN" servers
 */
func reciveTrHistory() {
	poller := zmq.NewPoller()
	for i, _ := range rcvTransactionHistory {
		poller.Add(rcvTransactionHistory[i], zmq.POLLIN)
	}

	for {
		sockets, err := poller.Poll(-1)
		if err != nil {
			continue //  Interrupted
		}

		for _, socket := range sockets {
			switch socket.Socket {
			case rcvTransactionHistory[1]:
				msg, _ := rcvTransactionHistory[1].RecvBytes(0)
				if len(msg) >= 198 {
					ProcessTrSokets(msg)
				}
			case rcvTransactionHistory[2]:
				msg, _ := rcvTransactionHistory[2].RecvBytes(0)
				if len(msg) >= 198 {
					ProcessTrSokets(msg)
				}
			case rcvTransactionHistory[3]:
				msg, _ := rcvTransactionHistory[3].RecvBytes(0)
				if len(msg) >= 198 {
					ProcessTrSokets(msg)
				}
			case rcvTransactionHistory[4]:
				msg, _ := rcvTransactionHistory[4].RecvBytes(0)
				if len(msg) >= 198 {
					ProcessTrSokets(msg)
				}
			case rcvTransactionHistory[5]:
				msg, _ := rcvTransactionHistory[5].RecvBytes(0)
				if len(msg) >= 198 {
					ProcessTrSokets(msg)
				}
			case rcvTransactionHistory[6]:
				msg, _ := rcvTransactionHistory[6].RecvBytes(0)
				if len(msg) >= 198 {
					ProcessTrSokets(msg)
				}
			case rcvTransactionHistory[7]:
				msg, _ := rcvTransactionHistory[7].RecvBytes(0)
				if len(msg) >= 198 {
					ProcessTrSokets(msg)
				}
			case rcvTransactionHistory[8]:
				msg, _ := rcvTransactionHistory[8].RecvBytes(0)
				if len(msg) >= 198 {
					ProcessTrSokets(msg)
				}
			case rcvTransactionHistory[9]:
				msg, _ := rcvTransactionHistory[9].RecvBytes(0)
				if len(msg) >= 198 {
					ProcessTrSokets(msg)
				}
			}
		}
	}
}

var globalCounter int

/*
function ProcessTrSokets
function for reading data from sockets and writing them to a binary array
 */
func ProcessTrSokets(msg []byte) {
	var UserId uint64
	var InOut uint64
	var ConfirmedBlocks uint64
	var TrTime uint64
	var CoinID uint64

	var TransactionHash string
	var TrFrom string
	var TrTo string
	var TrValue string

	var TransactionHashLen uint64
	var TrFromLen uint64
	var TrToLen uint64
	var TrValueLen uint64
	var CoinIDLen uint64
	var MailValue uint64

	var typeFl int

	if len(msg)%198 == 0 {
		TransactionHashLen = 98
		TrFromLen = 140
		TrToLen = 182
		TrValueLen = 190
		CoinIDLen = 198
		typeFl = 198
	} else {
		TransactionHashLen = 96
		TrFromLen = 138
		TrToLen = 180
		TrValueLen = 200
		CoinIDLen = 208
		typeFl = 208
	}

	var offset uint64
	offset = 0
	for ij:=0; ij < (len(msg)/typeFl); ij++ {

		UserId = binary.LittleEndian.Uint64(msg[0+offset:8+offset])
		InOut = binary.LittleEndian.Uint64(msg[8+offset:16+offset])
		ConfirmedBlocks = binary.LittleEndian.Uint64(msg[16+offset:24+offset])
		TrTime = binary.LittleEndian.Uint64(msg[24+offset:32+offset])

		TransactionHash = string(bytes.Trim(msg[32+offset:TransactionHashLen+offset], "\x00"))
		TrFrom = string(bytes.Trim(msg[TransactionHashLen+offset:TrFromLen+offset], "\x00"))
		TrTo = string(bytes.Trim(msg[TrFromLen+offset:TrToLen+offset], "\x00"))

		if len(msg)%198 == 0 {
			TrValue2 := binary.LittleEndian.Uint64(msg[TrToLen+offset:TrValueLen+offset])
			TrValue = fmt.Sprintf("%v", TrValue2)
			MailValue = TrValue2
		} else {
			TrValue = string(bytes.Trim(msg[TrToLen+offset:TrValueLen+offset], "\x00"))
			MailValue, _ = strconv.ParseUint(TrValue, 10, 64)
		}

		CoinID = binary.LittleEndian.Uint64(msg[TrValueLen+offset:CoinIDLen+offset])

		offset = offset + uint64(typeFl)

		//Send info about receiving coins
		if InOut == 1 && UserId != 1 && ConfirmedBlocks >= 6 {
			var MailData mailStruct
			MailData.sndTo = TrTo
			MailData.sndFrom = TrFrom
			MailData.CoinID = CoinID
			MailData.value = MailValue
			MailData.InOut = InOut
			MailData.mail = WclientsId.data[UserId].email
			go sendMail(MailData)
		}

		if UserId == 1 { // COLD WALLET TRANSACTION!!!
			//ColdWalletTransactions.db
			if InOut == 1 {
				ColdWalletTrunsactionsIN[TransactionHash] = UserTrunsaction{UserId: UserId, TransactionHash: TransactionHash, InOut: InOut, TrTime: TrTime, ConfirmedBlocks: ConfirmedBlocks, TrTo: TrTo, TrValue: TrValue, TrFrom: TrFrom, CoinID: CoinID}
			} else {
				ColdWalletTrunsactionsOUT[TransactionHash] = UserTrunsaction{UserId: UserId, TransactionHash: TransactionHash, InOut: InOut, TrTime: TrTime, ConfirmedBlocks: ConfirmedBlocks, TrTo: TrTo, TrValue: TrValue, TrFrom: TrFrom, CoinID: CoinID}
			}
		} else {
			if len(TransactionHash) > 10 && len(TrValue) > 2 {
				var updateFlag bool
				updateFlag = false

				//searching for the transaction that already exist in a binary array
				j := 0
				globalCounter++
				AllHistoryBin.mux.Lock()
				for n := 0; n < (len(AllHistoryBin.data) / 192); n++ {
					CoinIDCMP := binary.LittleEndian.Uint16(AllHistoryBin.data[10+j:12+j])
					if CoinIDCMP != uint16(CoinID) {
						j = j + 192
						continue
					}

					uIDCMP := binary.LittleEndian.Uint64(AllHistoryBin.data[0+j:8+j])
					InOutCMP := binary.LittleEndian.Uint16(AllHistoryBin.data[8+j:10+j])

					TrToCMP := string(bytes.Trim(AllHistoryBin.data[24+j:66+j], "\x00"))
					TrFromCMP := string(bytes.Trim(AllHistoryBin.data[66+j:108+j], "\x00"))
					TransactionHashCMP := string(bytes.Trim(AllHistoryBin.data[126+j:192+j], "\x00"))

					if uIDCMP == UserId && InOutCMP == uint16(InOut) && TrToCMP == TrTo && TrFromCMP == TrFrom && TransactionHashCMP == TransactionHash {
						//update comfirmed number
						//DON'T USE PutUint in AllHistoryBin!!!
						var tmpBuf= make([]byte, 2)
						binary.LittleEndian.PutUint16(tmpBuf[0:2], uint16(ConfirmedBlocks))
						copy(AllHistoryBin.data[14+j:16+j], tmpBuf)
						updateFlag = true
						break
					}
					if CoinID == 1 && InOut == 0 || CoinID == 3 && InOut == 0 { // sent Bitcoin
						if uIDCMP == UserId && InOutCMP == uint16(InOut) && strings.ToLower(TrToCMP) == strings.ToLower(TrTo) && TransactionHashCMP == TransactionHash {
							//update comfirmed number
							//DON'T USE PutUint in AllHistoryBin!!!
							var tmpBuf= make([]byte, 2)
							binary.LittleEndian.PutUint16(tmpBuf[0:2], uint16(ConfirmedBlocks))
							copy(AllHistoryBin.data[14+j:16+j], tmpBuf)
							updateFlag = true
							break
						}
					}
					j = j + 192
				}
				AllHistoryBin.mux.Unlock()

				//adding new transaction
				if updateFlag == false {
					var tmpBuf= make([]byte, 192)
					Uid := UserId
					binary.LittleEndian.PutUint64(tmpBuf[0:8], Uid)
					binary.LittleEndian.PutUint16(tmpBuf[8:10], uint16(InOut))
					binary.LittleEndian.PutUint16(tmpBuf[10:12], uint16(CoinID))
					binary.LittleEndian.PutUint16(tmpBuf[12:14], 0)
					binary.LittleEndian.PutUint16(tmpBuf[14:16], uint16(ConfirmedBlocks))
					binary.LittleEndian.PutUint64(tmpBuf[16:24], TrTime)

					copy(tmpBuf[24:66], []byte(TrTo))
					copy(tmpBuf[66:108], []byte(TrFrom))
					copy(tmpBuf[108:126], []byte(TrValue))
					copy(tmpBuf[126:192], []byte(TransactionHash))

					if len(AllHistoryBin.data) <= 1 {
						AllHistoryBin.mux.Lock()
						AllHistoryBin.data = tmpBuf
						AllHistoryBin.mux.Unlock()
					} else {
						AllHistoryBin.mux.Lock()
						AllHistoryBin.data = append(AllHistoryBin.data, tmpBuf...)
						AllHistoryBin.mux.Unlock()
					}
				}
				//globalCounter++
			}
		}
	}
}

/*
function sendBforCheck
sending a request to check balance, to make a new transaction
 */
func sendBforCheck(CoinID uint64, userId uint64, value uint64) {
	buf := make([]byte, 24)
	binary.LittleEndian.PutUint64(buf[0:8], userId)
	binary.LittleEndian.PutUint64(buf[8:16], CoinID)

	binary.LittleEndian.PutUint64(buf[16:24], value)

	if len(buf) == 24 {
		sndCheckBalance.SendBytes(buf, 0)
	}
}

/*
function reciveTransactionAnswer
function for listening sockets
listening data from "OUT" servers
 */
func reciveTransactionAnswer() {
	poller := zmq.NewPoller()
	for i, _ := range rcvTransactionAnswer {
		poller.Add(rcvTransactionAnswer[i], zmq.POLLIN)
	}
	for {
		sockets, err := poller.Poll(-1)
		if err != nil {
			continue //  Interrupted
		}
		for _, socket := range sockets {
			switch socket.Socket {
				case rcvTransactionAnswer[1]:
					result,_ := rcvTransactionAnswer[1].RecvBytes(0) // eat the incoming message
					if len(result) < 3{
						continue
					}
					proccessTransactionAnswer(result)
				case rcvTransactionAnswer[2]:
					result,_ := rcvTransactionAnswer[2].RecvBytes(0) // eat the incoming message
					if len(result) < 3{
						continue
					}
					proccessTransactionAnswer(result)
				case rcvTransactionAnswer[3]:
					result,_ := rcvTransactionAnswer[3].RecvBytes(0) // eat the incoming message
					if len(result) < 3{
						continue
					}
					proccessTransactionAnswer(result)
				case rcvTransactionAnswer[4]:
					result,_ := rcvTransactionAnswer[4].RecvBytes(0) // eat the incoming message
					if len(result) < 3{
						continue
					}
					proccessTransactionAnswer(result)
				case rcvTransactionAnswer[5]:
					result,_ := rcvTransactionAnswer[5].RecvBytes(0) // eat the incoming message
					if len(result) < 3{
						continue
					}
					proccessTransactionAnswer(result)
				case rcvTransactionAnswer[6]:
					result,_ := rcvTransactionAnswer[6].RecvBytes(0) // eat the incoming message
					if len(result) < 3{
						continue
					}
					proccessTransactionAnswer(result)
				case rcvTransactionAnswer[7]:
					result,_ := rcvTransactionAnswer[7].RecvBytes(0) // eat the incoming message
					if len(result) < 3{
						continue
					}
					proccessTransactionAnswer(result)
				case rcvTransactionAnswer[8]:
					result,_ := rcvTransactionAnswer[8].RecvBytes(0) // eat the incoming message
					if len(result) < 3{
						continue
					}
					proccessTransactionAnswer(result)
				case rcvTransactionAnswer[9]:
					result,_ := rcvTransactionAnswer[9].RecvBytes(0) // eat the incoming message
					if len(result) < 3{
						continue
					}
					proccessTransactionAnswer(result)
			}
		}
	}
}

/*
function proccessransactionAnswer
function for reading data from sockets and writing them to response channel
 */
func proccessTransactionAnswer(msg []byte) {
	if len(msg) == 3 {
		sTmp := make([]byte, 8)
		binary.LittleEndian.PutUint64(sTmp[0:8], 3)
		usersChannels.sendHTTPresponse <-sTmp
	}
	if len(msg) == 66 {
		sTmp := make([]byte, 74)
		binary.LittleEndian.PutUint64(sTmp[0:8], 1)
		copy(sTmp[8:74], []byte(msg))
		usersChannels.sendHTTPresponse <-sTmp
	}
	if len(msg) == 64 {
		sTmp := make([]byte, 74)
		binary.LittleEndian.PutUint64(sTmp[0:8], 1)
		copy(sTmp[8:74], []byte(msg))
		usersChannels.sendHTTPresponse <-sTmp
	}
}

/*
function reciveBalanceAnswer
function for listening sockets
listening data from a Balance serve
if responded == 1 new transaction will be created
 */
func reciveBalanceAnswer() {
	poller := zmq.NewPoller()
	poller.Add(rcvCheckBalance, zmq.POLLIN)
	for {
		_, err := poller.Poll(-1)
		if err != nil {
			fmt.Println(err)
			continue //  Interrupted
		}
		msg, _ := rcvCheckBalance.RecvBytes(0)
		if len(msg) != 32 {
			continue
		}
		userId := binary.LittleEndian.Uint64(msg[0:8])
		CoinID := binary.LittleEndian.Uint64(msg[8:16])
		response := binary.LittleEndian.Uint64(msg[16:24])
		value := binary.LittleEndian.Uint64(msg[24:32])

		if response == 0 {
			sTmp := make([]byte, 8)
			binary.LittleEndian.PutUint64(sTmp[0:8], response)
			usersChannels.sendHTTPresponse <-sTmp
			continue
		}

		if response == 1 {
			var sndFromLen int
			var sndToLen int
			var userKey string
			var sndFrom string
			var sndTo string

			addrSendMap.mux.Lock()
			for key, addrData := range addrSendMap.data[userId] {
				if addrData.CoinID == CoinID {
					sndFrom = addrData.sndFrom
					sndTo = addrData.sndTo
					delete(addrSendMap.data[userId], key)
					break
				}
			}
			addrSendMap.mux.Unlock()

			UsersHasAddr.mux.Lock()
			for PrivateKey, data := range UsersHasAddr.data[userId] {
				if  strings.ToLower(data.PublicKey) ==  strings.ToLower(sndFrom) {
					userKey = string(PrivateKey)
					break
				}
			}
			UsersHasAddr.mux.Unlock()

			if CoinID == 1  || CoinID == 3 {
				sndFromLen = 16 + len(userKey)
				sndToLen = sndFromLen + len(sndTo)
				bufLen := sndToLen + len(sndFrom)
				var buf = make([]byte, bufLen)
				binary.LittleEndian.PutUint64(buf[0:8], userId)
				binary.LittleEndian.PutUint64(buf[8:16], value)
				copy(buf[16:sndFromLen], []byte(userKey))
				copy(buf[sndFromLen:sndToLen], []byte(sndTo))
				copy(buf[sndToLen:bufLen], []byte(sndFrom))

				sndNewTransaction[int(CoinID)].SendBytes(buf,0)
			} else {
				sndFromLen = 16 + len(userKey)
				sndToLen = sndFromLen + len(sndTo)
				bufLen := sndToLen
				var buf = make([]byte, bufLen)
				binary.LittleEndian.PutUint64(buf[0:8], userId)
				binary.LittleEndian.PutUint64(buf[8:16], value)
				copy(buf[16:sndFromLen], []byte(userKey))
				copy(buf[sndFromLen:sndToLen], []byte(sndTo))

				sndNewTransaction[int(CoinID)].SendBytes(buf,0)
			}
		}
	}
}

func SendMoneyTo(sndFrom string, sndTo string, CoinID uint64, userId uint64, value uint64) {
	if len(sndTo) != 0 &&  len(sndFrom) != 0{
		sendBforCheck(CoinID, userId, value)
	}
}

/*
function handleWconn
handling http request
 */
func handleWconn(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Connection", "close")

	cmd := r.FormValue("data")
	data := strings.Split(cmd, ",")

	/*
		request struct
		data[0] - type of request //get balance, create wallet, send money
		data[1] - user session
		data[2] - user id
		data[3] - send to wallet // type int, from 1 till 7 (9)
		data[4-9] - request params
	*/

	if (len(data) < 6) {
		w.Write([]byte("nFAIL182"))
		return //wrong request
	}

	userId, _ := strconv.ParseUint(data[2], 10, 0)


	if (Wclients.data[data[1]].userId != userId){
		w.Write([]byte("FAIL197"))
		return //wrong request
	}


	fmt.Println("globalCounter------ ", globalCounter)

	//Check the type of requests
	if (data[0] == "Subscribe") {
		email := strings.TrimSpace(r.FormValue("email"))
		result := subscribe(email)

		w.Write([]byte(fmt.Sprintf("%v", result)))
		return
	}
	if (data[0] == "createWallet") {
		coinType := r.FormValue("coinType")
		CoinID, _ := strconv.ParseUint(r.FormValue("CoinID"), 10, 64)
		walletAddress := getWalletAddress(coinType, userId, CoinID)
		w.Write([]byte(walletAddress))
		return
	}
	if (data[0] == "DeleteWallet") {
		CoinID, _ := strconv.ParseUint(r.FormValue("CoinID"), 10, 64)
		walletAddress := r.FormValue("walletAddress")
		walletAddress = strings.TrimSpace(walletAddress)
		UsersHasAddr.mux.Lock()
		for key, data := range UsersHasAddr.data[userId] {
			if data.CoinID == CoinID && strings.ToLower(data.PublicKey) == strings.ToLower(walletAddress) {
				/*
				Delete From map, but not from file
				When server will be restarted, will load last address
				TODO think about "IN" servers
				*/
				delete(UsersHasAddr.data[userId], key)
				UsersHasAddr.mux.Unlock()
				w.Write([]byte("1"))
				return
			}
		}
		UsersHasAddr.mux.Unlock()
		w.Write([]byte("Error"))
		return
	}
	if (data[0] == "SendCoinsTo" || data[0] == "QuickSendCoins") {
		var sndFrom string
		sndTo := r.FormValue("sndTo")
		sndTo = strings.TrimSpace(sndTo)
		CoinID, _ := strconv.ParseUint(r.FormValue("CoinID"), 10, 64)
		value, _ := strconv.ParseUint(r.FormValue("value"), 10, 64)

		if value <= 0 {
			msg := fmt.Sprintf("%v,%v","Error: Set the value up!", 0)
			w.Write([]byte(msg))
			return
		}

		if  data[0] == "QuickSendCoins" {
			UsersHasAddr.mux.Lock()
			Wallets, is_ok := UsersHasAddr.data[userId]
			UsersHasAddr.mux.Unlock()

			if is_ok {
				for _, WalletData := range Wallets {
					if WalletData.CoinID == CoinID {
						sndFrom = WalletData.PublicKey
						break
					}
				}
				if len(sndFrom) < 10 {
					msg := fmt.Sprintf("%v,%v","ERROR: Not Found Sender Address", 0)
					w.Write([]byte(msg))
					return
				}
			} else {
				msg := fmt.Sprintf("%v,%v","ERROR: Not Found User", 0)
				w.Write([]byte(msg))
				return
			}
		} else {
			sndFrom = r.FormValue("sndFrom")
			sndFrom = strings.TrimSpace(sndFrom)
		}

		var tmpData  = make(map[uint64]struct {
			sndFrom string
			sndTo string
			CoinID uint64
		})

		var tmpStruct struct{
			sndFrom string
			sndTo string
			CoinID uint64
		}

		tmpStruct.sndTo = sndTo
		tmpStruct.sndFrom = sndFrom
		tmpStruct.CoinID = CoinID

		addrSendMap.mux.Lock()
		if len(addrSendMap.data[userId]) > 0 {
			for key, prevData := range addrSendMap.data[userId] {
				tmpData[key] = prevData
			}
			pointer := len(tmpData) + 1
			tmpData[uint64(pointer)] = tmpStruct
		} else {
			tmpData[0] = tmpStruct
		}
		addrSendMap.data[userId] = tmpData
		addrSendMap.mux.Unlock()

		SendMoneyTo(sndFrom, sndTo, CoinID, userId, value)

		var msg string
		var response uint64
		for{
			data, _  := <- usersChannels.sendHTTPresponse
			response = binary.LittleEndian.Uint64(data[0:8])
			if len(data) == 74 {
				msg = string(bytes.Trim(data[8:74], "\x00"))
			}
			break
		}
		if response == 0 {
			msg := fmt.Sprintf("%v,%v","ERROR: Not Enough Money", 0)
			w.Write([]byte(msg))
			return
		}
		if response == 1 {
			var MailData mailStruct

			MailData.sndTo = sndTo
			MailData.sndFrom = sndFrom
			MailData.CoinID = CoinID
			MailData.value = value
			MailData.InOut = 2
			MailData.mail = WclientsId.data[userId].email
			go sendMail(MailData)
			answer := fmt.Sprintf("%v,%v",msg, 1)

			var tmpBuf= make([]byte, 192)
			Uid := userId
			binary.LittleEndian.PutUint64(tmpBuf[0:8], Uid)
			binary.LittleEndian.PutUint16(tmpBuf[8:10], uint16(0))
			binary.LittleEndian.PutUint16(tmpBuf[10:12], uint16(CoinID))
			binary.LittleEndian.PutUint16(tmpBuf[12:14], 0)
			binary.LittleEndian.PutUint16(tmpBuf[14:16], uint16(0))
			binary.LittleEndian.PutUint64(tmpBuf[16:24], uint64(time.Now().Unix()))

			Trvalue := fmt.Sprintf("%v", value)
			copy(tmpBuf[24:66], []byte(strings.ToLower(sndTo)))
			copy(tmpBuf[66:108], []byte(strings.ToLower(sndFrom)))
			copy(tmpBuf[108:126], []byte(Trvalue))
			copy(tmpBuf[126:192], []byte(msg))

			if len(AllHistoryBin.data) <= 1 {
				AllHistoryBin.mux.Lock()
				AllHistoryBin.data = tmpBuf
				AllHistoryBin.mux.Unlock()
			} else {
				AllHistoryBin.mux.Lock()
				AllHistoryBin.data = append(AllHistoryBin.data, tmpBuf...)
				AllHistoryBin.mux.Unlock()
			}
			w.Write([]byte(answer))
			return
		}
		if response == 3 {
			msg := fmt.Sprintf("%v,%v","ERROR: Sending Error", 0)
			w.Write([]byte(msg))
			return
		}
		return
	}
	if (data[0] == "DisplayAllWallets") {
		response := displayAllUserWallets(userId)
		w.Write([]byte(response))
		return
	}
	if (data[0] == "CoutTransactions") {
		response := coutTransactions(userId)
		w.Write([]byte(response))
		return
	}
	if (data[0] == "DisplayAllWalletsTR") {
		response := displayAllUserWallets(userId)
		w.Write([]byte(response))
		return
	}
	if (data[0] == "DisplayAllHistory") {
		filter, _ := strconv.ParseUint(data[3], 10, 0)
		response := displayAllHistory(userId, filter, "")
		w.Write([]byte(response))
		return
	}
	if (data[0] == "DisplayWalletHistory") {
		address := data[3]
		filter, _ := strconv.ParseUint(data[4], 10, 0)
		response := displayAllHistory(userId, filter, strings.ToLower(address))
		w.Write([]byte(response))
		return
	}
}

func coutTransactions(userId uint64) string{
	_, userIsOk := WclientsId.data[userId]
	var transactions int
	if userIsOk {
		transactions = 0
		AllHistoryBin.mux.Lock()
		AllHistoryBinCMP := AllHistoryBin.data
		AllHistoryBin.mux.Unlock()
		j := 0
		for n:=0; n < (len(AllHistoryBinCMP)/192);n++ {
			uID := binary.LittleEndian.Uint64(AllHistoryBinCMP[0+j : 8+j])
			if userId == uID {
				transactions++
			}
			j = j + 192
		}

		result := fmt.Sprintf("%v", transactions)
		return result
	}
	return ""
}

/*
function displayAllHistory
reading data from a binary array and send it to user
 */
func displayAllHistory(userId uint64, filter uint64, address string) string{
	_, userIsOk := WclientsId.data[userId]
	if userIsOk {
		if filter == 0 {
			filter = 1;
		}
		var allHistory string

		j := 0

		AllHistoryBin.mux.Lock()
		AllHistoryBinCMP := AllHistoryBin.data
		AllHistoryBin.mux.Unlock()

		//var ij int
		//var records int
		//ij = 0
		//
		//records = len(AllHistoryBinCMP) / 192
		//
		//if records > 100 {
		//	ij = records - 100
		//	j = ij*192
		//}

		var records int
		records = 0

		for n:=0; n < ((len(AllHistoryBinCMP)/192));n++ {

			if records == 100 {
				continue
			}

			uID := binary.LittleEndian.Uint64(AllHistoryBinCMP[0+j:8+j])
			InOut := binary.LittleEndian.Uint16(AllHistoryBinCMP[8+j:10+j])
			CoinID := binary.LittleEndian.Uint16(AllHistoryBinCMP[10+j:12+j])
			ConfirmedBlocks := binary.LittleEndian.Uint16(AllHistoryBinCMP[14+j:16+j])
			TrTime := binary.LittleEndian.Uint64(AllHistoryBinCMP[16+j:24+j])

			TrTo := string(bytes.Trim(AllHistoryBinCMP[24+j:66+j], "\x00"))
			TrFrom := string(bytes.Trim(AllHistoryBinCMP[66+j:108+j], "\x00"))
			TrValue := string(bytes.Trim(AllHistoryBinCMP[108+j:126+j], "\x00"))
			TransactionHash := string(bytes.Trim(AllHistoryBinCMP[126+j:192+j], "\x00"))

			if address == "" {
				if filter == 1 { //All History
					if uID == userId {
						if filter == 1 {
							records++
							allHistory = fmt.Sprintf("%v,%v,%v,%v,%v,%v,%v,%v %v", TrValue, TrFrom, TrTo, TransactionHash, ConfirmedBlocks, InOut, TrTime, CoinID, allHistory)
						}
					} else {
						j = j+192
						continue
					}
				}
				if filter == 2 { //Recived
					if uID == userId && InOut == 1 {
						records++
						allHistory = fmt.Sprintf("%v,%v,%v,%v,%v,%v,%v,%v %v", TrValue, TrFrom, TrTo, TransactionHash, ConfirmedBlocks, InOut, TrTime, CoinID, allHistory)
					} else {
						j = j+192
						continue
					}
				}
				if filter == 3 { //Sent
					if uID == userId && InOut == 0 {
						records++
						allHistory = fmt.Sprintf("%v,%v,%v,%v,%v,%v,%v,%v %v", TrValue, TrFrom, TrTo, TransactionHash, ConfirmedBlocks, InOut, TrTime, CoinID, allHistory)
					} else {
						j = j+192
						continue
					}
				}
			} else {
				if filter == 1 { //All History
					if uID == userId {
						if filter == 1 && address == strings.ToLower(TrTo) {
							records++
							allHistory = fmt.Sprintf("%v,%v,%v,%v,%v,%v,%v,%v %v", TrValue, TrFrom, TrTo, TransactionHash, ConfirmedBlocks, InOut, TrTime, CoinID, allHistory)
						}
						if filter == 1 && address == strings.ToLower(TrFrom) {
							records++
							allHistory = fmt.Sprintf("%v,%v,%v,%v,%v,%v,%v,%v %v", TrValue, TrFrom, TrTo, TransactionHash, ConfirmedBlocks, InOut, TrTime, CoinID, allHistory)
						}
					} else {
						j = j+192
						continue
					}
				}
				if filter == 2 && address == strings.ToLower(TrTo) { //Recived
					if uID == userId && InOut == 1 {
						records++
						allHistory = fmt.Sprintf("%v,%v,%v,%v,%v,%v,%v,%v %v", TrValue, TrFrom, TrTo, TransactionHash, ConfirmedBlocks, InOut, TrTime, CoinID, allHistory)
					} else {
						j = j+192
						continue
					}
				}
				if filter == 3 && address == strings.ToLower(TrFrom) { //Sent
					if uID == userId && InOut == 0 {
						records++
						allHistory = fmt.Sprintf("%v,%v,%v,%v,%v,%v,%v,%v %v", TrValue, TrFrom, TrTo, TransactionHash, ConfirmedBlocks, InOut, TrTime, CoinID, allHistory)
					} else {
						j = j+192
						continue
					}
				}
			}

			j = j+192
		}

		return allHistory
	}
	return ""
}

/*
function displayAllUserWallets
displaying all user wallets
 */
func displayAllUserWallets(userId uint64) string {
	_, userIsOk := WclientsId.data[userId]

	var allWallets string

	if userIsOk {
		UsersHasAddr.mux.Lock()
		Wallets, is_ok := UsersHasAddr.data[userId]
		UsersHasAddr.mux.Unlock()
		if is_ok {
			for _, WalletData := range Wallets {
				allWallets = fmt.Sprintf("%v,%v %v", WalletData.PublicKey, WalletData.CoinID, allWallets)
			}
			return allWallets
		}
		return "No Wallets"
	}
	return "FAIL197"
}

/*
function updeteListningServers
send info about new user wallet address
 */
func updeteListningServers(userId uint64, publicKey string, privateKey string, CoinID uint64) {
	if len(privateKey) != 0 &&  len(publicKey) != 0{
		var publicKeyLen int
		var privateKeyLen int
		publicKeyLen = 8 + len(publicKey)
		privateKeyLen = publicKeyLen + len(privateKey)
		bufLen := privateKeyLen
		var buf = make([]byte, bufLen)

		binary.LittleEndian.PutUint64(buf[0:8], userId)
		copy(buf[8:publicKeyLen], []byte(publicKey))
		copy(buf[publicKeyLen:privateKeyLen], []byte(privateKey))

		sndNewUsrAddr[int(CoinID)].SendBytes(buf,0)
	}
}

/*
function getWalletAddress
give the user a new wallet address
 */
func getWalletAddress(coinType string, userId uint64, CoinID uint64) string {
	if CoinID == 0 {
		return ""
	}
	// 1 user can create 1 address for Wallet
	UsersHasAddr.mux.Lock()
	for _, data := range UsersHasAddr.data[userId] {
		if data.CoinID == CoinID {
			UsersHasAddr.mux.Unlock()
			return ""
		}
	}
	UsersHasAddr.mux.Unlock()

	_, userIsOk := WclientsId.data[userId]
	if userIsOk {
		var userAddress string
		var KeyToDelete string
		var coinTypeID string

		if coinType == "Ethereum" || coinType == "Ether1" || coinType == "EthClassic" {
			coinTypeID = "Ethereum"
			coinType = "Ethereum"
		}
		if coinType == "Litecoin" || coinType == "Ravecoin" || coinType == "Bitcoin" {
			coinTypeID = "Bitcoin"
		}
		FreeCoinAdresses.mux.Lock()
		for PrivateKey, PublicKey := range FreeCoinAdresses.data[coinType] {
			userAddress = PublicKey
			KeyToDelete = PrivateKey
			break
		}
		delete(FreeCoinAdresses.data[coinType], KeyToDelete)
		FreeCoinAdresses.mux.Unlock()

		//update users' wallet address storage
		writeFileWallet(KeyToDelete, userAddress, userId, coinTypeID, CoinID)
		go updeteListningServers(userId , userAddress , KeyToDelete, CoinID)

		deleteAddFromFree(KeyToDelete, userAddress, coinType)
		response := fmt.Sprintf("%v,%v", userAddress, CoinID)
		return response
	}
	return ""
}

/*
function deleteAddFromFree
updating the wallet address storage
 */
func deleteAddFromFree(PrivateKey string, PublicKey string, coinType string) {
	string := fmt.Sprintf("%v%v", PrivateKey, PublicKey)
	stringLen := len(string)
	var file *os.File
	var err error
	var PrivateKeyLen int64
	var PublicKeyLen int64
	if stringLen == 106 {
		PrivateKeyLen = 64
		PublicKeyLen = 	42
		file, err = os.OpenFile("../users/FreeKeysEthereum.db", os.O_APPEND|os.O_WRONLY, 0666)
	}
	if stringLen == 86 {
		PrivateKeyLen = 52
		PublicKeyLen = 34
		file, err = os.OpenFile(fmt.Sprintf("../users/FreeKeys%v.db", coinType), os.O_APPEND|os.O_WRONLY, 0666)
	}
	defer file.Close()
	if err != nil {
		log.Fatal("Open getFreeAddress file error ",err)
	}

	file.Truncate(0)
	file.Seek(0,0)

	FreeCoinAdresses.mux.Lock()
	for PrivateKey, PublicKey := range FreeCoinAdresses.data[coinType] {
		b := make([]byte, PrivateKeyLen)
		copy(b[:], []byte(PrivateKey))
		bufferedWriter := bufio.NewWriter(file)
		_, err = bufferedWriter.Write( b, )
		if err != nil {
			log.Fatal(err)
		}

		b = make([]byte, PublicKeyLen)
		copy(b[:], []byte(PublicKey))
		_, err = bufferedWriter.Write( b, )
		if err != nil {
			log.Fatal(err)
		}
		bufferedWriter.Flush()
	}
	FreeCoinAdresses.mux.Unlock()
	file.Close()

}

/*
function getFreeAddress
load addresses from the storage
 */
func getFreeAddress(currency string, currencyType string) {
	var keysData = make(map[string]string)

	file, err := os.Open(fmt.Sprintf("../users/FreeKeys%v.db",currency))
	defer file.Close()
	if err != nil {
		log.Fatal("Open getFreeAddress file error ",err)
	}
	fi, err := file.Stat()
	if err != nil {
		log.Fatal("Could not obtain stat", err)
	}

	if currencyType == "BTC" {
		for i := uint64(0); i < uint64(fi.Size()/86); i++ {
			KeysData := readNextBytes(file, 86)
			PrivateKey := string(KeysData[0:52])
			PublicKey := string(KeysData[52:86])
			keysData[PrivateKey] = PublicKey
		}
		FreeCoinAdresses.mux.Lock()
		FreeCoinAdresses.data[currency] = keysData
		FreeCoinAdresses.mux.Unlock()
	}
	if currencyType == "ETH" {
		for i := uint64(0); i < uint64(fi.Size()/106); i++ {
			KeysData := readNextBytes(file, 106)
			PrivateKey := string(KeysData[0:64])
			PublicKey := string(KeysData[64:106])
			keysData[PrivateKey] = PublicKey
		}
		FreeCoinAdresses.mux.Lock()
		FreeCoinAdresses.data[currency] = keysData
		FreeCoinAdresses.mux.Unlock()
	}

	file.Close()
}

/*
function writeFileWallet
saving users' new wallet address
*/
func writeFileWallet(PrivateKey string, PublicKey string, userId uint64, coinType string, CoinID uint64) {
	var file *os.File
	var err error
	if coinType == "Bitcoin" {
		file, err = os.OpenFile("../users/usersBTC.db", os.O_APPEND|os.O_WRONLY, 0666)
	}
	if coinType == "Ethereum" {
		file, err = os.OpenFile("../users/usersETH.db", os.O_APPEND|os.O_WRONLY, 0666)
	}
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	var keysData = make(map[string]struct {
		CoinID uint64
		PublicKey string
	})

	var TMPstructData TMPstruct
	TMPstructData.CoinID = CoinID
	TMPstructData.PublicKey = PublicKey

	UsersHasAddr.mux.Lock()
	if len(UsersHasAddr.data[userId]) > 0 {
		for key, prevData := range UsersHasAddr.data[userId] {
			fmt.Println(key, prevData)
			keysData[key] = prevData
		}
		keysData[PrivateKey] = TMPstructData
	} else {
		keysData[PrivateKey] = TMPstructData
	}

	UsersHasAddr.data[userId] = keysData
	UsersHasAddr.mux.Unlock()

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, userId)
	bufferedWriter := bufio.NewWriter(file)
	_, err = bufferedWriter.Write(b,)
	if err != nil {
		log.Fatal(err)
	}

	b = make([]byte, 8)
	binary.LittleEndian.PutUint64(b, CoinID)
	_, err = bufferedWriter.Write(b,)
	if err != nil {
		log.Fatal(err)
	}

	_, err = bufferedWriter.Write( []byte(PrivateKey), )
	if err != nil {
		log.Fatal(err)
	}
	_, err = bufferedWriter.Write( []byte(PublicKey), )
	if err != nil {
		log.Fatal(err)
	}
	bufferedWriter.Flush()

	file.Close()
}

/*
function loadUserAddr
loading the user address from the storage
 */
func loadUserAddr(coinType string) {
	var Path string
	var fileLen int64
	var PrivateKeyLen int64
	var PublicKeyLen int64
	if coinType == "BTC" {
		Path = "../users/usersBTC.db"
		fileLen = 102
		PrivateKeyLen = 68
		PublicKeyLen = 102
	}
	if coinType == "ETH" {
		fileLen = 122
		PrivateKeyLen = 80
		PublicKeyLen = 	122
		Path = "../users/usersETH.db"
	}
	file, err := os.Open(Path)
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}
	fi, err := file.Stat()
	if err != nil {
		//Failed to get stat, handle error
		log.Fatal("Failed to get stat ", err)
	}
	if err != nil {
		log.Fatal(err)
	}
	var i uint64

	for i = 0; i < uint64(fi.Size()/(fileLen)); i++ {
		var offset = int(i) + 1
		if offset == 0 {
			offset = 1
		}
		data := readNextBytes(file, (int(fileLen)))

		var PrivateKey string
		var PublicKey string

		var keysData = make(map[string]struct {
			CoinID uint64
			PublicKey string
		})
		var TMPstructData TMPstruct

		if coinType == "BTC" {
			PrivateKey = string(bytes.Trim(data[16:PrivateKeyLen], "\x00"))
			PublicKey = string(bytes.Trim(data[PrivateKeyLen:PublicKeyLen], "\x00"))
		}
		if coinType == "ETH" {
			PrivateKey = string(bytes.Trim(data[16:PrivateKeyLen], "\x00"))
			PublicKey = string(bytes.Trim(data[PrivateKeyLen:PublicKeyLen], "\x00"))
		}

		UserId := binary.LittleEndian.Uint64(data[0:8])
		CoinID :=  binary.LittleEndian.Uint64(data[8:16])

		if CoinID != 1 && CoinID != 4 && CoinID != 3 {
			continue
		}

		TMPstructData.CoinID = CoinID
		TMPstructData.PublicKey = PublicKey

		UsersHasAddr.mux.Lock()
		if len(UsersHasAddr.data[UserId]) > 0 {
			var keyFalg bool
			keyFalg = true
			for key, prevData := range UsersHasAddr.data[UserId] {
				if prevData.CoinID == CoinID {
					prevData.PublicKey = PublicKey
					keyFalg = false
					key = PrivateKey
				}
				keysData[key] = prevData
			}
			if keyFalg == true {
				keysData[PrivateKey] = TMPstructData
			}
		} else {
			keysData[PrivateKey] = TMPstructData
		}
		UsersHasAddr.data[UserId] = keysData
		UsersHasAddr.mux.Unlock()
	}
}

/*
function checkWalletsAuth
function for listening sockets
listening data from Auth server
receiving info about user authorization or about new user
 */
func checkWalletsAuth() {
	poller := zmq.NewPoller()
	poller.Add(rcvWalletsAuth, zmq.POLLIN)

	var userId uint64
	var userSession string

	for {
		_, err := poller.Poll(-1)
		if err != nil {
			log.Printf("Socket error: %v", err)
			break //  Interrupted
		}

		msg, _ := rcvWalletsAuth.RecvBytes(0)
		s := strings.Split(string(msg), ",")

		if len(s) < 3 {
			//wrong request
			continue
		}
		if (s[2] != "" && s[1] != "" && s[3] != "") {
			userId, _ = strconv.ParseUint(s[2], 10, 64)
			userSession = s[1]
			Wclients.mux.Lock()
			Wclients.data[userSession] = Session{userId, userSession, time.Now().Unix(),  s[3]}
			Wclients.mux.Unlock()

			WclientsId.mux.Lock()
			WclientsId.data[userId] = Session{userId, userSession, time.Now().Unix(), s[3]}
			WclientsId.mux.Unlock()

			if s[0] == "n" { //handle registration
				var coinType string
				for CoinID, data := range WTypeAddr {
					if CoinID == 0 || data.WalletName == "" {
						continue
					}
					if data.WalletName == "Ethereum" || data.WalletName == "Ether1" || data.WalletName == "EthClassic" {
						coinType = "Ethereum"
					}
					if data.WalletName == "Litecoin" || data.WalletName == "Ravecoin" || data.WalletName == "Bitcoin" {
						coinType = "Bitcoin"
					}
					getWalletAddress(coinType, userId, uint64(CoinID))
				}
			}
		} else if(s[0] != "" && s[2] != "") {
			//logout
			userID, _ := strconv.ParseUint(s[2], 10, 64)
			WclientsId.mux.Lock()
			delSession := WclientsId.data[userID].userSession
			delete(WclientsId.data, userID)

			WclientsId.mux.Unlock()
			Wclients.mux.Lock()
			delete(Wclients.data, delSession)
			Wclients.mux.Unlock()

			continue
		} else {
			continue
		}

	}
}

/*
function readNextBytes
function returns N bytes from the file
 */
func readNextBytes(file *os.File, number int) []byte {
	bytes := make([]byte, number)

	_, err := file.Read(bytes)
	if err != nil {
		log.Fatal(err)
	}

	return bytes
}
