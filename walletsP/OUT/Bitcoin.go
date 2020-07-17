package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"wallets/walletsP/LocalConfig"
	"fmt"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcutil"
	zmq "github.com/pebbe/zmq4"
	"log"
	"os"
	"sync"
	"time"
)

var CustomNetwork = &chaincfg.TestNet3Params
var ColdWallet string // Colod Wallet address

//var ColdWalletsTxIn = make(map[uint64] UserTransaction) //Storage map for Cold Wallet transaction
var ColdWalletsTxIn struct {
	data map[uint64] UserTransaction
	mux sync.Mutex // mutex for concurrent map
}

//var usersTrnData = make(map[uint64] UserTransaction)
var usersTrnData struct {
	data map[uint64] UserTransaction
	mux sync.Mutex // mutex for concurrent map
}

var rcvTransactionHistory *zmq.Socket //ZMQ, socket variable
var sndOutTransaction *zmq.Socket //ZMQ, socket variable
var rcvColdTxIn *zmq.Socket //ZMQ, socket variable
var sndBalance *zmq.Socket //ZMQ, socket variable
var rcvTransaction *zmq.Socket //ZMQ, socket variable
var sndTransactionAnswer *zmq.Socket //ZMQ, socket variable

//New type for storege Users Transactions
type UserTransaction map[string]struct {
	UserId uint64
	TransactionHash string
	Vout uint32
}

/*
function readNextBytes
function return N bytes from file
 */
func readNextBytes(file *os.File, number int) []byte {
	bytes := make([]byte, number)
	size, err := file.Read(bytes)
	checkError(err) //check for errors

	if (size != number) {
		fmt.Println("size does not fit")
		return []byte{}
	}

	return bytes
}

/*
function createAddressMap
args: users addresses
function returns address map in btcutil format
 */
func createAddressMap(addresses map[string]uint64) map[btcutil.Address]btcutil.Amount {
	var addressMap = make(map[btcutil.Address]btcutil.Amount)
	defaultNet := CustomNetwork
	for address, amount := range addresses {
		c_address, err := btcutil.DecodeAddress(address, defaultNet)
		if err != nil {
			fmt.Println(err)
		}
		addressMap[c_address] = btcutil.Amount(amount)
	}
	return addressMap
}

/*
function CreateInputs
args: client(connection to bitcoin Core), user ID, Value, user Private Key
function returns map of inputs, map of addresses, private key of cold wallet, flag for only cold wallet input
 */
func CreateInputs(client *rpcclient.Client, userId uint64, Value uint64, PrivateKey string) ([]btcjson.TransactionInput, map[string]uint64, string, string) {
	//crate new VARS
	var inputs []btcjson.TransactionInput
	var addressMap = make(map[string]uint64)
	var txHash string
	var ColdKey string
	var PrivateKeyCMP string
	var Vout uint32
	var ReturnToAcc uint64
	var ValueCMP uint64
	ValueCMP = Value
	Value += 1000 //Add Fee cost to transaction value

	usersTrnData.mux.Lock()
	userTransactions, is_ok := usersTrnData.data[userId]
	if is_ok { //check if isset user' TxIn
		for _, TrData := range userTransactions { // range user' inputs
			if TrData.UserId == userId {
				PrivateKeyCMP = PrivateKey
				txHash = TrData.TransactionHash
				Vout = TrData.Vout
				blockHash, err := chainhash.NewHashFromStr(txHash)
				checkError(err)
				Tx, err := client.GetRawTransactionVerbose(blockHash)
				if Tx == nil {
					continue
				}
				VData := Tx.Vout[Vout]
				VData.Value = VData.Value * 100000000
				DataValue := uint64(VData.Value)
				if DataValue < Value {
					Value = Value - DataValue
					//add new input
					inputs = append(inputs, btcjson.TransactionInput{Txid:txHash, Vout:Vout, Witness: "", ScriptSig: ""})
					//continue searching
					continue
				} else if  DataValue >= Value {
					//add new input
					inputs = append(inputs, btcjson.TransactionInput{Txid:txHash, Vout:Vout, Witness: "", ScriptSig: ""})
					ReturnToAcc = DataValue - Value //-fee
					//if DataValue == Value {
						Value = 0
					//}
					if ReturnToAcc != 0 {
						//adding cold wallet address to return remaining money of input
						addressMap[ColdWallet] = ReturnToAcc
						//return money to account
					}
					break
				}
			}
		}
	}
	usersTrnData.mux.Unlock()

	if (ValueCMP - Value) != ValueCMP { //if there is not enough money on user inputs, we will use cold wallet entries
		ColdKey = LocalConfig.BitcoinColdWalletPass // Private Key of cold Wallet
		//get inputs from Cold Wallet
		ColdWalletsTxIn.mux.Lock()
		for _, TrData := range ColdWalletsTxIn.data[1]{
			txHash = TrData.TransactionHash
			Vout = TrData.Vout

			blockHash, err := chainhash.NewHashFromStr(txHash)
			checkError(err)
			Tx, err := client.GetRawTransactionVerbose(blockHash)
			if Tx == nil {
				continue
			}
			VData := Tx.Vout[Vout]
			VData.Value = VData.Value * 100000000
			DataValue := uint64(VData.Value)
			if DataValue < Value {
				Value = Value - DataValue
				inputs = append(inputs, btcjson.TransactionInput{Txid:txHash, Vout:Vout, Witness: "", ScriptSig: ""})
				//continue search
				continue
			} else if DataValue == Value {
				inputs = append(inputs, btcjson.TransactionInput{Txid:txHash, Vout:Vout, Witness: "", ScriptSig: ""})
				//all is good, go to transaction
				break
			} else if  DataValue > Value {
				inputs = append(inputs, btcjson.TransactionInput{Txid:txHash, Vout:Vout, Witness: "", ScriptSig: ""})
				ReturnToAcc = DataValue - Value //-fee
				addressMap[ColdWallet] = ReturnToAcc
				//all is good, return the money to account
				break
			}
		}
		ColdWalletsTxIn.mux.Unlock()
	}

	return inputs, addressMap, ColdKey, PrivateKeyCMP
}

