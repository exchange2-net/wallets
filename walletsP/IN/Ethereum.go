//3000000000000000000 - 3 eth
//10000
package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"wallets/walletsP/LocalConfig"
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var sndBalance *zmq.Socket
var rcvNewUserAddr *zmq.Socket
var sndTransactionHistory *zmq.Socket

var sendBufSlice struct {
	data []byte
	mux sync.Mutex
}

var usersTrunsactionsT struct {
	data map[string]UserTrunsaction
	mux sync.Mutex
}

var  usersTrunsactionsOUT struct {
	data map[string] UserTrunsaction
	mux sync.Mutex
}

var ConfirmedTransactions struct {
	data map[string]TransactionsTmp
	mux sync.Mutex
}

var ConfirmedTransactionsOUT struct {
	data map[string]TransactionsTmp
	mux sync.Mutex
}

var PendingTransactions struct {
	data map[string]TransactionsTmp
	mux sync.Mutex
}

var PendingTransactionsOUT struct {
	data map[string]TransactionsTmp
	mux sync.Mutex
}

var PendingAdderess struct {
	data map[uint64]string
	mux sync.Mutex
}

var UsersHasAddr struct {
	data map[uint64]HasAddress
	mux sync.Mutex
}

var balanceSlice  = make([]priceUpdate, 1, 1)

type TransactionsTmp struct {
	BlockHash        string `json:"blockHash"`
	BlockNumber      string `json:"blockNumber"`
	From             string `json:"from"`
	Gas              string `json:"gas"`
	GasPrice         string `json:"gasPrice"`
	Hash             string `json:"hash"`
	Input            string `json:"input"`
	Nonce            string `json:"nonce"`
	To               string `json:"to"`
	TransactionIndex string `json:"transactionIndex"`
	Value            string `json:"value"`
	V                string `json:"v"`
	R                string `json:"r"`
	S                string `json:"s"`
}

var fileOL *os.File
var file200 *os.File
var file1000 *os.File
var file10000 *os.File

type Storage struct {
	file200 *os.File
	file1000 *os.File
	file10000 *os.File
	name string
	historyRegistry [1000000] Records //рассчитано на один миллион пользователей
}
type Records struct {
	pos200 int64
	pos1000 int64
	pos10000 int64
	records uint64
	file uint8
}

var TransactionsHistory Storage

type transactions struct {
	id uint16
	transactionsHistory []byte
}

var userOrdersSlice  = make([]transactions, 1, 1)
var userTransactionMap [1000000] int

var TMPTransactionData struct {
	data []transactions
	mux sync.Mutex
}

