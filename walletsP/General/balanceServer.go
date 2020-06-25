package main

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	exchange "exchange2/Library"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	zmq "github.com/pebbe/zmq4"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"wallets/walletsP/LocalConfig" //set ip adress and other settings
	"wallets/walletsP/Pairs"
)

const RunType = "Wallets"
//ZMQ socket Vars
var SendToObook = make(map[int]*zmq.Socket)
var pairData Pairs.Pair
var rcvCheckBalance *zmq.Socket
var sndCheckBalance *zmq.Socket
var sndObook *zmq.Socket
var sndObook0 *zmq.Socket
var rcvAccountAuth *zmq.Socket
var rcvBalance = make(map[int]*zmq.Socket)
var sndBalance *zmq.Socket

var transactionsCount uint64
var CurrencyCourse uint64
var USDinEUR_Price float64
var file10klog *os.File
var globalCounter int
var tcpPairs[1000]Pairs.Pair
var delimiter[1000]uint64
var authSessions [10000]string //TODO make it 100k? make it as a service so you can run thousands of servers with users

//Maps
//var currencies[100][10000]uint64
var currencies struct{
	data [100][10000]uint64
	mux sync.Mutex // mutex for the concurrent map
}

var CoinsPriceUSD struct {
	data map[uint64]float64
	mux sync.Mutex // mutex for the concurrent map
}

var walletDelimetrMap struct {
	data [1000]uint64
	mux sync.Mutex // mutex for the concurrent map
}

var clients struct {
	data map[string]Session
	mux sync.Mutex // mutex for the concurrent map
}

//channels
var WriteTolog = make(chan []byte, 1000000)

//types
type Session struct {
	userId uint64
	userSession string //TODO protect from oversize
	time int64
}

type CourseData struct {
	Time struct{} `json:"time"`
	Disclaimer string `json:"disclaimer"`
	Bpi struct {
		Usd struct {
			Code string `json:"code"`
			Rate_float float64 `json:"rate_float"`
		} `json:"usd"`
	} `json:"bpi"`
}

type USDinEUR_Pricetype struct {
	Rates struct {
		Usd float64`json:"USD"`
	} `json:"rates"`
}

type CourseDataMarket struct {
	Status struct {} `json:"status"`
	Data []struct{
		Name string `json:"name"`
		Symbol string `json:"symbol"`
		Quote struct {
			Usd struct {
				Price float64 `json:"price"`
			} `json:"USD"`
		} `json:"quote"`
	} `json:"data"`
}

type Message struct{
	EventType string  `json:"e"`//"e": "trade",     // Event type
	EventTime uint64  `json:"E"`//"E": 123456789,   // Event time
	Symbol string `json:"s"`//"s": "BNBBTC",    // Symbol
	TradeId uint64 `json:"t"`//"t": 12345,       // Trade ID
	Price string `json:"p"`//"p": "0.001",     // Price
	Amount string `json:"q"`//"q": "100",       // Quantity
	OId uint64 `json:"b"`//"b": 88,          // Buyer order ID
	SId uint64 `json:"a"`//"a": 50,          // Seller order ID
	TradeTime uint64 `json:"T"`//"T": 123456785,   // Trade time
	IsMmaker bool `json:"m"`//"m": true,        // Is the buyer the market maker?
	Ignore bool `json:"M"`//"M": true         // Ignore
}
/*
function readAccounsBalance
read the user balance from the database and write it to the map
 */
func readAccounsBalance(fileName string) {
	file, err := os.Open(fileName)
	defer file.Close()
	if err != nil {
		fmt.Println("File is not found: ", fileName)
		return
	}

	fi, err := file.Stat()
	if err != nil {
		log.Fatal("Failed to get stats: ", err)
	}

	for i := int64(0); i < fi.Size()/18; i++ {
		data := readNextBytes(file, 18)
		if (len(data)<18){
			fmt.Println(len(data))
		}

		currency := binary.LittleEndian.Uint16(data[0:2])
		userId := binary.LittleEndian.Uint64(data[2:10])
		balance := binary.LittleEndian.Uint64(data[10:18])
		currencies.mux.Lock()
		currencies.data[currency][userId] = balance
		currencies.mux.Unlock()
	}
}