/*
function CreateRawTransaction
args: send coins to, amount of coins, Fee, client(connection to bitcoin Core), sender private key, sender ID
function returns TxId or error
 */
func CreateRawTransaction(sendTo string, amount uint64, Fee btcutil.Amount, client *rpcclient.Client, PrivateKey string, userId uint64) *chainhash.Hash {
	var addressMap = make(map[string]uint64)
	currentTime := int64(0)
	inputs,addressMap, ColdKey, PrivateKeyCMP := CreateInputs(client, userId, amount, PrivateKey)

	if len(inputs) == 0{
		return  nil //no inputs, user can't send transaction...
	}
	_, is_ok := addressMap[sendTo]
	if is_ok {
		addressMap[sendTo] = addressMap[sendTo] + amount
	} else {
		addressMap[sendTo] = amount
	}

	ConvertedMap := createAddressMap(addressMap)
	client.SetTxFee(Fee)
	TransactionBody, err := client.CreateRawTransaction(inputs, ConvertedMap, &currentTime)
	if err != nil {
		return nil
	}
	if TransactionBody == nil {
		return nil
	}
	var SignInputs *[]string

	var ColdKeyFlag bool
	var PrivateKeyFlag bool

	ColdKeyFlag = false
	PrivateKeyFlag = false

	if len(PrivateKeyCMP) > 2 && len(ColdKey) > 3{
		SignInputs = &[]string{PrivateKey, ColdKey}
		ColdKeyFlag = true
		PrivateKeyFlag = true
	} else if len(ColdKey) > 3 {
		SignInputs = &[]string{ColdKey}
		ColdKeyFlag = true
	}  else if len(PrivateKeyCMP) > 3 {
		SignInputs = &[]string{PrivateKey}
		PrivateKeyFlag = true
	}

	SignTransaction, is_ok, err  := client.SignRawTransactionWithKeys(TransactionBody, SignInputs)

	if is_ok {
		var signedTx bytes.Buffer
		SignTransaction.Serialize(&signedTx)
		TxId,err := client.SendRawTransaction(SignTransaction, 0.0)
		if err != nil {
			return nil
		}

		//deleting used user' inputs
		for _, inputsdata := range inputs {
			usersTrnData.mux.Lock()
			if PrivateKeyFlag == true {
				_, is_ok :=  usersTrnData.data[userId]
				if is_ok {
					for _, userTrData := range usersTrnData.data[userId] {
						if inputsdata.Txid == userTrData.TransactionHash {
							delete(usersTrnData.data[userId], inputsdata.Txid)
						}
					}
				}
			}
			usersTrnData.mux.Unlock()

			ColdWalletsTxIn.mux.Lock()
			if ColdKeyFlag == true {
				for _, coldData := range ColdWalletsTxIn.data[1] {
					if inputsdata.Txid == coldData.TransactionHash {
						delete(ColdWalletsTxIn.data[1], inputsdata.Txid)
					}
				}
			}
			ColdWalletsTxIn.mux.Unlock()
		}

		return TxId
	}
	return nil
}