type UserTrunsaction struct {
	UserId uint64
	TransactionHash string
	TrTime int64
	InOut int
	ConfirmedBlocks uint64
	TrFrom string
	TrTo string
	TrValue uint64
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

type priceUpdate struct{
	coin uint16
	add uint64
	userId uint32
}

/*
function sendRequest
Send API requests to infura
*/
func sendRequestETH(request string) []byte {
	url := fmt.Sprintf("%v%v", LocalConfig.InfuraRinkebyApiURL, LocalConfig.InfuraProjectID)
	str := request
	var jsonStr = []byte(str)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	return body
}

/*
function getBlock
Create API request and get ETH block by param(latest)
*/
func getBlockETH(blockNum string) string {
	result := sendRequestETH(fmt.Sprintf("{\"jsonrpc\":\"2.0\",\"method\":\"eth_getBlockByNumber\",\"params\":[\"%s\", true],\"id\":1}", blockNum))
	headerBody := strings.SplitN(string(result), "\r\n\r\n", 2)

	if len(result) < 1 {
		return ""
	}
	data := &LocalConfig.Block{}
	if len(headerBody)>0 {
		err := json.Unmarshal(result, data)
		checkError(err)
	}

	if len(data.Result.Number) < 3 {
		time.Sleep(20*time.Second)
		return ""
	}
	blockNumber, err := strconv.ParseUint(data.Result.Number[2:len(data.Result.Number)], 16, 64)
	checkError(err)
	// range block transactions
	for _, transaction := range data.Result.Transactions {

		reg_expr := regexp.MustCompile(".+x")
		reg_result := reg_expr.ReplaceAllString(transaction.Value, "")
		amount, _ := strconv.ParseUint(reg_result, 16, 64)

		// do not work with 0 transaction value
		if amount > 0 {
			//пcheck for correct transaction
			if len(transaction.Hash) != 66 {
				continue
			}
			//check if we have added a transaction to pending transactions
			PendingTransactions.mux.Lock()
			_, ok := PendingTransactions.data[transaction.Hash];
			PendingTransactions.mux.Unlock()
			if ok {
				continue
			}
			//check if we added the balance to the user account
			ConfirmedTransactions.mux.Lock()
			_, ok2 := ConfirmedTransactions.data[transaction.Hash];
			ConfirmedTransactions.mux.Unlock()
			if ok2 {
				continue
			}
			//finding our users' addresses and adding to the Pending map
			UsersHasAddr.mux.Lock()
			for userID, userData := range UsersHasAddr.data {
				for _, keysData := range userData {
					if string(transaction.To) == strings.ToLower(string(keysData.PublicKey)) {
						PendingAdderess.mux.Lock()
						PendingAdderess.data[userID] = transaction.To
						PendingAdderess.mux.Unlock()
						save_TransactionETH(transaction, data.Result.Timestamp, userID, 1, 0)
					}
					if string(transaction.From) == strings.ToLower(string(keysData.PublicKey)) {
						save_TransactionETH(transaction, data.Result.Timestamp, userID, 0, 0)
					}
				}
			}
			UsersHasAddr.mux.Unlock()
		}
	}
	return toHex(blockNumber)
}

/*
function change_trnsctn_status
loop is that it will check the number of confirmations
*/
func change_trnsctn_statusETH() {
	for {
		//check all OUR users transactions
		PendingTransactions.mux.Lock()
		PendingTransactionsTMP := PendingTransactions.data
		PendingTransactions.mux.Unlock()

		for _, TranXon := range PendingTransactionsTMP {
			ConfirmedTransactions.mux.Lock()
			_, ok2 := ConfirmedTransactions.data[TranXon.Hash];
			ConfirmedTransactions.mux.Unlock()
			if ok2 {
				delete(PendingTransactions.data, TranXon.Hash)
				continue
			}

			var status bool
			var destination uint64

			status, destination = check_transuctionETH(TranXon.BlockNumber,TranXon)

			if status == true {
				Uid :=  usersTrunsactionsT.data[TranXon.Hash].UserId //get user ID
				userSliceId := &userTransactionMap[Uid] //get user storege slice
				userSliceIdCMP := userTransactionMap[Uid]

				if (userSliceIdCMP == 0){
					userSliceIdCMP = len(userOrdersSlice)
					userTransactionMap[Uid] = userSliceIdCMP
					TMPTransactionData.data = append(TMPTransactionData.data, transactions{uint16(Uid), []byte{} })
					userOrdersSlice = append(userOrdersSlice, transactions{uint16(Uid), []byte{} })
				}

				TMPTransactionLIST := &TMPTransactionData.data[*userSliceId] //work with user storege list
				TMPOrdersLIST := &userOrdersSlice[*userSliceId]

				SaveConfirmedTransactions(TranXon)

				reg_expr := regexp.MustCompile(".+x")
				reg_result := reg_expr.ReplaceAllString(TranXon.Value, "")
				amount, _ := strconv.ParseUint(reg_result, 16, 64)

				usersTrunsactionsT.mux.Lock()
				usersTrunsactionsT.data[TranXon.Hash] = UserTrunsaction{UserId: usersTrunsactionsT.data[TranXon.Hash].UserId,
					TransactionHash: usersTrunsactionsT.data[TranXon.Hash].TransactionHash,
					InOut: 1,
					TrTime: usersTrunsactionsT.data[TranXon.Hash].TrTime,
					ConfirmedBlocks: 6,
					TrTo: usersTrunsactionsT.data[TranXon.Hash].TrTo,
					TrValue: usersTrunsactionsT.data[TranXon.Hash].TrValue,
					TrFrom:usersTrunsactionsT.data[TranXon.Hash].TrFrom,
					CoinID:4}
				usersTrunsactionsT.mux.Unlock()

				if amount != 0  && usersTrunsactionsT.data[TranXon.Hash].UserId != 0 {
					if usersTrunsactionsT.data[TranXon.Hash].InOut == 1{
						addFundsETH(TranXon.To, amount)
					}
				}

				//write comfirmed transactions
				buf := convertTransactionsToBin(usersTrunsactionsT.data[TranXon.Hash])
				TMPTransactionLIST.transactionsHistory = append(TMPTransactionLIST.transactionsHistory, buf...)
				TMPOrdersLIST.transactionsHistory = append(TMPTransactionLIST.transactionsHistory, buf...)
				//sending info to te Wallets server
				generateTrHbuff(usersTrunsactionsT.data[TranXon.Hash])

				delete(PendingTransactions.data, TranXon.Hash)
			} else {
				dstntn := destination+1
				if dstntn > usersTrunsactionsT.data[TranXon.Hash].ConfirmedBlocks {
					usersTrunsactionsT.mux.Lock()
					usersTrunsactionsT.data[TranXon.Hash] = UserTrunsaction{UserId: usersTrunsactionsT.data[TranXon.Hash].UserId,
						TransactionHash: usersTrunsactionsT.data[TranXon.Hash].TransactionHash,
						InOut: 1,
						TrTime: usersTrunsactionsT.data[TranXon.Hash].TrTime,
						ConfirmedBlocks: dstntn,
						TrTo: usersTrunsactionsT.data[TranXon.Hash].TrTo,
						TrValue: usersTrunsactionsT.data[TranXon.Hash].TrValue,
						TrFrom:usersTrunsactionsT.data[TranXon.Hash].TrFrom,
						CoinID:4}
					usersTrunsactionsT.mux.Unlock()

					generateTrHbuff(usersTrunsactionsT.data[TranXon.Hash])
				}
			}
		}
		time.Sleep(35*time.Second)
	}
}

/*
function change_trnsctn_statusOUT
the same function like change_trnsctn_status but for transactions that was sent from our servers
*/
func change_trnsctn_statusOUTETH() {
	for {
		//check all OUR users transactions
		PendingTransactionsOUT.mux.Lock()
		PendingTransactionsTMP := PendingTransactionsOUT.data
		PendingTransactionsOUT.mux.Unlock()

		for _, TranXon := range PendingTransactionsTMP {
			ConfirmedTransactions.mux.Lock()
			_, ok2 := ConfirmedTransactionsOUT.data[TranXon.Hash];
			ConfirmedTransactions.mux.Unlock()
			if ok2 {
				delete(PendingTransactionsOUT.data, TranXon.Hash)
				continue
			}

			var status bool
			var destination uint64

			status, destination = check_transuctionETH(TranXon.BlockNumber,TranXon)
			if status == true {
				Uid :=  usersTrunsactionsOUT.data[TranXon.Hash].UserId //get user ID
				userSliceId := &userTransactionMap[Uid] //get user storege slice
				userSliceIdCMP := userTransactionMap[Uid]

				if (userSliceIdCMP == 0){
					userSliceIdCMP = len(userOrdersSlice)
					userTransactionMap[Uid] = userSliceIdCMP
					TMPTransactionData.data = append(TMPTransactionData.data, transactions{uint16(Uid), []byte{} })
					userOrdersSlice = append(userOrdersSlice, transactions{uint16(Uid), []byte{} })
				}

				TMPTransactionLIST := &TMPTransactionData.data[*userSliceId] //work with user storege list
				TMPOrdersLIST := &userOrdersSlice[*userSliceId]

				SaveConfirmedTransactionsOUT(TranXon)

				usersTrunsactionsOUT.mux.Lock()
				usersTrunsactionsOUT.data[TranXon.Hash] = UserTrunsaction{UserId: usersTrunsactionsOUT.data[TranXon.Hash].UserId,
					TransactionHash: usersTrunsactionsOUT.data[TranXon.Hash].TransactionHash,
					InOut: 0,
					TrTime: usersTrunsactionsOUT.data[TranXon.Hash].TrTime,
					ConfirmedBlocks: 6,
					TrTo: usersTrunsactionsOUT.data[TranXon.Hash].TrTo,
					TrValue: usersTrunsactionsOUT.data[TranXon.Hash].TrValue,
					TrFrom:usersTrunsactionsOUT.data[TranXon.Hash].TrFrom,
					CoinID:4}
				usersTrunsactionsOUT.mux.Unlock()

				//write comfirmed transactions
				buf := convertTransactionsToBin(usersTrunsactionsOUT.data[TranXon.Hash])
				TMPTransactionLIST.transactionsHistory = append(TMPTransactionLIST.transactionsHistory, buf...)
				TMPOrdersLIST.transactionsHistory = append(TMPTransactionLIST.transactionsHistory, buf...)
				//sending info to te Wallets server
				generateTrHbuff(usersTrunsactionsOUT.data[TranXon.Hash])

				delete(PendingTransactionsOUT.data, TranXon.Hash)
			} else {
				dstntn := destination+1
				if dstntn > usersTrunsactionsOUT.data[TranXon.Hash].ConfirmedBlocks {
					usersTrunsactionsOUT.mux.Lock()
					usersTrunsactionsOUT.data[TranXon.Hash] = UserTrunsaction{UserId: usersTrunsactionsOUT.data[TranXon.Hash].UserId,
						TransactionHash: usersTrunsactionsOUT.data[TranXon.Hash].TransactionHash,
						InOut: 0,
						TrTime: usersTrunsactionsOUT.data[TranXon.Hash].TrTime,
						ConfirmedBlocks: dstntn,
						TrTo: usersTrunsactionsOUT.data[TranXon.Hash].TrTo,
						TrValue: usersTrunsactionsOUT.data[TranXon.Hash].TrValue,
						TrFrom:usersTrunsactionsOUT.data[TranXon.Hash].TrFrom,
						CoinID:4}
					usersTrunsactionsOUT.mux.Unlock()

					generateTrHbuff(usersTrunsactionsOUT.data[TranXon.Hash])
				}
			}
		}
		time.Sleep(35*time.Second)
	}
}

/*
function check_transuction
function that searches broken transactions and counts confirmations
*/
func check_transuctionETH(blockNum string, cmp_trnsctn TransactionsTmp) (bool, uint64) {
	request_TRN := fmt.Sprintf("{\"jsonrpc\":\"2.0\",\"method\":\"eth_getBlockByNumber\",\"params\":[\"%s\", true],\"id\":1}", blockNum)
	result_TRN := sendRequestETH(request_TRN)

	request_Crrnt := "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getBlockByNumber\",\"params\":[\"latest\", true],\"id\":1}"
	result_Crrnt := sendRequestETH(request_Crrnt)

	headerBody_TRN := strings.SplitN(string(result_TRN), "\r\n\r\n", 2)
	headerBody_Crrnt := strings.SplitN(string(result_Crrnt), "\r\n\r\n", 2)

	if len(result_TRN) < 1 {
		return false, 0
	}
	if len(result_Crrnt) < 1 {
		return false, 0
	}

	data_TRN := &LocalConfig.Block{}
	data_Crrnt := &LocalConfig.Block{}

	if len(headerBody_TRN) > 0 {
		err := json.Unmarshal(result_TRN, data_TRN)
		checkError(err)
	}
	if len(headerBody_Crrnt)>0 {
		err := json.Unmarshal(result_Crrnt, data_Crrnt)
		checkError(err)
	}

	if len(data_TRN.Result.Number) < 3 {
		return false, 0
	}
	if len(data_Crrnt.Result.Number) < 3 {
		return false, 0
	}

	TRN_blockNumber, err := strconv.ParseUint(data_TRN.Result.Number[2:len(data_TRN.Result.Number)], 16, 64)
	checkError(err)
	Crrnt_blockNumber, err := strconv.ParseUint(data_Crrnt.Result.Number[2:len(data_Crrnt.Result.Number)], 16, 64)
	checkError(err)

	var destination uint64
	//count confirmations
	destination = Crrnt_blockNumber - TRN_blockNumber
	//count confirmations
	if destination >= 6 {
		points_of_distinction := 0
		//seach transaction chenges
		for _, transaction := range data_TRN.Result.Transactions {
			if transaction.Hash == cmp_trnsctn.Hash {
				if transaction.BlockHash == cmp_trnsctn.BlockHash {
					points_of_distinction++ //1
				}
				if transaction.From == cmp_trnsctn.From{
					points_of_distinction++ //2
				}
				if transaction.BlockNumber == cmp_trnsctn.BlockNumber {
					points_of_distinction++ //3
				}
				if transaction.Gas == cmp_trnsctn.Gas {
					points_of_distinction++ //4
				}
				if transaction.GasPrice == cmp_trnsctn.GasPrice {
					points_of_distinction++ //5
				}
				if transaction.Input == cmp_trnsctn.Input {
					points_of_distinction++  //6
				}
				if transaction.Nonce == cmp_trnsctn.Nonce {
					points_of_distinction++ //7
				}
				if transaction.To == cmp_trnsctn.To {
					points_of_distinction++ //8
				}
				if transaction.TransactionIndex == cmp_trnsctn.TransactionIndex {
					points_of_distinction++ //9
				}
				if transaction.Value == cmp_trnsctn.Value {
					points_of_distinction++ //10
				}
				if transaction.V == cmp_trnsctn.V {
					points_of_distinction++ //11
				}
				if transaction.R == cmp_trnsctn.R {
					points_of_distinction++ //12
				}
				if transaction.S == cmp_trnsctn.S {
					points_of_distinction++ //13
				}
				if points_of_distinction == 13 {
					return true, destination
				}
			}
		}
	} else {
		return false, destination
	}
	return false, 0
}

/*
function readNextBytes
function returns N bytes from file
*/
func readNextBytesETH(file *os.File, number int) []byte {
	bytes := make([]byte, number)

	size, err := file.Read(bytes)
	if err != nil {
		log.Fatal(err, size)
	}
	if (size != number) {
		return []byte{}
	}
	return bytes
}

/*
function addFunds
search a pending address in the user address list
call function addToBalance
*/
func addFundsETH(address string, funds uint64) {
	PendingAdderess.mux.Lock()
	for userID, P_address := range PendingAdderess.data {
		if address == P_address {
			addToBalanceETH(4, uint32(userID), uint64(funds/1000000000))
		}
	}
	PendingAdderess.mux.Unlock()
}

/*
function addToBalance
add data to slice
*/
func addToBalanceETH(coin uint16, userId uint32, add uint64) {
	balanceSlice = append(balanceSlice, priceUpdate{coin:coin, userId:userId, add:add})
}
/*
function toHex
convert uint64 value to hex value
*/
func toHex(i uint64) string {
	return fmt.Sprintf("0x%x", i)
}

/*
function loadLastBlock
load the last processed block from database
*/
func loadLastBlockETH() string {
	b, err := ioutil.ReadFile("../Ethereum/lastETHBlock.log")
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return string(b)
}

/*
function saveLastBlock
saving the last processed block in database
*/
func saveLastBlockETH(block string) {
	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile("../Ethereum/lastETHBlock.log", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	_, err = f.Write([]byte(block))
	if err != nil {
		log.Fatal(err)
	}

	f.Close()
}

/*
function generateBalanceBuf
Convert data to binary format
*/
func generateBalanceBufETH() []byte {
	var buf []byte

	for i := range balanceSlice {
		if(i==0){
			continue
		}
		lineBuf := make([]byte, 14)

		binary.LittleEndian.PutUint16(lineBuf[0:2], balanceSlice[i].coin)
		binary.LittleEndian.PutUint32(lineBuf[2:6], balanceSlice[i].userId)
		binary.LittleEndian.PutUint64(lineBuf[6:14], balanceSlice[i].add)

		buf = append(buf, lineBuf...)
	}

	balanceSlice  = make([]priceUpdate, 1, 1)
	return buf
}

/*
function write_transaction_to
saved transaction hash to database
*/
func write_transaction_toETH(filename string, data map[string]TransactionsTmp) {
	file, err := os.OpenFile(filename, os.O_WRONLY, 0666)
	checkError(err)
	defer file.Close()

	for _, TranXon := range data {
		b := make([]byte, 66)
		copy(b[:], []byte(TranXon.BlockHash))
		bufferedWriter := bufio.NewWriter(file)
		_, err = bufferedWriter.Write( b, )
		checkError(err)
		bufferedWriter.Flush()

		b = make([]byte, 15) //trim
		copy(b[:], []byte(TranXon.BlockNumber))
		bufferedWriter = bufio.NewWriter(file)
		_, err = bufferedWriter.Write( b, )
		checkError(err)
		bufferedWriter.Flush()

		b = make([]byte, 42)
		copy(b[:], []byte(TranXon.From))
		bufferedWriter = bufio.NewWriter(file)
		_, err = bufferedWriter.Write( b, )
		checkError(err)
		bufferedWriter.Flush()

		b = make([]byte, 15) //trim
		copy(b[:], []byte(TranXon.Gas))
		bufferedWriter = bufio.NewWriter(file)
		_, err = bufferedWriter.Write( b, )
		checkError(err)
		bufferedWriter.Flush()

		b = make([]byte, 15) //trim
		copy(b[:], []byte(TranXon.GasPrice))
		bufferedWriter = bufio.NewWriter(file)
		_, err = bufferedWriter.Write( b, )
		checkError(err)
		bufferedWriter.Flush()

		b = make([]byte, 66)
		copy(b[:], []byte(TranXon.Hash))
		bufferedWriter = bufio.NewWriter(file)
		_, err = bufferedWriter.Write( b, )
		checkError(err)
		bufferedWriter.Flush()

		b = make([]byte, 350) //trim
		copy(b[:], []byte(TranXon.Input))
		bufferedWriter = bufio.NewWriter(file)
		_, err = bufferedWriter.Write( b, )
		checkError(err)
		bufferedWriter.Flush()

		b = make([]byte, 15) //trim
		copy(b[:], []byte(TranXon.Nonce))
		bufferedWriter = bufio.NewWriter(file)
		_, err = bufferedWriter.Write( b, )
		checkError(err)
		bufferedWriter.Flush()

		b = make([]byte, 42)
		copy(b[:], []byte(TranXon.To))
		bufferedWriter = bufio.NewWriter(file)
		_, err = bufferedWriter.Write( b, )
		checkError(err)
		bufferedWriter.Flush()

		b = make([]byte, 10) //trim
		copy(b[:], []byte(TranXon.TransactionIndex))
		bufferedWriter = bufio.NewWriter(file)
		_, err = bufferedWriter.Write( b, )
		checkError(err)
		bufferedWriter.Flush()

		b = make([]byte, 35) //trim
		copy(b[:], []byte(TranXon.Value))
		bufferedWriter = bufio.NewWriter(file)
		_, err = bufferedWriter.Write( b, )
		checkError(err)
		bufferedWriter.Flush()

		b = make([]byte, 10) //trim
		copy(b[:], []byte(TranXon.V))
		bufferedWriter = bufio.NewWriter(file)
		_, err = bufferedWriter.Write( b, )
		checkError(err)
		bufferedWriter.Flush()

		b = make([]byte, 70) //trim
		copy(b[:], []byte(TranXon.R))
		bufferedWriter = bufio.NewWriter(file)
		_, err = bufferedWriter.Write( b, )
		checkError(err)
		bufferedWriter.Flush()

		b = make([]byte, 70) //trim
		copy(b[:], []byte(TranXon.S))
		bufferedWriter = bufio.NewWriter(file)
		_, err = bufferedWriter.Write( b, )
		checkError(err)
		bufferedWriter.Flush()
		//total 821 bytes
	}

	file.Close()
}

func SaveConfirmedTransactionsOUT(transaction TransactionsTmp) {
	ConfirmedTransactionsOUT.mux.Lock()
	ConfirmedTransactionsOUT.data[transaction.Hash] = transaction
	ConfirmedTransactionsOUT.mux.Unlock()
}

/*
function SaveConfirmedTransactions
saved Confirmed transaction hash to database
*/
func SaveConfirmedTransactions(transaction TransactionsTmp) {
	ConfirmedTransactions.mux.Lock()
	ConfirmedTransactions.data[transaction.Hash] = transaction
	ConfirmedTransactions.mux.Unlock()
}

func load_trunsuction_to(filename string, tr_type string) {
	file, err := os.Open(filename)
	checkError(err)
	defer file.Close()
	fi, err := file.Stat()
	checkError(err)
	var i uint64

	var transaction TransactionsTmp

	for i = 0; i < uint64(fi.Size()/(821)); i++ {
		data := readNextBytesETH(file, 821)

		transaction.BlockHash = string(data[0:66])
		transaction.BlockNumber = string(bytes.Trim(data[66:81], "\x00"))
		transaction.From = string(bytes.Trim(data[81:123], "\x00"))
		transaction.Gas = string(bytes.Trim(data[123:138], "\x00"))
		transaction.GasPrice = string(bytes.Trim(data[138:153], "\x00"))
		transaction.Hash = string(bytes.Trim(data[153:219], "\x00"))
		transaction.Input = string(bytes.Trim(data[219:569], "\x00"))
		transaction.Nonce = string(bytes.Trim(data[569:584], "\x00"))
		transaction.To = string(bytes.Trim(data[584:626], "\x00"))
		transaction.TransactionIndex = string(bytes.Trim(data[626:636], "\x00"))
		transaction.Value = string(bytes.Trim(data[636:671], "\x00"))
		transaction.V = string(bytes.Trim(data[671:681], "\x00"))
		transaction.R = string(bytes.Trim(data[681:751], "\x00"))
		transaction.S = string(bytes.Trim(data[751:821], "\x00"))

		if tr_type == "Confirmed" {
			ConfirmedTransactions.data[transaction.Hash] = transaction
		}
		if tr_type == "ConfirmedOUT" {
			ConfirmedTransactionsOUT.data[transaction.Hash] = transaction
		}

		if tr_type == "Pending" {
			loopForLoad(transaction)
			//PendingTransactions.data[transaction.Hash] = transaction
		}
		if tr_type == "PendingOUT" {
			loopForLoad(transaction)
			//PendingTransactionsOUT.data[transaction.Hash] = transaction
		}

	}
	file.Close()
}

func loopForLoad(transaction TransactionsTmp) {
	UsersHasAddr.mux.Lock()
	time := fmt.Sprintf("%v", transaction.BlockNumber)
	for userID, userData := range UsersHasAddr.data {
		for _, keysData := range userData {
			if string(transaction.To) == strings.ToLower(string(keysData.PublicKey)) {
				save_TransactionETH(transaction, time, userID, 1, 1)
			}
			if string(transaction.From) == strings.ToLower(string(keysData.PublicKey)) {
				save_TransactionETH(transaction, time, userID, 0, 1)
			}
		}
	}
	UsersHasAddr.mux.Unlock()
}

func LoadConfirmedTransactionsETH() {
	filename := "../Ethereum/ConfirmedTransactions.db"
	load_trunsuction_to(filename, "Confirmed")
}

func LoadConfirmedTransactionsOUTETH() {
	filename := "../Ethereum/ConfirmedTransactionsOUT.db"
	load_trunsuction_to(filename, "ConfirmedOUT")
}

/*
function save_Transaction
saving the user transaction to database
sending Tx ata to the wallets server
*/
func save_TransactionETH(transaction TransactionsTmp, Blocktime string, userID uint64, InOrOut int, isLoad int) {
	reg_expr := regexp.MustCompile(".+x")
	reg_result := reg_expr.ReplaceAllString(transaction.Value, "")
	amount, _ := strconv.ParseUint(reg_result, 16, 64)
	TrValue := uint64(amount/1000000000)
	Time ,_ := strconv.ParseUint(Blocktime[2:len(Blocktime)], 16, 64)

	var saveTrTo string
	var saveTrFrom string

	if InOrOut == 1 {
		PendingTransactions.mux.Lock()
		PendingTransactions.data[transaction.Hash] = transaction
		PendingTransactions.mux.Unlock()

		saveTrTo = transaction.To
		saveTrFrom = transaction.From

		usersTrunsactionsT.mux.Lock()
		usersTrunsactionsT.data[transaction.Hash] = UserTrunsaction{UserId: userID,
			TransactionHash: transaction.Hash,
			InOut: 1,
			TrTime: int64(Time),
			ConfirmedBlocks: 1,
			TrTo: saveTrTo,
			TrValue: TrValue,
			TrFrom:saveTrFrom,
			CoinID:4}
		usersTrunsactionsT.mux.Unlock()
		if isLoad == 0 {
			generateTrHbuff(usersTrunsactionsT.data[transaction.Hash])
		}
	} else {
		PendingTransactionsOUT.mux.Lock()
		PendingTransactionsOUT.data[transaction.Hash] = transaction
		PendingTransactionsOUT.mux.Unlock()

		saveTrTo = transaction.To
		saveTrFrom = transaction.From

		usersTrunsactionsOUT.mux.Lock()
		usersTrunsactionsOUT.data[transaction.Hash] = UserTrunsaction{UserId: userID,
			TransactionHash: transaction.Hash,
			InOut: 0,
			TrTime: int64(Time),
			ConfirmedBlocks: 1,
			TrTo: saveTrTo,
			TrValue: TrValue,
			TrFrom:saveTrFrom,
			CoinID:4}
		usersTrunsactionsOUT.mux.Unlock()

		if isLoad == 0 {
			generateTrHbuff(usersTrunsactionsOUT.data[transaction.Hash])
		}
	}
}

func load_TransactionsETH() {
	filename := "../Ethereum/transactions.db"
	load_trunsuction_to(filename, "Pending")
}

func load_TransactionsOUTETH() {
	filename := "../Ethereum/transactionsOUT.db"
	load_trunsuction_to(filename, "PendingOUT")
}

/*
function generateTrHbuff
converting the transaction info to a binary type and sending it to Wallets server
*/
func generateTrHbuff(transaction UserTrunsaction) {
	buf := make([]byte, 198)

	binary.LittleEndian.PutUint64(buf[0:8], transaction.UserId)
	binary.LittleEndian.PutUint64(buf[8:16], uint64(transaction.InOut))
	binary.LittleEndian.PutUint64(buf[16:24], transaction.ConfirmedBlocks)
	binary.LittleEndian.PutUint64(buf[24:32], uint64(transaction.TrTime))
	copy(buf[32:98], []byte(transaction.TransactionHash))
	copy(buf[98:140], []byte(transaction.TrFrom))
	copy(buf[140:182], []byte(transaction.TrTo))
	binary.LittleEndian.PutUint64(buf[182:190], transaction.TrValue)
	binary.LittleEndian.PutUint64(buf[190:198], transaction.CoinID)

	if len(buf) == 198 {
		if len(sendBufSlice.data) <= 1 {
			sendBufSlice.mux.Lock()
			sendBufSlice.data = buf
			sendBufSlice.mux.Unlock()
		} else {
			sendBufSlice.mux.Lock()
			sendBufSlice.data = append(sendBufSlice.data, buf...)
			sendBufSlice.mux.Unlock()
		}
	}
}

/*
function addNewUserToMap
add a new user address to the address list
*/
func addNewUserToMap(address string, privateKey string, userId uint64) {
	var tmp TMPstruct

	var keysData = make(map[string]struct {
		CoinID uint64
		PublicKey string
	})

	tmp.CoinID = 4
	tmp.PublicKey = address
	UsersHasAddr.mux.Lock()
	if len(UsersHasAddr.data[userId]) > 0 {
		for key, prevData := range UsersHasAddr.data[userId] {
			keysData[key] = prevData
		}
		keysData[privateKey] = tmp
	} else {
		keysData[privateKey] = tmp
	}

	UsersHasAddr.data[userId] = keysData
	UsersHasAddr.mux.Unlock()
}

/*
function rcvnewAddr
handling requests from the Wallet server for adding new user address to the address list
*/
func rcvnewAddr() {
	poller := zmq.NewPoller()
	poller.Add(rcvNewUserAddr, zmq.POLLIN)
	for {
		_, err := poller.Poll(-1)
		if err !=nil {
			fmt.Println(err)
		}
		msg,_ := rcvNewUserAddr.RecvBytes(0)
		if len(msg) != 114 {
			continue
		}
		userId := binary.LittleEndian.Uint64(msg[0:8])
		address := string(msg[8:50])
		privateKey := string(msg[50:114])

		go addNewUserToMap(address, privateKey, userId)
	}
}

/*
function returnOffset
function returns write offset
*/
func returnOffset(size uint64, records uint64) (uint64, uint64, uint64, uint64, uint64, uint64, uint64, uint64, int64, int64, int64) {
	begin := uint64(0)
	end := records + size
	var begin200 uint64
	var begin1000 uint64
	var begin10000 uint64
	var beginFile uint64
	var offset200 uint64
	var offset1000 uint64
	var offset10000 uint64
	var end200 uint64
	var end1000 uint64
	var end10000 uint64
	var endFile uint64

	if (200 > records) {
		offset200 = records
		if (end < 200) {
			return begin, size, 0, 0, 0, 0, 0, 0, int64(offset200*192), 0, 0
		} else {
			begin200 = 0
			end200 = 200-records
			begin1000 = end200
			begin = end200
		}
	}
	if (1200 > records) {
		offset1000 = records-200+1
		if (records < 200) {
			offset1000 = 0
		}
		if (end < 1200) {
			return begin200, end200, begin, size, 0, 0, 0, 0, int64(offset200*192), int64(offset1000*192), 0
		} else {
			end1000 = 1200-records
			begin10000 = end1000
			begin = end1000
		}
	}
	if (11200 > records) {
		offset10000  = records-1200+1
		if (records < 1200) {
			offset10000 = 0
		}
		if (end < 11200) {
			return begin200, end200, begin1000, end1000, begin, size, 0, 0, int64(offset200*192), int64(offset1000*192), int64(offset10000*192)
		} else {
			end10000 = 11200-records
			beginFile = end10000
			endFile = size
		}
	}
	if (11200 <= records) {
		endFile = size
	}
	return begin200, end200, begin1000, end1000, begin10000, end10000, beginFile, endFile, int64(offset200*192), int64(offset1000*192), int64(offset10000*192)
}

/*
function convertTransactionsToBin
function for converting data from map to binary array
*/
func convertTransactionsToBin(trData UserTrunsaction) []byte {
	var tmpBuf = make([]byte, 192)

	Uid := trData.UserId
	if len(trData.TransactionHash) > 10 && trData.UserId > 2 {
		binary.LittleEndian.PutUint64(tmpBuf[0:8], Uid)
		binary.LittleEndian.PutUint16(tmpBuf[8:10], uint16(trData.InOut))
		binary.LittleEndian.PutUint16(tmpBuf[10:12], uint16(trData.CoinID))
		binary.LittleEndian.PutUint16(tmpBuf[12:14], uint16(trData.ConfirmedBlocks))
		binary.LittleEndian.PutUint64(tmpBuf[14:22], uint64(trData.TrTime))
		binary.LittleEndian.PutUint64(tmpBuf[22:30], trData.TrValue)

		copy(tmpBuf[30:72], []byte(trData.TrTo)) //len 42
		copy(tmpBuf[72:114], []byte(trData.TrFrom)) //len 42
		copy(tmpBuf[114:192], []byte(trData.TransactionHash)) //len 66 + 12 empty bytes

		return tmpBuf
	}
	return tmpBuf
}

/*
function writeAllData
write transaction history
*/
func writeAllDataETH() {
	TMPTransactionData.mux.Lock()
	sliceCpy := make([]transactions, len(TMPTransactionData.data))
	TMPTransactionData.mux.Unlock()

	var flag bool
	flag = false
	TMPTransactionData.mux.Lock()
	for n:=0;n<len(sliceCpy);n++ {
		if len(TMPTransactionData.data[n].transactionsHistory) <= 1 {
			continue
		}

		sliceCpy[n] = transactions{TMPTransactionData.data[n].id, TMPTransactionData.data[n].transactionsHistory}
		TMPTransactionData.data[n] =  transactions{TMPTransactionData.data[n].id, []byte{}}
		flag = true
	}
	TMPTransactionData.mux.Unlock()
	if flag == false {
		return
	}
	TMPTransactionData.mux.Lock()
	for n := range sliceCpy {
		if sliceCpy[n].transactionsHistory != nil  && len(sliceCpy[n].transactionsHistory) >= 192 {
			if len(sliceCpy[n].transactionsHistory) % 192 != 0{
				os.Exit(1)
			}
			_ = writeDataToFileETH(sliceCpy[n].transactionsHistory, sliceCpy[n].id, &TransactionsHistory)
			sliceCpy[n].transactionsHistory = make([]byte,1)
		}

	}
	writeLogTofileETH(&TransactionsHistory)
	TMPTransactionData.mux.Unlock()
}

/*
function writeLogTofile
write data to map database for transactions storage
*/
func writeLogTofileETH(dataStorage *Storage) {
	//Start := time.Now()
	file, err := os.OpenFile(fmt.Sprintf("../Ethereum/%vMap.db",dataStorage.name), os.O_WRONLY|os.O_CREATE, 0666)
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}

	buffer := make([]byte,48)
	var replace bool
	for n := range userOrdersSlice {
		userId := userOrdersSlice[n].id
		recordData := &dataStorage.historyRegistry[userId]

		if (recordData.records == 0){
			continue
		}
		writeBuf := make([]byte,48)
		binary.LittleEndian.PutUint64(writeBuf[0:8], uint64(recordData.pos200))
		binary.LittleEndian.PutUint64(writeBuf[8:16], uint64(recordData.pos1000))
		binary.LittleEndian.PutUint64(writeBuf[16:24], uint64(recordData.pos10000))
		binary.LittleEndian.PutUint64(writeBuf[24:32], recordData.records)
		binary.LittleEndian.PutUint64(writeBuf[32:40], uint64(recordData.file))
		binary.LittleEndian.PutUint64(writeBuf[40:48], uint64(userId))

		if (replace != true) {
			buffer = writeBuf
			replace = true
		}else{
			buffer=append(buffer,writeBuf...)
		}
	}

	file.Write(buffer)
	file.Close()
}

/*
function writeDataToFile
write data to a specific storage
*/
func writeDataToFileETH(listTransactions []byte, userId uint16, dataFiles *Storage) (int) {
	if listTransactions == nil {
		return 1
	}
	recordData := &dataFiles.historyRegistry[userId]
	begin200, end200, begin1000, end1000, begin10000, end10000, beginFile, endFile, offset200, offset1000, offset10000 := returnOffset(uint64(len(listTransactions)/192), recordData.records)

	if uint64(len(listTransactions)/192) == recordData.records {
		fmt.Println("DEBUG TRUE")
		fmt.Println(recordData.records)
	} else {
		fmt.Println("DEBUG FLASE")
		fmt.Println(recordData.records)
		fmt.Println(uint64(len(listTransactions)/192))
	}

	if (end200 != 0) {
		if (recordData.file != 1) {
			recordData.file = 1
			fi, err := dataFiles.file200.Stat()
			if (err != nil) {
				log.Fatal(err)
			}

			recordData.pos200 = fi.Size()
			dataFiles.file200.Seek(0, 2)


			_, err = dataFiles.file200.Write(make([]byte,200*192))
			if err != nil {
				log.Fatal(err)
			}
		}
		dataFiles.file200.Seek(recordData.pos200 + offset200, 0)
		dataFiles.file200.Write(listTransactions[begin200*192:end200*192])
	}
	if (end1000 != 0) {
		if (recordData.file != 2) {
			recordData.file = 2
			fi, err := dataFiles.file1000.Stat()

			if (err != nil) {
				log.Fatal(err)
			}
			recordData.pos1000 = fi.Size()
			dataFiles.file1000.Seek(0, 2)
			_, err = dataFiles.file1000.Write(make([]byte,1000*192))
			if err != nil {
				log.Fatal(err)
			}
		}
		dataFiles.file1000.Seek(recordData.pos1000 + offset1000, 0)
		dataFiles.file1000.Write(listTransactions[begin1000*192:end1000*192])
	}
	if ( end10000 != 0) {
		if (recordData.file != 3) {
			recordData.file = 3
			fi, err := dataFiles.file10000.Stat()
			if (err != nil) {
				log.Fatal(err)
			}
			recordData.pos10000 = fi.Size()
			dataFiles.file10000.Seek(0, 2)
			_, err = dataFiles.file10000.Write(make([]byte,10000*192))
			if err != nil {
				log.Fatal(err)
			}
		}
		dataFiles.file10000.Seek(recordData.pos10000 + offset10000, 0)
		dataFiles.file10000.Write(listTransactions[begin10000*192:end10000*192])
	}
	if (endFile != 0) {
		file, err := os.OpenFile(fmt.Sprintf("../Ethereum/%vU%d.db", dataFiles.name, userId), os.O_WRONLY|os.O_CREATE, 0666) //TODO filter userid leave only numbers!!!
		if err != nil {
			log.Fatal( err)
		}
		recordData.file = 4
		file.Seek(0,2)
		_, err = file.Write(listTransactions[beginFile*192:endFile*192]) //TODO write error handling
		if err != nil {
			log.Fatal(err)
		}
		file.Close()
	}
	recordData.records = recordData.records +  uint64(len(listTransactions)/192)
	return 1
}

/*
function readMapData
args: storage
function reads data from a storage and writes it to a binary array
*/
func readMapData(dataStorage *Storage) {
	var userId uint64
	var userSliceId int

	file, err := os.Open(fmt.Sprintf("../Ethereum/%vMap.db",dataStorage.name))

	defer file.Close()

	if err != nil {
		return
	}

	fi, err := file.Stat()
	if err != nil {
		log.Fatal(err) // Could not obtain stat, handle error
	}
	var i uint64

	for i = 0; i < uint64(fi.Size()/48); i++ {
		data := readNextBytesETH(file, 48)
		if (len(data) < 48) {
			fmt.Println(len(data))
			break
		}
		if (i == uint64(len(dataStorage.historyRegistry))) {
			fmt.Println(len(dataStorage.historyRegistry), uint64(fi.Size()/48))
			break
		}

		userId = binary.LittleEndian.Uint64(data[40:48])

		dataStorage.historyRegistry[userId].pos200 = int64(binary.LittleEndian.Uint64(data[0:8]))
		dataStorage.historyRegistry[userId].pos1000 = int64(binary.LittleEndian.Uint64(data[8:16]))
		dataStorage.historyRegistry[userId].pos10000 = int64(binary.LittleEndian.Uint64(data[16:24]))
		dataStorage.historyRegistry[userId].records = binary.LittleEndian.Uint64(data[24:32])
		dataStorage.historyRegistry[userId].file = uint8(binary.LittleEndian.Uint64(data[32:40]))

		var new_read_offset uint64
		var read_strings uint64

		if dataStorage.historyRegistry[userId].records >= 200 {
			read_strings = 200
		} else {
			read_strings = dataStorage.historyRegistry[userId].records
		}

		if dataStorage.name == "transactionsHistoryETH" {
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

			userSliceId = userTransactionMap[userId]

			if (userSliceId == 0) {
				//set len slice to array
				userSliceId = len(userOrdersSlice)
				userTransactionMap[userId] = userSliceId

				TMPTransactionData.data = append(TMPTransactionData.data, transactions{uint16(userId),[]byte{}})
				userOrdersSlice = append(userOrdersSlice, transactions{uint16(userId), []byte{}})
			}
			userOrdersSlice[userSliceId].transactionsHistory = readFileDataETH(userId,read_strings, new_read_offset, dataStorage)
		}
	}
}

/*
function readFileDataOffsetETH
args: User ID, how many records will be read from storage, read offset, storage
function returns read offset
*/
func readFileDataOffsetETH(userId uint64, amount uint64, offset uint64, dataFiles *Storage) (uint64, uint64, uint64, uint64, uint64, uint64, uint64, uint64) {
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

	if (begin < 200) {
		readOffset200 = begin
		if (begin + amount < 200) {
			read200 = amount
			return readOffset200, read200, readOffset1000, read1000, readOffset10000, read10000, readOffsetFile, readFile
		} else {
			read200 = 200 - begin
			readOffset1000 = 0
		}
	}

	if (begin < 1200) {
		if(begin > 200) {
			readOffset1000 = begin - 200
		}

		if (begin + amount < 1200) {
			read1000 = amount-read200
			return readOffset200, read200, readOffset1000, read1000, readOffset10000, read10000, readOffsetFile, readFile
		} else {
			read1000 = 1000 - readOffset1000
		}
	}

	if (begin < 11200) {
		if (begin > 1200) {
			readOffset10000 = begin - 1200
		}

		if (begin + amount < 11200) {
			read10000 = amount-read200-read1000
			return readOffset200, read200, readOffset1000, read1000, readOffset10000, read10000, readOffsetFile, readFile
		} else {
			read10000 = 10000 - readOffset10000
		}
	}

	if (begin > 11200) {
		readOffsetFile = offset //seek from the end to amount
		readFile = amount
	} else {
		readOffsetFile = recordData.records - 11200
		readFile = amount - read200 - read1000  - read10000
	}
	return readOffset200, read200, readOffset1000, read1000, readOffset10000, read10000, readOffsetFile, readFile
}

/*
function readFileDataETH
args: User ID,  how many records will be read from storage, read offset, storage
function returns records from storage
*/
func readFileDataETH(userId uint64, amount uint64, offset uint64, dataFiles *Storage) ([]byte) {
	buffer := make([]byte,1)
	recordData := &dataFiles.historyRegistry[userId]

	if (recordData.records < offset || recordData.records == 0) {
		return []byte{0}
	}

	ro200, r200, ro1000, r1000, ro10000, r10000, roFile, rFile := readFileDataOffsetETH(userId, amount, offset, dataFiles)

	if (r200 > 0){
		dataFiles.file200.Seek(recordData.pos200+int64(ro200)*192,0)
		data := readNextBytesETH(dataFiles.file200, 192*int(r200))

		if (len(buffer) == 1) {
			buffer = data
		} else {
			buffer = append(buffer, data...)
		}
	}

	if (r1000 > 0) {
		dataFiles.file1000.Seek(recordData.pos1000+int64(ro1000)*192,0)
		data := readNextBytesETH(dataFiles.file1000, 192*int(r1000))
		if (len(buffer) == 1) {
			buffer = data
		} else {
			buffer = append(buffer, data...)
		}
	}

	if (r10000 > 0) {
		dataFiles.file10000.Seek(recordData.pos10000+int64(ro10000)*192,0)
		data := readNextBytesETH(dataFiles.file10000, 192*int(r10000))
		if (len(buffer) == 1) {
			buffer = data
		} else {
			buffer = append(buffer, data...)
		}
	}

	if (rFile > 0) {
		fileX, err := os.Open(fmt.Sprintf("../Ethereum/%vU%d.db", dataFiles.name, userId))

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

		data := readNextBytesETH(fileX, 192*int(rFile))
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
function sendTrhistoryETH
sending transaction history to the wallets server
*/
func sendTrhistoryETH() {
	for{
		if len(sendBufSlice.data) >= 198 {
			sendBufSlice.mux.Lock()
			sndTransactionHistory.SendBytes(sendBufSlice.data, 0)
			sendBufSlice.data  = []byte{}
			sendBufSlice.mux.Unlock()
		}
		time.Sleep(10*time.Second)
	}
}

/*
function saveTransactionsETH
saving users' transactions
*/
func saveTransactionsETH() {
	for {
		filename := "../Ethereum/transactions.db"
		PendingTransactions.mux.Lock()
		write_transaction_toETH(filename, PendingTransactions.data)
		PendingTransactions.mux.Unlock()

		filename = "../Ethereum/transactionsOUT.db"
		PendingTransactionsOUT.mux.Lock()
		write_transaction_toETH(filename, PendingTransactionsOUT.data)
		PendingTransactionsOUT.mux.Unlock()

		ConfirmedTransactions.mux.Lock()
		filename = "../Ethereum/ConfirmedTransactions.db"
		write_transaction_toETH(filename, ConfirmedTransactions.data)
		ConfirmedTransactions.mux.Unlock()

		ConfirmedTransactionsOUT.mux.Lock()
		filename = "../Ethereum/ConfirmedTransactionsOUT.db"
		write_transaction_toETH(filename, ConfirmedTransactionsOUT.data)
		ConfirmedTransactionsOUT.mux.Unlock()

		time.Sleep(10*time.Second)
	}
}
func writeAllDataETHLoop() {
	for {
		if len(TMPTransactionData.data) >= 1 {
			time.Sleep(1*time.Second)
			writeAllDataETH()
		}
		time.Sleep(1*time.Minute)
	}
}

func main() {
	// make maps with without nil data
	usersTrunsactionsT.data = make(map[string] UserTrunsaction)
	usersTrunsactionsOUT.data = make(map[string] UserTrunsaction)
	ConfirmedTransactions.data = make(map[string]TransactionsTmp)
	ConfirmedTransactionsOUT.data = make(map[string]TransactionsTmp)
	PendingTransactions.data = make(map[string]TransactionsTmp)
	PendingTransactionsOUT.data = make(map[string]TransactionsTmp)
	PendingAdderess.data = make(map[uint64]string)
	UsersHasAddr.data = make(map[uint64]HasAddress)
	TMPTransactionData.data=  make([]transactions, 1, 1)
	//-----

	//Creating connection with Balance server (send)
	//add founds
	sndBalance, _ = zmq.NewSocket(zmq.PUSH)
	defer sndBalance.Close()
	sndBalance.SetRcvhwm(1100000)
	sndBalanceAddress := fmt.Sprintf("%v%v", LocalConfig.LocalTcpIpAddr, LocalConfig.ETHBalancePort)
	err := sndBalance.Bind(sndBalanceAddress)
	checkError(err)

	//Creating connection with Wallets  server (receive)
	//receive new wallet address
	rcvNewUserAddr, _ = zmq.NewSocket(zmq.PULL)
	defer rcvNewUserAddr.Close()
	rcvNewUserAddr.SetRcvhwm(1100000)
	rcvAddress := fmt.Sprintf("%v%v", LocalConfig.LocalTcpIpAddr, LocalConfig.ETHrcvNewAddressPort)
	rcvNewUserAddr.Connect(rcvAddress)

	//Creating connection with Wallets  server (send)
	//send history
	sndTransactionHistory, _ = zmq.NewSocket(zmq.PUSH)
	defer sndTransactionHistory.Close()
	sndTransactionHistoryAddress := fmt.Sprintf("%v%v", LocalConfig.LocalTcpIpAddr, LocalConfig.ETHTrnsctnHistoryPort)
	err = sndTransactionHistory.Bind(sndTransactionHistoryAddress)
	checkError(err)

	//Loading data from Ethereum Transactions Storage
	TransactionsHistory.file200, err = os.OpenFile("../Ethereum/transactionsHistory200.db", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer TransactionsHistory.file200.Close()

	TransactionsHistory.file1000, err = os.OpenFile("../Ethereum/transactionsHistory1000.db", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer TransactionsHistory.file1000.Close()

	TransactionsHistory.file10000, err = os.OpenFile("../Ethereum/transactionsHistory10000.db", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer TransactionsHistory.file10000.Close()
	TransactionsHistory.name = "transactionsHistoryETH"

	//Read data from Ethereum Transactions Storage
	readMapData(&TransactionsHistory)

	loadUsersWalletsETH()
	LoadConfirmedTransactionsETH()
	LoadConfirmedTransactionsOUTETH()
	load_TransactionsETH()
	load_TransactionsOUTETH()

	go change_trnsctn_statusETH()
	go change_trnsctn_statusOUTETH()
	go rcvnewAddr()
	go sendTrhistoryETH()
	go saveTransactionsETH()
	go writeAllDataETHLoop()

	for {
		lastProcessedBlock := loadLastBlockETH() //load from file // load last block
		latestBlock := getBlockETH("latest")

		var lastBlockNumber uint64

		if len(latestBlock) == 0 {
			time.Sleep(1*time.Second)
			continue
		}

		if len(lastProcessedBlock) == 0 {
			lastBlockNumber = 0
		} else {
			lastBlockNumber, err = strconv.ParseUint(lastProcessedBlock[2:len(lastProcessedBlock)], 16, 64)
		}

		checkError(err)
		latestBlockNumber, err := strconv.ParseUint(latestBlock[2:len(latestBlock)], 16, 64)
		checkError(err)
		if (latestBlockNumber > lastBlockNumber) {
			saveLastBlockETH(latestBlock)
			for i:= latestBlockNumber; i > lastBlockNumber; i-- {
				getBlockETH(toHex(i))
				if i == lastBlockNumber {
					break
				}
			}
		}

		buf := generateBalanceBufETH()
		if len(buf) > 10 {
			sndBalance.SendBytes(buf,0)
		}
		// 1 request in 1 min
		fmt.Println("------ SLEEP 1 MIN ------")
		time.Sleep(1*time.Minute)
	}
	return
	os.Exit(0)
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func loadUsersWalletsETH() {
	var coinType string
	coinType = "ETH"
	var Path string
	var fileLen int64
	var PrivateKeyLen int64
	var PublicKeyLen int64
	if coinType == "BTC" {
		Path = "../users/usersBTC.db"
		fileLen = 102
		PrivateKeyLen = 60
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
		// Failed to get stat, handle error
		log.Fatal("Failed to get stat: ", err)
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
		data := readNextBytesETH(file, (int(fileLen)))
		var PrivateKey string
		var PublicKey string
		//
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

		TMPstructData.CoinID = CoinID
		TMPstructData.PublicKey = PublicKey

		if len(UsersHasAddr.data[UserId]) > 0 {
			for key, prevData := range UsersHasAddr.data[UserId] {
				keysData[key] = prevData
			}
			keysData[PrivateKey] = TMPstructData
		} else {
			keysData[PrivateKey] = TMPstructData
		}
		UsersHasAddr.data[UserId] = keysData

		if CoinID == 4 {
			PendingAdderess.mux.Lock()
			PendingAdderess.data[UserId] = strings.ToLower(string(PublicKey))
			PendingAdderess.mux.Unlock()
		}
	}
}