/*
function readNextBytes
function returns N bytes from the file
 */
func readNextBytes(file *os.File, number int) []byte {
	bytes := make([]byte, number)

	size, err := file.Read(bytes)
	if err != nil {
		log.Fatal(err, size)
	}

	if (size != number) {
		//size does not fit
		return []byte{}
	}

	return bytes
}
/*
function writeTransactionsLog
function for writing balanse log, for debug tasks
 */
func writeTransactionsLog() {
	var err error
	file10klog, err = os.OpenFile("../tcp_balance/balance10k.log", os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}

	defer file10klog.Close()

	for {
		msg := <- WriteTolog

		_, err = file10klog.Write(msg)
		if err != nil {
			log.Println(err)
		}
	}
}

/*
function storeData
function for write to database.
Writing every 10 transaction and write logs
 */
func storeData() {
	//every 10 transactions
	//dump all currencies
	//dump all users
	//rename OLD file to balance_date_time_10k
	var err error

	for {

		time.Sleep(10*time.Second)
		if (transactionsCount < 10) {
			continue
		}

		os.Rename("../tcp_balance/balance10k.log", fmt.Sprintf("../tcp_balance/balance10k%v.log", time.Now().UnixNano()))
		os.Rename("../tcp_balance/balance10k.db", fmt.Sprintf("../tcp_balance/balance10k%v.db", time.Now().UnixNano()))

		file10klog.Close()
		file10klog, err = os.OpenFile("../tcp_balance/balance10k.log", os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			log.Fatal(err)
		}

		var err error
		var fileOL *os.File
		fileOL, err = os.OpenFile("../tcp_balance/balance10k.db", os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			log.Fatal(err)
		}

		buf := []byte{}
		currencies.mux.Lock()
		transactionsCount = 0
		for i := range currencies.data {
			for u := range currencies.data[i] {
				//if balance >0 write to file uid->balance
				if (currencies.data[i][u] > 0) {
					//convert to msg
					//write to file
					buf = append(buf, exchange.ConvertBalanceMsg(uint16(i), uint64(u), currencies.data[i][u])...)
				}
			}
		}
		currencies.mux.Unlock()

		_, err = fileOL.Write(buf)
		if err != nil {
			log.Println(err)
		}

		fileOL.Close()
	}
}

var addr = flag.String("addr", "stream.binance.com:9443", "http service address")

func get_current_Сourse() {
	flag.Parse()
	log.SetFlags(0)

	u := url.URL{Scheme: "wss", Host: *addr, Path: "/stream"}
	u.RawQuery = "streams=btcusdt@depth/ethusdt@depth"

	c, statusFlag, err := websocket.DefaultDialer.Dial(u.String(), nil)

	if err != nil {
		//---------------
		time.Sleep(time.Second * 5)
		get_current_Сourse()
		//---------------
		return
	}

	if statusFlag == nil || statusFlag.StatusCode != 101 {
		time.Sleep(time.Second * 5)
		get_current_Сourse()
		//---------------
		return
	}
	defer c.Close()

	var btcFlag bool
	var ethFlag bool

	btcFlag = false
	ethFlag = false

	for {
		statuscode, message, err := c.ReadMessage()

		if err != nil  || statuscode !=1 {
			time.Sleep(time.Second * 2)
			get_current_Сourse()
			return
		}

		if message == nil {
			defer c.Close()
			time.Sleep(time.Second * 5)
			get_current_Сourse()
			return
		}
		var data map[string]interface{}
		err = json.Unmarshal([]byte(message), &data)
		if err != nil {
			time.Sleep(time.Second * 5)
			get_current_Сourse()
			return
		}
		if len(data) < 2 {
			time.Sleep(time.Second * 5)
			get_current_Сourse()
			return
		}
		splitData := strings.Split(string(data["stream"].(string)), "@")

		readData :=  data["data"].(map[string]interface{})
		if len(readData) == 0 || readData == nil{
			fmt.Println("len(readData) == 0")
			continue
		}
		dataMap := readData["a"].([]interface{})
		if len(dataMap) == 0 || dataMap == nil{
			fmt.Println("len(dataMap) == 0")
			continue
		}
		valueR := dataMap[0].([]interface{})
	    var result float64
		tmp := fmt.Sprintf("%v", valueR[0])
		result,_ = strconv.ParseFloat(tmp, 64)

		if splitData[0] == "btcusdt" && btcFlag == false{
			CoinsPriceUSD.mux.Lock()
			CoinsPriceUSD.data[1] = result
			CoinsPriceUSD.mux.Unlock()
			btcFlag = true
		}
		if splitData[0] == "ethusdt" && ethFlag == false{
			CoinsPriceUSD.mux.Lock()
			CoinsPriceUSD.data[4] = result
			CoinsPriceUSD.mux.Unlock()
			ethFlag = true
		}

	}
}