/*
function sndOutTrn
args: user ID, amount coins, Transaction ID, send coins To, send coins from
function of sending transaction data to the Bitcoin listening server…
 */
func sndOutTrn(userId uint64, Value uint64, TxData string, sndTo string, trFrom string) {
	buf := make([]byte, 148)

	binary.LittleEndian.PutUint64(buf[0:8], userId)
	binary.LittleEndian.PutUint64(buf[8:16], (Value*10)) //Value*10 - for correct history Value
	copy(buf[16:80], []byte(TxData))
	copy(buf[80:114], []byte(sndTo))
	copy(buf[114:148], []byte(trFrom))

	sndOutTransaction.SendBytes(buf, 0)
	time.Sleep(5*time.Second)
}

/*function listen_wallet
function of listening socket for the request to make a new transaction
 */
func listen_wallet(client  *rpcclient.Client) {
	poller := zmq.NewPoller()
	poller.Add(rcvTransaction, zmq.POLLIN)
	for {
		_, err := poller.Poll(-1)
		if err !=nil {
			fmt.Println("error, receve addr: ", err)
		}
		msg,_ := rcvTransaction.RecvBytes(0)
		if len(msg) != 136 {
			buf := make([]byte, 3)
			response := "0x0"
			copy(buf[0:3], []byte(response))
			sndTransactionAnswer.SendBytes(buf, 0)
			continue
		}
		userId := binary.LittleEndian.Uint64(msg[0:8])
		value := binary.LittleEndian.Uint64(msg[8:16])
		privateKeyStr := string(msg[16:68])
		sndTo := string(msg[68:102])
		trFrom :=  string(msg[102:136])

		fee := btcutil.Amount(1000)
		valueCPM := value
		value = (value / 10) - 1000

		Tx := CreateRawTransaction(sndTo, value, fee, client, privateKeyStr, userId)
		if Tx != nil {
			saveTxIn()
			saveColdTxIn()
		}

		if Tx == nil {
			buf := make([]byte, 3)
			response := "0x0"
			copy(buf[0:3], []byte(response))
			sndTransactionAnswer.SendBytes(buf, 0)
			lineBuf := make([]byte, 14)
			binary.LittleEndian.PutUint16(lineBuf[0:2], 1)
			binary.LittleEndian.PutUint32(lineBuf[2:6], uint32(userId))
			binary.LittleEndian.PutUint64(lineBuf[6:14],valueCPM)

			sndBalance.SendBytes(lineBuf,0)
			time.Sleep(1*time.Second)
		} else {
			buf := make([]byte, 64)
			var result string
			result = fmt.Sprintf("%v", Tx)
			copy(buf[0:64], []byte(result))
			go sndOutTrn(userId, value, result, sndTo, trFrom)
			sndTransactionAnswer.SendBytes(buf, 0)
			time.Sleep(1*time.Second)
		}
	}
}

/*
function reciveTrHistory
function of listening socket. Receive users' Tx In
 */
func reciveTrHistory() {
	poller := zmq.NewPoller()
	poller.Add(rcvTransactionHistory, zmq.POLLIN)
	for {
		_, err := poller.Poll(-1)
		if err != nil {
			continue //  Interrupted
		}
		msg, _ := rcvTransactionHistory.RecvBytes(0)
		if len(msg) == 8 {
			fmt.Println("rcvTransactionHistory Ping")
		}
		if len(msg) < 80 {
			continue
		}
		var offset uint64
		offset = 0
		for ij := 0; ij < (len(msg) / 80); ij++ {
			// Create TMP VARS for Transaction
			var TMPTrunsaction struct {
				UserId          uint64
				TransactionHash string
				Vout            uint32
			}
			var TMPUserTransaction = make(map[string]struct {
				UserId          uint64
				TransactionHash string
				Vout            uint32
			})
			//---------
			UserId := binary.LittleEndian.Uint64(msg[0+offset:8+offset])
			Vout := binary.LittleEndian.Uint64(msg[8+offset:16+offset])
			TransactionHash := string(bytes.Trim(msg[16+offset:80+offset], "\x00"))

			usersTrnData.mux.Lock()
			TMPTrunsaction.UserId = UserId
			TMPTrunsaction.Vout = uint32(Vout)
			TMPTrunsaction.TransactionHash = TransactionHash
			if len(usersTrnData.data[UserId]) >= 1 {
				for _, userData := range usersTrnData.data[UserId] {
					TMPUserTransaction[userData.TransactionHash] = userData
				}
				TMPUserTransaction[TransactionHash] = TMPTrunsaction
			} else {
				TMPUserTransaction[TransactionHash] = TMPTrunsaction
			}
			usersTrnData.data[UserId] = TMPUserTransaction
			usersTrnData.mux.Unlock()

			offset = offset + 80
		}
		saveTxIn()
	}
}

/*
function reciveColdTxIn
function of listening socket. Receive cold wallet Tx In
*/
func reciveColdTxIn() {
	poller := zmq.NewPoller()
	poller.Add(rcvColdTxIn, zmq.POLLIN)
	for {
		_, err := poller.Poll(-1)
		if err != nil {
			eer2 := fmt.Sprintf("Socket error: %v", err)
			fmt.Println(eer2)
			continue //  Interrupted
		}
		msg, _ := rcvColdTxIn.RecvBytes(0)
		if len(msg) == 8 {
			fmt.Println("Recive PING")
		}
		if len(msg) != 80 {
			continue
		}
		// Create TMP VARS for Transaction
		var TMPTrunsaction struct {
			UserId uint64
			TransactionHash string
			Vout uint32
		}

		var TMPUserTransaction = make(map[string]struct {
			UserId uint64
			TransactionHash string
			Vout uint32
		})

		UserId := binary.LittleEndian.Uint64(msg[0:8])
		Vout :=  binary.LittleEndian.Uint64(msg[8:16])
		TransactionHash := string(bytes.Trim(msg[16:80], "\x00"))

		ColdWalletsTxIn.mux.Lock()
		TMPTrunsaction.UserId = UserId
		TMPTrunsaction.Vout = uint32(Vout)
		TMPTrunsaction.TransactionHash = TransactionHash
		if len(ColdWalletsTxIn.data[UserId]) >= 1 {
			for _, userData := range ColdWalletsTxIn.data[UserId] {
				TMPUserTransaction[userData.TransactionHash] = userData
			}
			TMPUserTransaction[TransactionHash] = TMPTrunsaction
		} else {
			TMPUserTransaction[TransactionHash] = TMPTrunsaction
		}
		ColdWalletsTxIn.data[UserId] = TMPUserTransaction
		ColdWalletsTxIn.mux.Unlock()

		saveColdTxIn()
	}
}

/*
function saveColdTxIn
function is saving cold Wallet' inputs
*/
func saveColdTxIn() {
	file, err := os.OpenFile("../Bitcoin/ColdWalletsTxIn.db", os.O_CREATE|os.O_WRONLY, 0666)
	defer file.Close()
	if err != nil {
		log.Fatal("read file bans 1072",err)
	}
	file.Truncate(0)
	file.Seek(0,0)
	ColdWalletsTxIn.mux.Lock()
	for _, Data := range ColdWalletsTxIn.data {
		for _, Transactions := range Data {
			b := make([]byte, 8)
			binary.LittleEndian.PutUint64(b, Transactions.UserId)
			bufferedWriter := bufio.NewWriter(file)
			_, err = bufferedWriter.Write(b, )
			if err != nil {
				log.Fatal("writeFile 1019 ", err)
			}
			bufferedWriter.Flush()
			b = make([]byte, 8)
			binary.LittleEndian.PutUint64(b, uint64(Transactions.Vout))
			bufferedWriter = bufio.NewWriter(file)
			_, err = bufferedWriter.Write(b, )
			if err != nil {
				log.Fatal("writeFile 1019 ", err)
			}
			bufferedWriter.Flush()
			b = make([]byte, 64)
			copy(b[:], []byte(Transactions.TransactionHash))
			_, err = bufferedWriter.Write(b, )
			if err != nil {
				log.Fatal("writeFile 1026", err)
			}

			bufferedWriter.Flush()
		}
	}
	ColdWalletsTxIn.mux.Unlock()
	file.Close()
}