/*
function get_current_course
Get coins from coinmarketcap API to USD data
 */
func get_current_course() {
	client := &http.Client{}
	req, err := http.NewRequest("GET", LocalConfig.MarketCupApiURL, nil)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	//set query params
	q := url.Values{}
	q.Add("start", "1")
	q.Add("limit", "1000")
	q.Add("convert", "USD")
	q.Add("cryptocurrency_type", "coins")

	req.Header.Set("Accepts", "application/json")
	req.Header.Add("X-CMC_PRO_API_KEY", LocalConfig.MarketCupApiKey)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req);
	if err != nil {
		fmt.Println("Error sending request to server")
		os.Exit(1)
	}

	respBody, _ := ioutil.ReadAll(resp.Body)
	data := &CourseDataMarket{}
	err = json.Unmarshal(respBody, data)

	for _, PairsData := range Pairs.WalletList {
		for _, RecivedData := range data.Data {
			if strings.ToLower(PairsData.WalletName) == strings.ToLower(RecivedData.Name) {
				CoinsPriceUSD.mux.Lock()
				CoinsPriceUSD.data[PairsData.WalletID] = RecivedData.Quote.Usd.Price
				CoinsPriceUSD.mux.Unlock()
			}
		}
	}
}

/*
function get_course
update course data every 12 minutes
 */
func get_course() {
	for {
		get_current_Сourse()
		time.Sleep(2 * time.Minute)
	}
}

func main() {
	// make maps with no zero data
	CoinsPriceUSD.data = make(map[uint64]float64)
	//walletDelimetrMap.data = make(map[uint64]uint64)
	clients.data = make(map[string]Session)
	//------
	delimiter = Pairs.Delimiter
	tcpPairs = Pairs.PairList
	walletDelimetrMap.data = delimiter

	maxOpenFiles()
	go get_course()
	go getUSDinEUR_Price()

	readAccounsBalance("../tcp_balance/balance10k.db")
	readAccounsBalance("../tcp_balance/balance10k.log")

	go writeTransactionsLog()
	go storeData()

	if RunType == "exchange" {
		go exchange.SendDataToObook(SendToObook)

		var sndObookAddress = make(map[int]string)
		for i := 0; i <= 10; i++  {
			SendToObook[i],_ = zmq.NewSocket(zmq.PUSH)
			defer SendToObook[i].Close()
			SendToObook[i].SetRcvhwm(1100000)
			sndObookAddress[i] = fmt.Sprintf("%v%v", LocalConfig.LocalTcpIpAddr, tcpPairs[i].ObookP)
			SendToObook[i].Bind(sndObookAddress[i])
			time.Sleep(10*time.Millisecond)
		}
	}

	rcvAccountAuth, _ = zmq.NewSocket(zmq.PULL)
	defer rcvAccountAuth.Close()
	rcvAccountAuth.SetRcvhwm(1100000)
	rcvAccounAddress := fmt.Sprintf("%v%v", LocalConfig.LocalTcpIpAddr, LocalConfig.BalancercvAccountAuthPort)
	rcvAccountAuth.Connect(rcvAccounAddress)

	var rcvBalanceAddresses = make(map[int]string)

	var ij int
	for i := LocalConfig.BalancercvBalanceBegin; i <= LocalConfig.BalancercvBalanceend; i++  {
		if i == 5 || i == 9 {
			ij++
			continue
		}
		rcvBalance[ij],_ = zmq.NewSocket(zmq.PULL)
		defer rcvBalance[i].Close()
		rcvBalance[ij].SetRcvhwm(1100000)
		rcvBalanceAddresses[ij] = fmt.Sprintf("%v%v", LocalConfig.LocalTcpIpAddr, i)
		rcvBalance[ij].Connect(rcvBalanceAddresses[ij])
		ij++
		time.Sleep(10*time.Millisecond)
	}

	//Creating connection with Wallets servers (receive)
	rcvCheckBalance, _ = zmq.NewSocket(zmq.PULL)
	defer rcvCheckBalance.Close()
	rcvCheckBalance.SetRcvhwm(1100000)
	rcvCheckBalance.Connect(fmt.Sprintf("%v%v", LocalConfig.LocalTcpIpAddr, LocalConfig.BalancercvCheckBlncPort))

	//Creating connection with ... servers (send)
	sndBalance, _ = zmq.NewSocket(zmq.PUSH)
	defer sndBalance.Close()
	sndBalance.SetRcvhwm(1100000)
	sndBalance.Bind(LocalConfig.BalancesndBalance)

	//Creating connection with Wallets servers (send)
	sndCheckBalance, _ = zmq.NewSocket(zmq.PUSH)
	defer sndCheckBalance.Close()
	sndCheckBalance.SetRcvhwm(1100000)
	sndCheckBalance.Bind(fmt.Sprintf("%v%v", LocalConfig.LocalTcpIpAddr, LocalConfig.BalancesndCheckBlncPort))


	go CheckBalance()
	go rcvAccountAuthCh()
	runtime.GOMAXPROCS(runtime.NumCPU())

	var servers = map[int] *http.ServeMux{}

	for i := 0; i <= 199; i++  {
		servers[i] = http.NewServeMux()
		servers[i].HandleFunc("/", handleconn)
		port := 10000+i
		server_port := fmt.Sprintf(":%v", port)
		go func() {
			http.ListenAndServe(server_port, servers[i])
		}()
		time.Sleep(10*time.Millisecond)
	}

	for{
		globalCounter = 0
		time.Sleep(time.Second*20)
		fmt.Println("totalprocessed",globalCounter)
	}
}

/*
function CheckBalance
receive a request to send coins from the Wallet server.
Check balance and subtract from the balance coins
 */
func CheckBalance() {
	poller := zmq.NewPoller()
	poller.Add(rcvCheckBalance, zmq.POLLIN)
	for {
		_, err := poller.Poll(-1)
		if err != nil {
			fmt.Println("Socket error: ", err)
			continue //  Interrupted
		}
		msg,_ := rcvCheckBalance.RecvBytes(0)
		fmt.Println(msg)
		if len(msg) != 24 {
			continue
		}
		var response uint64
		response = 0
		userId := binary.LittleEndian.Uint64(msg[0:8])
		coin := binary.LittleEndian.Uint64(msg[8:16])
		value := binary.LittleEndian.Uint64(msg[16:24])
		if coin == 4 {
			value = value + 420000 //+fee ETH
		} else if coin == 1 {
			value = value + 10000 //+fee BTC
		}
		currencies.mux.Lock()
		current_balance := currencies.data[coin][userId]
		currencies.mux.Unlock()

		//continue //for tests
		if current_balance < value {
			response = 0
		} else {
			response = 1
			currencies.mux.Lock()
			currencies.data[coin][userId] = currencies.data[coin][userId] - value
			currencies.mux.Unlock()
		}

		buf := make([]byte, 32)
		binary.LittleEndian.PutUint64(buf[0:8], userId)
		binary.LittleEndian.PutUint64(buf[8:16], coin)
		binary.LittleEndian.PutUint64(buf[16:24], response)
		binary.LittleEndian.PutUint64(buf[24:32], value)

		if len(buf) == 32 {
			sndCheckBalance.SendBytes(buf,0)
		}

		buf = make([]byte, 32)
	}
}

/*
fucntion rcvAccountAuthCh
recive AUTH data from auth server
 */