/*
function saveTxIn
function is saving user' inputs
*/
func saveTxIn() {
	file, err := os.OpenFile("../Bitcoin/UsersTxIn.db", os.O_CREATE|os.O_WRONLY, 0666)
	defer file.Close()
	if err != nil {
		log.Fatal("read file bans 1072",err)
	}
	file.Truncate(0)
	file.Seek(0,0)

	usersTrnData.mux.Lock()
	for _, Data := range usersTrnData.data {
		for _, Transactions := range Data{
			b := make([]byte, 8)
			binary.LittleEndian.PutUint64(b, Transactions.UserId)
			bufferedWriter := bufio.NewWriter(file)
			_, err = bufferedWriter.Write( b, )
			if err != nil{
			log.Fatal("writeFile 1019 ", err)
		}
			bufferedWriter.Flush()
			b = make([]byte, 8)
			binary.LittleEndian.PutUint64(b, uint64(Transactions.Vout))
			bufferedWriter = bufio.NewWriter(file)
			_, err = bufferedWriter.Write( b, )
			if err != nil{
			log.Fatal("writeFile 1019 ", err)
		}
			bufferedWriter.Flush()
			b = make([]byte, 64)
			copy(b[:], []byte(Transactions.TransactionHash))
			_, err = bufferedWriter.Write( b, )
			if err != nil{
			log.Fatal("writeFile 1026", err)
		}

			bufferedWriter.Flush()
		}
	}
	usersTrnData.mux.Unlock()
	file.Close()
}

/*
function loadColdWalletTxIn
function is loading cold wallet' inputs from file
*/
func loadColdWalletTxIn() {
	file, err := os.Open("../Bitcoin/ColdWalletsTxIn.db")
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}
	fi, err := file.Stat()
	if err != nil {
		log.Fatal("Could not obtain stat", err)
	}
	var i uint64
	ColdWalletsTxIn.mux.Lock()
	for i = 0; i < uint64(fi.Size()/(80)); i++ {
		// Create TMP VARS for Transaction
		var TMPTrunsaction struct {
			UserId uint64
			TransactionHash string
			Vout uint32
		}
		var TMPUserTransaction = make(map[string]struct {
			UserId uint64
			TransactionHash string
			Vout uint32
		})
		//---------
		data := readNextBytes(file, (80))
		UserId := binary.LittleEndian.Uint64(data[0:8])
		Vout := binary.LittleEndian.Uint64(data[8:16])
		TransactionHash := string(bytes.Trim(data[16:80], "\x00"))

		TMPTrunsaction.UserId = UserId
		TMPTrunsaction.Vout = uint32(Vout)
		TMPTrunsaction.TransactionHash = TransactionHash

		if len(ColdWalletsTxIn.data[UserId]) >= 1 {
			for _, userData := range ColdWalletsTxIn.data[UserId] {
				TMPUserTransaction[userData.TransactionHash] = userData
			}
			TMPUserTransaction[TransactionHash] = TMPTrunsaction
		} else {
			TMPUserTransaction[TransactionHash] = TMPTrunsaction
		}
		ColdWalletsTxIn.data[UserId] = TMPUserTransaction
	}
	ColdWalletsTxIn.mux.Unlock()
	file.Close()
}