func rcvAccountAuthCh() {

	poller := zmq.NewPoller()
	poller.Add(rcvAccountAuth, zmq.POLLIN)
	for i, _ := range rcvBalance {
		poller.Add(rcvBalance[i], zmq.POLLIN)
	}
	for {
		sockets, err := poller.Poll(-1)
		if err != nil {
			fmt.Println(err)
			continue //  Interrupted
		}

		for _, socket := range sockets {
			switch socket.Socket {
			case rcvAccountAuth:
				msg,_ := rcvAccountAuth.RecvBytes(0)
				s := strings.Split(string(msg), ",")
				if (s[2] != "" && s[1] != "") {
					userId, _ := strconv.ParseUint(s[2], 10, 64)
					authSessions[userId] = s[1]

					clients.mux.Lock()
					clients.data[s[1]] = Session{userId, s[1], time.Now().Unix()}
					clients.mux.Unlock()

					if (s[0] == "n") {
						if userId == 1 || userId == 2 {
							currencies.mux.Lock()
							currencies.data[2][userId]=100000000000
							currencies.data[1][userId]=100000000000
							currencies.data[4][userId]=100000000000
							currencies.mux.Unlock()
							WriteTolog <- exchange.ConvertBalanceMsg(2, userId, 1000000)
							WriteTolog <- exchange.ConvertBalanceMsg(1, userId, 100000000000)
							WriteTolog <- exchange.ConvertBalanceMsg(4, userId, 100000)
						}
					}
					if (s[0] == "a") {
						add, _ := strconv.ParseUint(s[1], 10, 64)
						coin, _ := strconv.ParseUint(s[3], 10, 64)

						currencies.mux.Lock()
						currencies.data[coin][userId]= currencies.data[coin][userId]+add
						currencies.mux.Unlock()

						WriteTolog <- exchange.ConvertBalanceMsg(uint16(coin), userId, currencies.data[coin][userId])
					}
					if (s[0] == "r") {
						remove, _ := strconv.ParseUint(s[1], 10, 64)
						coin, _ := strconv.ParseUint(s[3], 10, 64)

						currencies.mux.Lock()
						currencies.data[coin][userId]= currencies.data[coin][userId]+remove
						currencies.mux.Unlock()

						WriteTolog <- exchange.ConvertBalanceMsg(uint16(coin), userId, currencies.data[coin][userId])
					}
				}

			case rcvBalance[1]:
				result,_ := rcvBalance[1].RecvBytes(0) // eat the incoming message
				if len(result) >= 14 {
					rcvBalanceCon(result)
				}
			case rcvBalance[2]:
				result,_ := rcvBalance[2].RecvBytes(0) // eat the incoming message
				if len(result) >= 14 {
					rcvBalanceCon(result)
				}
			case rcvBalance[3]:
				result,_ := rcvBalance[3].RecvBytes(0) // eat the incoming message
				if len(result) >= 14 {
					rcvBalanceCon(result)
				}
			case rcvBalance[4]:
				result,_ := rcvBalance[4].RecvBytes(0) // eat the incoming message
				if len(result) >= 14 {
					rcvBalanceCon(result)
				}
			case rcvBalance[5]:
				result,_ := rcvBalance[5].RecvBytes(0) // eat the incoming message
				if len(result) >= 14 {
					rcvBalanceCon(result)
				}
			case rcvBalance[6]:
				result,_ := rcvBalance[6].RecvBytes(0) // eat the incoming message
				if len(result) >= 14 {
					rcvBalanceCon(result)
				}
			case rcvBalance[7]:
				result,_ := rcvBalance[7].RecvBytes(0) // eat the incoming message
				if len(result) >= 14 {
					rcvBalanceCon(result)
				}
			case rcvBalance[8]:
				result,_ := rcvBalance[8].RecvBytes(0) // eat the incoming message
				if len(result) >= 14 {
					rcvBalanceCon(result)
				}
			case rcvBalance[9]:
				result,_ := rcvBalance[9].RecvBytes(0) // eat the incoming message
				if len(result) >= 14 {
					rcvBalanceCon(result)
				}
			case rcvBalance[10]:
				result,_ := rcvBalance[10].RecvBytes(0) // eat the incoming message
				if len(result) >= 14 {
					rcvBalanceCon(result)
				}
			}
		}
	}
}

/*
function rcvBalanceCon
reading data that was received from sockets.
Receiving balance from Coins listening servers
 */