/*
function loadTxIn
function is loading users' inputs from file
*/
func loadTxIn() {
	file, err := os.Open("../Bitcoin/UsersTxIn.db")
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}
	fi, err := file.Stat()
	if err != nil {
		log.Fatal("failed to get stat", err)
	}
	var i uint64

	usersTrnData.mux.Lock()
	for i = 0; i < uint64(fi.Size()/(80)); i++ {
		// Create TMP VARS for Transaction
		var TMPTrunsaction struct {
			UserId uint64
			TransactionHash string
			Vout uint32
		}
		var TMPUserTransaction = make(map[string]struct {
			UserId uint64
			TransactionHash string
			Vout uint32
		})

		data := readNextBytes(file, (80))
		UserId := binary.LittleEndian.Uint64(data[0:8])
		Vout := binary.LittleEndian.Uint64(data[8:16])
		TransactionHash := string(bytes.Trim(data[16:80], "\x00"))

		TMPTrunsaction.UserId = UserId
		TMPTrunsaction.Vout = uint32(Vout)
		TMPTrunsaction.TransactionHash = TransactionHash

		if len(usersTrnData.data[UserId]) >= 1 {
			for _, userData := range usersTrnData.data[UserId] {
				TMPUserTransaction[userData.TransactionHash] = userData
			}
			TMPUserTransaction[TransactionHash] = TMPTrunsaction
		} else {
			TMPUserTransaction[TransactionHash] = TMPTrunsaction
		}

		usersTrnData.data[UserId] = TMPUserTransaction
	}
	usersTrnData.mux.Unlock()
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func main() {
	//make maps with no nil data
	ColdWalletsTxIn.data = make(map[uint64] UserTransaction)
	usersTrnData.data = make(map[uint64] UserTransaction)
	//-----
	CustomNetwork.DefaultPort = LocalConfig.BitcoinDefaultPort // network Params
	ColdWallet = LocalConfig.BitcoinColdWallet //cold wallet address

	connCfg := &rpcclient.ConnConfig{ //bitcoin core connection config
		Host:         LocalConfig.BitcoinHost,
		User:         LocalConfig.BitcoinCoreUser,
		Pass:         LocalConfig.BitcoinCorePassw,
		HTTPPostMode: true, // Bitcoin core only supports HTTP POST mode
		DisableTLS:   true, // Bitcoin core does not provide TLS by default
	}

	client, err := rpcclient.New(connCfg, nil) //Create Connection
	checkError(err) //check for errors
	defer client.Shutdown()

	sndBalance, err = zmq.NewSocket(zmq.PUSH) //create new socket for return coins if we have some error
	defer sndBalance.Close()
	sndBalance.SetRcvhwm(1100000)
	err = sndBalance.Bind(fmt.Sprintf("%v%v", LocalConfig.LocalTcpIpAddr, LocalConfig.BTCBalanceOUTPort));
	checkError(err) //check for errors

	rcvTransactionHistory,_ = zmq.NewSocket(zmq.PULL) //create new socket for recieve new user TxIn
	defer rcvTransactionHistory.Close()
	rcvTransactionHistory.SetRcvhwm(1100002)
	rcvTransactionHistoryAddresses := fmt.Sprintf("%v%v", LocalConfig.LocalTcpIpAddr, LocalConfig.BTCrcvTrnsctnHstrOUTPort)
	err = rcvTransactionHistory.Connect(rcvTransactionHistoryAddresses)
	checkError(err) //check for errors

	rcvColdTxIn, _ = zmq.NewSocket(zmq.PULL) //create new socket for recieve new cold wallet TxIn
	defer rcvColdTxIn.Close()
	rcvColdTxIn.SetRcvhwm(1100001)
	rcvColdTxInAddress := fmt.Sprintf("%v%v", LocalConfig.LocalTcpIpAddr, LocalConfig.BTCrcvColdTxInPort)
	err = rcvColdTxIn.Connect(rcvColdTxInAddress)
	checkError(err) //check for errors

	sndOutTransaction, _ = zmq.NewSocket(zmq.PUSH)
	defer sndOutTransaction.Close()
	sndOutTransaction.SetRcvhwm(1100000)
	err = sndOutTransaction.Bind(fmt.Sprintf("%v%v", LocalConfig.LocalTcpIpAddr, LocalConfig.BTCsndOutTrnsctnPort))
	checkError(err) //check for errors

	sndTransactionAnswer, _ = zmq.NewSocket(zmq.PUSH)
	defer sndTransactionAnswer.Close()
	sndTransactionAnswer.SetRcvhwm(1100000)
	err = sndTransactionAnswer.Bind(fmt.Sprintf("%v%v", LocalConfig.LocalTcpIpAddr, LocalConfig.BTCsndTrnsctnAnswerPort))
	checkError(err) //check for errors

	rcvTransaction, _ = zmq.NewSocket(zmq.PULL)
	defer rcvTransaction.Close()
	rcvTransaction.SetRcvhwm(1100000)
	rcvAddress := fmt.Sprintf("%v%v", LocalConfig.LocalTcpIpAddr, LocalConfig.BTCrcvNewAddressOUTPort)
	err = rcvTransaction.Connect(rcvAddress)
	checkError(err) //check for errors

	loadColdWalletTxIn() //load TxIn from Cold Wallet
	loadTxIn() //load users TxIn
	go reciveTrHistory() //new concurrent process. listening sockets for new user TxIn
	go reciveColdTxIn() //new concurrent process. listening sockets for new cold wallet TxIn
	listen_wallet(client) //listening sockets for request to make new transaction

	os.Exit(1) // Exit with Error
	return
}