func rcvBalanceCon(result []byte) uint64 {
	var respBuf []byte
	for i:=0;i<len(result)/14;i++ {
		offset := i*14
		coin := binary.LittleEndian.Uint16(result[0+offset:2+offset])
		userId := binary.LittleEndian.Uint32(result[2+offset:6+offset])
		add := binary.LittleEndian.Uint64(result[6+offset:14+offset])
		if add != 0 && userId != 0 && coin != 0 {
			if coin >= 1000 {
				coin = coin / 1000
				currencies.mux.Lock()
				currencies.data[coin][userId] = currencies.data[coin][userId] - add
				currencies.mux.Unlock()
			} else {
				currencies.mux.Lock()
				currencies.data[coin][userId] = currencies.data[coin][userId] + add
				currencies.mux.Unlock()
			}

			WriteTolog <- exchange.ConvertBalanceMsg(coin, uint64(userId), currencies.data[coin][userId])

			lineBuf := make([]byte, 14)

			binary.LittleEndian.PutUint16(lineBuf[0:2], coin)
			binary.LittleEndian.PutUint32(lineBuf[2:6], userId)
			binary.LittleEndian.PutUint64(lineBuf[6:14], currencies.data[coin][userId])
			transactionsCount = transactionsCount + 1
			respBuf = append(respBuf, lineBuf...)
		}
	}
	return 0
}

/*
function getUSDinEUR_Price
API, covert USD to EUR
 */
func getUSDinEUR_Price() {
	for {

		client := &http.Client{}
		req, err := http.NewRequest("GET","https://api.ratesapi.io/api/latest", nil)
		if err != nil {
			log.Print(err)
			os.Exit(1)
		}

		req.Header.Set("Accepts", "application/json")
		resp, err := client.Do(req);
		if err != nil {
			fmt.Println("Error sending request to server")
			os.Exit(1)
		}
		fmt.Println(resp.Status);
		respBody, _ := ioutil.ReadAll(resp.Body)

		data := &USDinEUR_Pricetype{}
		err = json.Unmarshal(respBody, data)
		USDinEUR_Price = data.Rates.Usd
		time.Sleep(3 * time.Minute)
	}
}

/*
function countAllBalance#
converting coins and count total money in BTC
Coins percent comparing in USD equivalent
*/
func countAllBalance(userId uint64) string {
	var BalanceInUSD float64
	var BalanceInEUR float64
	var BalanceInBTC float64
	var balanceMap = make(map[uint64]float64);
	var balanceInPercent string
	BalanceInUSD = 0
	BalanceInEUR = 0
	BalanceInBTC = 0
	balanceInPercent = ""

	currencies.mux.Lock()
	for CoinID, data := range currencies.data {
		balanceValue := data[userId]
		CoinsPriceUSD.mux.Lock()
		_, is_ok := CoinsPriceUSD.data[uint64(CoinID)]
		CoinsPriceUSD.mux.Unlock()
		if is_ok {
			BalanceInUSDtmp := float64(float64(balanceValue) * float64(CoinsPriceUSD.data[uint64(CoinID)]))
			BalanceInUSD = BalanceInUSD + (BalanceInUSDtmp / float64(walletDelimetrMap.data[uint64(CoinID)]))
			balanceMap[uint64(CoinID)] = BalanceInUSDtmp / float64(walletDelimetrMap.data[uint64(CoinID)])
		}
	}
	currencies.mux.Unlock()

	BalanceInBTC = BalanceInUSD / float64(CoinsPriceUSD.data[uint64(1)])
	BalanceInEUR = BalanceInUSD / USDinEUR_Price
	//BalanceInUSD = 100%
	for CoinID, balance := range balanceMap {
		fmt.Println(balance)
		if balance != 0 {
			value := (100 * balance) / BalanceInUSD
			balanceInPercent = fmt.Sprintf("%v,%v|%v", CoinID, value, balanceInPercent)
		}
	}
	response := fmt.Sprintf("%v,%v,%v %v", BalanceInUSD, BalanceInEUR, BalanceInBTC, balanceInPercent)
	return response
}

/*
function handleconn
Handle HTTP requests from interface
 */
func handleconn(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Connection", "close")

	var status uint8  // 0 default status - buy/sell || status = 3 - user wants cancel order
	var oID uint64
	var response = 0
	oID = 0
	cmd := r.FormValue("data")
	data := strings.Split(cmd, ",")
	if (len(data) < 6) {
		w.Write([]byte("nFAIL182"))
		return //wrong request
	}
	//isBuy, session, userId_, pair_, price_, amount_
	userId, _ := strconv.ParseUint(data[2], 10, 0)

	if (clients.data[data[1]].userId != userId) {
		w.Write([]byte("FAIL197"))
		return
	}

	if RunType == "Wallets" {
		if (data[0] == "AllToUSD") {
			var response string
			CoinsPriceUSD.mux.Lock()
			for coinID, CoinData := range CoinsPriceUSD.data {
				response = fmt.Sprintf("%v,%v %v", coinID, CoinData, response)
			}
			CoinsPriceUSD.mux.Unlock()
			w.Write([]byte(response))
			return //wrong request
		}
		if (data[0] == "balanceW") {
			coin, _ := strconv.Atoi(data[3])

			currencies.mux.Lock()
			LastPrInUSD := float64(float64(currencies.data[coin][userId]) * float64(CoinsPriceUSD.data[uint64(coin)]))
			response := fmt.Sprintf("%v %v", currencies.data[coin][userId], LastPrInUSD)
			currencies.mux.Unlock()

			w.Write([]byte(response))
			return //wrong request
		}
		if (data[0] == "allbalanceW") {
			//get all balance in usd
			//get all balance in btc
			//get all balance in Uero
			response := countAllBalance(userId)
			w.Write([]byte(response))
			return //wrong request
		}
	}

	if RunType == "exchange" {
		if (data[0] == "balance") {
			coin, _ := strconv.Atoi(data[3])
			currencies.mux.Lock()
			response := fmt.Sprintf("%v",currencies.data[coin][userId])
			currencies.mux.Unlock()

			w.Write([]byte(response))
			return //wrong request
		}

		price, _ := strconv.ParseUint(data[4], 10, 64)
		amount, _ := strconv.ParseUint(data[5], 10, 64)
		pair, _ := strconv.Atoi(data[3])

		if (amount == 0) {
			w.Write([]byte("nFAIL316"))
			return //wrong request
		}
		//buySell, session, userId, pair, price, amount
		isBuy := true
		if (data[0] == "s") {
			isBuy = false
		}
		if data[0] == "cancel" {
			if data[6] == "0" {
				isBuy = false
			}
			oID,_ = strconv.ParseUint(data[7], 10, 64)
			status = 3
			response = 1
		} else {
			response = exchange.ProcessBuySell(isBuy, userId, price, amount, pair, data[1], status, &currencies.data, delimiter, globalCounter, WriteTolog, tcpPairs)
		}

		if (response == 1 ) {
			var blnc string
			if (isBuy == true) {
				blnc = fmt.Sprintf("%v",currencies.data[tcpPairs[pair].Market][userId])
			} else {
				blnc = fmt.Sprintf("%v",currencies.data[tcpPairs[pair].Coin][userId])
			}
			w.Write([]byte(blnc))
			transactionsCount = transactionsCount + 1

		} else {
			w.Write([]byte("Closed"))
		}

		if (response != 1){
			return
		}
		exchange.SendToObookCh<-exchange.ConvertToMsg(uint32(userId), uint16(pair), status, isBuy, uint64(globalCounter), oID, amount, amount, price)
	}
}

/*
function maxOpenFiles
limit for files
 */
func maxOpenFiles() {
	var rLimit syscall.Rlimit

	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		log.Println("Error Getting Rlimit ", err)
	}

	if rLimit.Cur < rLimit.Max {
		rLimit.Cur = rLimit.Max
		err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
		if err != nil {
			log.Println("Error Setting Rlimit ", err)
		}
	}
}

/*
DEBUG FUNCTION
 */
func write_debug_error(number string) {
	file, err := os.OpenFile("1a_test.db", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	defer file.Close()
	if err != nil {
		os.Exit(444)
	}

	b := make([]byte, 70)
	bufferedWriter := bufio.NewWriter(file)

	copy(b[:], []byte(number))
	_, err = bufferedWriter.Write( b, )
	if err != nil {
		os.Exit(445)
	}
	bufferedWriter.Flush()
	file.Close()
}