package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/ltcsuite/ltcd/btcec"
	"github.com/ltcsuite/ltcd/btcjson"
	"github.com/ltcsuite/ltcd/chaincfg"
	"github.com/ltcsuite/ltcd/chaincfg/chainhash"
	"github.com/ltcsuite/ltcd/rpcclient"
	"github.com/ltcsuite/ltcutil"
	zmq "github.com/pebbe/zmq4"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"
	"wallets/walletsP/LocalConfig"
)

var fileOL *os.File
var file200 *os.File
var file1000 *os.File
var file10000 *os.File

var sndTransactionHistory *zmq.Socket
var sndColdTxInOUT *zmq.Socket
var rcvOutTransaction *zmq.Socket
var sndTransactionHistoryOUT *zmq.Socket
var rcvNewUserAddr *zmq.Socket
var sndBalance *zmq.Socket

var TMPTransactionData struct {
	data []transactions
	mux sync.Mutex
}

var userOrdersSlice  = make([]transactions, 1, 1)
var balanceSlice = make([]priceUpdate, 1, 1)

type transactions struct {
	id uint16
	transactionsHistory []byte
}

var writeData struct {
	data map[string]*btcjson.TxRawResult
	mux sync.Mutex
}

var usersTrunsactionsT struct {
	data map[string]UserTrunsaction
	mux sync.Mutex
}

var ConfirmedTransactions struct {
	data map[string] *btcjson.TxRawResult
	mux sync.Mutex
}

var ConfirmedTransactionsOUT struct {
	data map[string] *btcjson.TxRawResult
	mux sync.Mutex
}

var PendingTransactions struct {
	data map[string] *btcjson.TxRawResult
	mux sync.Mutex
}

var PendingTransactionsOUT struct {
	data map[string] *btcjson.TxRawResult
	mux sync.Mutex
}

var sendBufSlice struct {
	data []byte
	mux sync.Mutex
}

var sendBufSliceOUT struct {
	data []byte
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

var  usersTrunsactionsOUT struct {
	data map[string] UserTrunsaction
	mux sync.Mutex
}

var ColdWalletTrunsactionsIN struct {
	data map[string] UserTrunsaction
	mux sync.Mutex
}

var CustomNetwork = &chaincfg.TestNet4Params
var TransactionsHistory Storage
var ColdWallet string
var StressTestFlag bool
var userTransactionMap [1000000] int

type Storage struct {
	file200 *os.File
	file1000 *os.File
	file10000 *os.File
	name string
	historyRegistry [1000000] Records
}

type Records struct {
	pos200 int64
	pos1000 int64
	pos10000 int64
	records uint64
	file uint8
}

type UserTrunsaction struct {
	UserId uint64
	TransactionHash string
	TrTime int64
	InOut int
	ConfirmedBlocks uint64
	TrFrom string
	TrTo string
	TrValue string
	CoinID uint64
	Vout uint64

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

type TransactionData struct {
	Txid          string  `json:"txid"` //+
	Blockhash     string  `json:"blockhash"`
	Blockheight   int     `json:"Blockheight"`
	Confirmations int     `json:"confirmations"`
	Time          int     `json:"time"`
	ValueOut      float64 `json:"valueOut"`
	ValueIn       float64 `json:"valueIn"`
	Fees          float64 `json:"fees"`
	Version  int    `json:"version"` //+
	Size     int    `json:"size"` //+
	Locktime int    `json:"locktime"`  //+
	Vin      []struct { //+
		Txid     string `json:"txid"`
		Vout     float64    `json:"vout"`
		Sequence int    `json:"sequence"`
		N        int    `json:"n"`
		ScriptSig struct {
			Hex           string  `json:"hex"`
			Asm           string  `json:"asm"`
		} `json:"scriptSig"`
		Value         int `json:"value"`
		LegacyAddress string  `json:"legacyAddress"`
		CashAddress   string  `json:"cashAddress"`
	} `json:"vin"`
	Vout []struct {
		Value string `json:"value"`
		N     int     `json:"n"`
		ScriptPubKey struct {
			Asm       string   `json:"asm"`
			Hex       string   `json:"hex"`
			Addresses []string `json:"addresses"`
			Type      string   `json:"type"`
			CashAddrs []string `json:"cashAddrs"`
		} `json:"scriptPubKey"`
		SpentTxId   string      `json:"spentTxId"`
		SpentIndex  int      `json:"spentIndex"`
		SpentHeight int      `json:"spentHeight"`
	} `json:"vout"`
}

type Block struct {
	Hash          string `json:"hash"`
	Confirmations int    `json:"confirmations"`
	Strippedsize  int    `json:"strippedsize"`
	Size          int    `json:"size"`
	Weight        int    `json:"weight"`
	Height        int    `json:"height"`
	Version       int    `json:"version"`
	VersionHex    string `json:"versionHex"`
	Merkleroot    string `json:"merkleroot"`
	Tx            []struct {
		Txid     string `json:"txid"`
		Hash     string `json:"hash"`
		Version  int    `json:"version"`
		Size     int    `json:"size"`
		Vsize    int    `json:"vsize"`
		Weight   int    `json:"weight"`
		Locktime int    `json:"locktime"`
		Vin      []struct {
			Coinbase string `json:"coinbase"`
			Sequence int64  `json:"sequence"`
		} `json:"vin"`
		Vout []struct {
			Value        float64 `json:"value"`
			N            int `json:"n"`
			ScriptPubKey struct {
				Asm       string   `json:"asm"`
				Hex       string   `json:"hex"`
				ReqSigs   int      `json:"reqSigs"`
				Type      string   `json:"type"`
				Addresses []string `json:"addresses"`
			} `json:"scriptPubKey"`
		} `json:"vout"`
		Hex string `json:"hex"`
	} `json:"tx"`
	Time              int     `json:"time"`
	Mediantime        int     `json:"mediantime"`
	Nonce             int     `json:"nonce"`
	Bits              string  `json:"bits"`
	Difficulty        float64 `json:"difficulty"`
	Chainwork         string  `json:"chainwork"`
	NTx               int     `json:"nTx"`
	Previousblockhash string  `json:"previousblockhash"`
}

type BestBlock struct {
	Result string      `json:"result"`
	Error  interface{} `json:"error"`
	ID     interface{} `json:"id"`
}

type Network struct {
	name        string
	symbol      string
	xpubkey     byte
	xprivatekey byte
}

var network = map[string]Network {
	"rdd": {name: "reddcoin", symbol: "rdd", xpubkey: 0x3d, xprivatekey: 0xbd},
	"dgb": {name: "digibyte", symbol: "dgb", xpubkey: 0x1e, xprivatekey: 0x80},
	"btc": {name: "bitcoin",  symbol: "btc", xpubkey: 0x00, xprivatekey: 0x80},
	"ltc": {name: "litecoin", symbol: "ltc", xpubkey: 0x30, xprivatekey: 0xb0},
}

func (network Network) GetNetworkParams() *chaincfg.Params {
	networkParams := &chaincfg.MainNetParams
	networkParams.PubKeyHashAddrID = network.xpubkey
	networkParams.PrivateKeyID = network.xprivatekey
	return networkParams
}

func (network Network) CreatePrivateKey() (*ltcutil.WIF, error) {
	secret, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return nil, err
	}
	return ltcutil.NewWIF(secret, network.GetNetworkParams(), true)
}


func (network Network) ImportWIF(wifStr string) (*ltcutil.WIF, error) {
	wif, err := ltcutil.DecodeWIF(wifStr)
	if err != nil {
		return nil, err
	}
	if !wif.IsForNet(network.GetNetworkParams()) {
		return nil, errors.New("The WIF string is not valid for the `" + network.name + "` network")
	}
	return wif, nil
}

func (network Network) GetAddress(wif *ltcutil.WIF) (*ltcutil.AddressPubKey, error) {
	return ltcutil.NewAddressPubKey(wif.PrivKey.PubKey().SerializeCompressed(), network.GetNetworkParams())
}

/*
function addFunds
searching for pending address in the user address list
call to addToBalance function
*/
func addFunds(address []string, funds float64) {
	PendingAdderess.mux.Lock()
	for _, checkAddress := range address {
		for userID, P_address := range PendingAdderess.data {
			if checkAddress == P_address {
				addToBalance(3, uint32(userID), uint64(funds*1000000000)) //0,0001
			}
		}
	}
	PendingAdderess.mux.Unlock()
}

/*
function loadLastBlock
load the last processed block from the database
*/
func loadLastBlock() string {
	b, err := ioutil.ReadFile("../Litecoin/lastBTCBlock.log")
	if err != nil {
		fmt.Println(err)
		return ""
	}

	return string(b)
}

/*
function saveLastBlock
saving the last processed block in the database
*/
func saveLastBlock(block string) {
	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile("../Litecoin/lastBTCBlock.log", os.O_CREATE|os.O_WRONLY, 0644)
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
function addToBalance
add data to slice
*/
func addToBalance(coin uint16, userId uint32, add uint64) {
	balanceSlice = append(balanceSlice, priceUpdate{coin:coin, userId:userId, add:add})
}

/*
function generateBalanceBuf
Convert data to binary format
*/
func generateBalanceBuf() []byte {
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
function readNextBytes
function returns N bytes from a file
*/
func readNextBytes(file *os.File, number int) []byte {
	bytes := make([]byte, number)

	size, err := file.Read(bytes)
	if err != nil {
		log.Fatal(err, size)
	}

	if (size != number){
		return []byte{}
	}

	return bytes
}

/*
function get_last_block
get the last block (block hash) from the Litecoin Core
*/
func get_last_block(client *rpcclient.Client) string {
	blockChainInfo, err := client.GetBlockChainInfo()
	if err != nil {
		log.Fatal(err)
	}
	return blockChainInfo.BestBlockHash
}

/*
function getblockHeight
get the block number
*/
func getblockHeight(client *rpcclient.Client, hash string) int64 {
	blockHash, err := chainhash.NewHashFromStr(hash)
	if err != nil {
		log.Fatal(err)
	}
	blockInfo, err := client.GetBlockVerbose(blockHash)
	if err != nil {
		log.Fatal(err)
	}
	return blockInfo.Height
}

/*
function get_block
get the block hash by the Int number
*/
func get_block(client *rpcclient.Client, i int64) string {
	blockHash, err := client.GetBlockHash(i)
	if err != nil {
		log.Fatal(err)
	}
	return blockHash.String()
}

/*
function getTransacionData
get the transaction info by hash
*/
func getTransacionData(client *rpcclient.Client, TxArr []string) {
	for _, TxData := range TxArr {
		blockHash, err := chainhash.NewHashFromStr(TxData) //convert string to chainhash
		Tx, err := client.GetRawTransactionVerbose(blockHash) // get transaction data
		if err != nil {
			continue //TODO ????
			log.Fatal(err)
		}
		//check transactions for existing in our system
		_, ok := PendingTransactions.data[Tx.Txid];
		if ok {
			continue
		}
		//проверяем добавляли ли мы баланс на акк. пользователю
		_, ok2 := ConfirmedTransactions.data[Tx.Txid];
		if ok2 {
			continue
		}
		go doLoopIn(Tx, client, 0)
	}
}

/*
send test data to the Litecoin OUT server
*/
func sendPing() {
	for {
		buf := make([]byte, 8)
		binary.LittleEndian.PutUint64(buf[0:8], 3)
		sndColdTxInOUT.SendBytes(buf,0)
		sndTransactionHistoryOUT.SendBytes(buf,0)
		time.Sleep(1 * time.Minute)
	}
}

/*
send Tx ID to the Litecoin OUT server
*/
func sendColdWalletData(Tx *btcjson.TxRawResult,  Vout uint64) {
	bufOUT := make([]byte, 80)
	binary.LittleEndian.PutUint64(bufOUT[0:8], 3)
	binary.LittleEndian.PutUint64(bufOUT[8:16], Vout)
	var txIn string
	txIn = fmt.Sprintf("%v", Tx.Txid)
	copy(bufOUT[16:80], []byte(txIn))
	sndColdTxInOUT.SendBytes(bufOUT,0)
	time.Sleep(1 * time.Second)
}

/*
function doLoopIn
check block transactions
*/
func doLoopIn(Tx *btcjson.TxRawResult, client *rpcclient.Client, isLoad int) {
	var ij uint64
	ij = 0
	for _, Tx2_ := range Tx.Vout { // Vout loop
		if len(Tx2_.ScriptPubKey.Addresses) >= 1 {
			UsersHasAddr.mux.Lock()
			//verification of customer address matches
			for userID, userData := range UsersHasAddr.data {
				for _, keysData := range userData {
					for _, checkAddress := range Tx2_.ScriptPubKey.Addresses {
						if string(checkAddress) == string(keysData.PublicKey) {
							PendingAdderess.mux.Lock()
							PendingAdderess.data[userID] = checkAddress //add an address to the Pending map (the user that will receive coins)
							PendingAdderess.mux.Unlock()
							save_Transaction(Tx, checkAddress, Tx2_.Value, userID, ij, isLoad)
						}
					}
				}
			}
			UsersHasAddr.mux.Unlock()
			//check for matching cold wallets
			for _, checkAddress := range Tx2_.ScriptPubKey.Addresses {
				if ColdWallet == checkAddress {
					sendColdWalletData(Tx, ij)
					PendingAdderess.mux.Lock()
					PendingAdderess.data[1] = checkAddress
					PendingAdderess.mux.Unlock()

					save_Transaction(Tx, checkAddress, Tx2_.Value, 1, ij, 0)
				}
			}
		}
		ij++
	}
}

/*
function save_Transaction
saving the user transaction in the database
sending Tx data to the wallets server
*/
func save_Transaction(Tx *btcjson.TxRawResult, TrTo string, value float64, userID uint64, Vout uint64, isLoad int) {
	PendingTransactions.mux.Lock()
	PendingTransactions.data[Tx.Txid] = Tx
	PendingTransactions.mux.Unlock()

	usersTrunsactionsT.mux.Lock()
	var tmpV uint64
	tmpV = uint64(value*1000000000)
	TrValue := fmt.Sprintf("%d", int(tmpV))

	if TrTo == ColdWallet {
		ColdWalletTrunsactionsIN.mux.Lock()
		ColdWalletTrunsactionsIN.data[Tx.Hash] = UserTrunsaction{UserId: userID,
			TransactionHash: Tx.Txid,
			InOut: 1, TrTime: Tx.Time,
			ConfirmedBlocks: Tx.Confirmations,
			TrTo: TrTo,
			TrValue: TrValue,
			TrFrom:"",
			CoinID: 3,
			Vout: Vout}
		ColdWalletTrunsactionsIN.mux.Unlock()
	} else {
		usersTrunsactionsT.data[Tx.Hash] = UserTrunsaction{UserId: userID,
			TransactionHash: Tx.Txid,
			InOut: 1,
			TrTime: Tx.Time,
			ConfirmedBlocks: Tx.Confirmations,
			TrTo: TrTo,
			TrValue: TrValue,
			TrFrom:"",
			CoinID: 3,
			Vout: Vout}
	}
	usersTrunsactionsT.mux.Unlock()
	if TrTo != ColdWallet {
		if isLoad == 0 {
			fmt.Println("ssssssssss")
			generateTrHbuff(usersTrunsactionsT.data[Tx.Hash], 1)
		}
	}
}

/*
function save_TransactionOUT
function for the transaction that was sent from our servers
saving the user transaction in the database
sending Tx data to the wallets server
*/
func save_TransactionOUT(Tx *btcjson.TxRawResult, TrTo string, trFrom string, value uint64, userID uint64, isLoad int) {
	PendingTransactionsOUT.mux.Lock()
	PendingTransactionsOUT.data[Tx.Txid] = Tx
	PendingTransactionsOUT.mux.Unlock()

	usersTrunsactionsOUT.mux.Lock()
	TrValue := fmt.Sprintf("%v", value)
	usersTrunsactionsOUT.data[Tx.Hash] = UserTrunsaction{UserId: userID,
		TransactionHash: Tx.Txid,
		InOut: 0,
		TrTime: Tx.Time,
		ConfirmedBlocks: Tx.Confirmations,
		TrTo: TrTo,
		TrValue: TrValue,
		TrFrom:trFrom,
		CoinID: 3}
	usersTrunsactionsOUT.mux.Unlock()

	if isLoad == 0 {
		generateTrHbuff(usersTrunsactionsOUT.data[Tx.Hash], 0)
	}
}

/*
function change_trnsctn_status
the loop that will check the number of confirmations
*/
func change_trnsctn_status(client *rpcclient.Client) {
	for {
		//check all OUR users transactions
		PendingTransactions.mux.Lock()
		PendingTransactionsTMP := PendingTransactions.data
		PendingTransactions.mux.Unlock()

		for _, TranXon := range PendingTransactionsTMP {
			ConfirmedTransactions.mux.Lock()
			_, ok2 := ConfirmedTransactions.data[TranXon.Txid];
			ConfirmedTransactions.mux.Unlock()
			if ok2 {
				delete(PendingTransactions.data, TranXon.Txid)
				continue
			}
			var status bool
			var destination uint64
			//checking number of confirmations
			if StressTestFlag == false {
				status,destination = check_transuction(client, TranXon.BlockHash, TranXon)
			} else {
				status, destination = StressTransuction(TranXon)
			}
			//confirmations == success
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

				TMPTransactionLIST := &TMPTransactionData.data[*userSliceId] //work with the user storage list
				TMPOrdersLIST := &userOrdersSlice[*userSliceId]

				SaveConfirmedTransactions(TranXon)

				//Change data stutus
				usersTrunsactionsT.mux.Lock()
				usersTrunsactionsT.data[TranXon.Hash] = UserTrunsaction{UserId: usersTrunsactionsT.data[TranXon.Hash].UserId,
					TransactionHash: usersTrunsactionsT.data[TranXon.Hash].TransactionHash,
					InOut: 1,
					TrTime: usersTrunsactionsT.data[TranXon.Hash].TrTime,
					ConfirmedBlocks: 6,
					TrTo: usersTrunsactionsT.data[TranXon.Hash].TrTo,
					TrValue: usersTrunsactionsT.data[TranXon.Hash].TrValue,
					TrFrom:usersTrunsactionsT.data[TranXon.Hash].TrFrom,
					CoinID:3}
				usersTrunsactionsT.mux.Unlock()

				//add founds to user
				for _, Vout := range TranXon.Vout {
					addFunds(Vout.ScriptPubKey.Addresses, Vout.Value)
				}
				if usersTrunsactionsT.data[TranXon.Hash].UserId != 0 && len(usersTrunsactionsT.data[TranXon.Hash].TransactionHash) > 10 {
					//write comfirmed transactions
					buf := convertTransactionsToBin(usersTrunsactionsT.data[TranXon.Hash])
					TMPTransactionLIST.transactionsHistory = append(TMPTransactionLIST.transactionsHistory, buf...)
					TMPOrdersLIST.transactionsHistory = append(TMPTransactionLIST.transactionsHistory, buf...)
					//sending info to the Wallets server
					generateTrHbuff(usersTrunsactionsT.data[TranXon.Hash], 0)
				}
				delete(PendingTransactions.data, TranXon.Txid)
			} else {
				//changing number of confirmations
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
						TrFrom:"",
						CoinID:3}
					usersTrunsactionsT.mux.Unlock()
					generateTrHbuff(usersTrunsactionsT.data[TranXon.Hash], 0)
				}
			}
		}
		time.Sleep(30 * time.Second)
	}
}

/*
function change_trnsctn_statusOUT
the same function like change_trnsctn_status but for transactions that was sent from our servers
*/
func change_trnsctn_statusOUT(client *rpcclient.Client) {
	for {
		PendingTransactionsOUT.mux.Lock()
		PendingTransactionsOUTTMP := PendingTransactionsOUT.data
		PendingTransactionsOUT.mux.Unlock()

		for _, TranXon := range PendingTransactionsOUTTMP {
			ConfirmedTransactionsOUT.mux.Lock()
			_, ok2 := ConfirmedTransactionsOUT.data[TranXon.Txid];
			ConfirmedTransactionsOUT.mux.Unlock()
			if ok2 {
				delete(PendingTransactionsOUT.data, TranXon.Txid)
				continue
			}
			var status bool
			var destination uint64

			if StressTestFlag == false {
				status,destination = check_transuction(client, TranXon.BlockHash, TranXon)
			} else {
				status, destination = StressTransuction(TranXon)
			}

			if status == true {
				if usersTrunsactionsOUT.data[TranXon.Hash].UserId == 0 {
					delete(PendingTransactionsOUT.data, TranXon.Txid)
					continue
				}
				SaveConfirmedTransactionsOUT(TranXon)

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

				usersTrunsactionsOUT.mux.Lock()
				usersTrunsactionsOUT.data[TranXon.Hash] = UserTrunsaction{UserId: usersTrunsactionsOUT.data[TranXon.Hash].UserId,
					TransactionHash: usersTrunsactionsOUT.data[TranXon.Hash].TransactionHash,
					InOut: 0,
					TrTime: usersTrunsactionsOUT.data[TranXon.Hash].TrTime,
					ConfirmedBlocks: 6,
					TrTo: usersTrunsactionsOUT.data[TranXon.Hash].TrTo,
					TrValue: usersTrunsactionsOUT.data[TranXon.Hash].TrValue,
					TrFrom: usersTrunsactionsOUT.data[TranXon.Hash].TrFrom,
					CoinID:3}
				usersTrunsactionsOUT.mux.Unlock()

				if usersTrunsactionsOUT.data[TranXon.Hash].UserId > 2 && len(usersTrunsactionsOUT.data[TranXon.Hash].TransactionHash) > 10 {
					//write comfirmed transactions
					var bufOUT []byte
					bufOUT = convertTransactionsToBin(usersTrunsactionsOUT.data[TranXon.Hash])
					TMPTransactionLIST.transactionsHistory = append(TMPTransactionLIST.transactionsHistory, bufOUT...)
					TMPOrdersLIST.transactionsHistory = append(TMPTransactionLIST.transactionsHistory, bufOUT...)
					generateTrHbuff(usersTrunsactionsOUT.data[TranXon.Hash], 0)
				}

				delete(PendingTransactionsOUT.data, TranXon.Txid)
			} else {
				dstntn := destination+1
				if (dstntn) > usersTrunsactionsOUT.data[TranXon.Hash].ConfirmedBlocks {
					usersTrunsactionsOUT.mux.Lock()
					usersTrunsactionsOUT.data[TranXon.Hash] = UserTrunsaction{UserId: usersTrunsactionsOUT.data[TranXon.Hash].UserId,
						TransactionHash: usersTrunsactionsOUT.data[TranXon.Hash].TransactionHash,
						InOut: 0,
						TrTime: usersTrunsactionsOUT.data[TranXon.Hash].TrTime,
						ConfirmedBlocks: dstntn,
						TrTo: usersTrunsactionsOUT.data[TranXon.Hash].TrTo,
						TrValue: usersTrunsactionsOUT.data[TranXon.Hash].TrValue,
						TrFrom: usersTrunsactionsOUT.data[TranXon.Hash].TrFrom,
						CoinID: 3}
					usersTrunsactionsOUT.mux.Unlock()
					generateTrHbuff(usersTrunsactionsOUT.data[TranXon.Hash], 0)
				}
			}
		}
		time.Sleep(30 * time.Second)
	}
}

/*
function transactionForCheck
get transaction data
*/
func transactionForCheck(client *rpcclient.Client, TxArr []string) []*btcjson.TxRawResult {
	var transaction  []*btcjson.TxRawResult
	for _, TxData := range TxArr {
		blockHash, err := chainhash.NewHashFromStr(TxData)
		if err != nil {
			continue
		}
		checkError(err)
		Tx, err := client.GetRawTransactionVerbose(blockHash)
		if err != nil {
			continue
		}
		checkError(err)
		transaction = append(transaction, Tx)
	}
	return transaction
}

/*
function check_transuction
the function that searches broken transactions and counts confirmations
*/
func check_transuction(client *rpcclient.Client, CheckHash string, TranXon *btcjson.TxRawResult) (bool, uint64) {
	//check transaction for existing
	blockHashForCheck, err := chainhash.NewHashFromStr(CheckHash)
	if err != nil {
		return false, 1
	}
	checkError(err)
	//check transaction for existing
	blockInfoForCheck, err := client.GetBlockVerbose(blockHashForCheck)
	if err != nil {
		return false, 1
	}
	checkError(err)
	BlockHeightForCheck := blockInfoForCheck.Height
	transactionsForCheck := transactionForCheck(client,  blockInfoForCheck.Tx)

	Last_blockInfo, err := client.GetBlockChainInfo()
	if err != nil {
		return false, 1
	}
	checkError(err)

	Last_blockHashstr := Last_blockInfo.BestBlockHash
	Last_blockHash, err := chainhash.NewHashFromStr(Last_blockHashstr)
	if err != nil {
		return false, 1
	}
	checkError(err)
	Last_blockData, err := client.GetBlockVerbose(Last_blockHash)
	if err != nil {
		return false, 1
	}
	checkError(err)
	Current_BlockHeight := Last_blockData.Height
	var destination uint64
	destination = uint64(Current_BlockHeight - BlockHeightForCheck)
	//count confirmations
	if destination >= 5 { // 5 == 6 blocks
		ii := 0
		points_of_distinction := 0
		//seach transaction chenges
		for _, TraArr := range transactionsForCheck {
			if TraArr.Txid == TranXon.Txid {
				points_of_distinction++  //1
			} else {
				continue;
			}

			if TraArr.Hex == TranXon.Hex {
				points_of_distinction++ //2
			} else {
				continue;
			}
			if TraArr.Hash == TranXon.Hash {
				points_of_distinction++ //3
			} else {
				continue;
			}
			if TraArr.Size == TranXon.Size {
				points_of_distinction++ //4
			} else {
				continue;
			}
			if TraArr.Vsize == TranXon.Vsize {
				points_of_distinction++ //5
			} else {
				continue;
			}
			if TraArr.Weight == TranXon.Weight {
				points_of_distinction++ //6
			} else {
				continue;
			}
			if TraArr.Version == TranXon.Version {
				points_of_distinction++ //7
			} else {
				continue;
			}
			if TraArr.LockTime == TranXon.LockTime {
				points_of_distinction++ //8
			} else {
				continue;
			}
			if TraArr.BlockHash == TranXon.BlockHash {
				points_of_distinction++ //9
			} else {
				continue;
			}
			if TraArr.Time == TranXon.Time {
				points_of_distinction++ //10
			} else {
				continue;
			}
			jj :=0
			for _, Tx2_Vin := range TraArr.Vin {
				if Tx2_Vin.Coinbase != TranXon.Vin[jj].Coinbase{
					return false, destination
				}
				if Tx2_Vin.Txid != TranXon.Vin[jj].Txid{
					return false, destination
				}
				if Tx2_Vin.Vout != TranXon.Vin[jj].Vout{
					return false, destination
				}
				if Tx2_Vin.Sequence != TranXon.Vin[jj].Sequence{
					return false, destination
				}
				jj++
			}
			kk := 0
			for _, Tx2_Vout := range TraArr.Vout {
				if Tx2_Vout.Value != TranXon.Vout[kk].Value{
					return false, destination
				}
				if Tx2_Vout.N != TranXon.Vout[kk].N{
					return false, destination
				}
				if Tx2_Vout.ScriptPubKey.Asm != TranXon.Vout[kk].ScriptPubKey.Asm{
					return false, destination
				}
				if Tx2_Vout.ScriptPubKey.Hex != TranXon.Vout[kk].ScriptPubKey.Hex{
					return false, destination
				}
				if Tx2_Vout.ScriptPubKey.ReqSigs != TranXon.Vout[kk].ScriptPubKey.ReqSigs{
					return false, destination
				}
				if Tx2_Vout.ScriptPubKey.Type != TranXon.Vout[kk].ScriptPubKey.Type{
					return false, destination
				}
				ll := 0
				for _, Addresses := range Tx2_Vout.ScriptPubKey.Addresses {
					if Addresses != TranXon.Vout[kk].ScriptPubKey.Addresses[ll] {
						return false, destination
					}
					ll++
				}
				kk++
			}
			ii++
		}
		if points_of_distinction == (10*ii) {
			return true, destination
		} else {
			return false, destination
		}
	} else {
		return false, destination
	}
	return false, 0
}

/*
function write_transaction_to
saved transaction hash to the database
*/
func write_transaction_to(filename string, data map[string]*btcjson.TxRawResult, TrType string) {
	file, err := os.OpenFile(filename, os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("cmd.Run25() failed with %s\n", err)
	}
	defer file.Close()
	writeData.mux.Lock()
	writeData.data = data
	writeData.mux.Unlock()
	writeData.mux.Lock()
	for _, TranXon := range writeData.data {
		var b []byte
		var bufferedWriter *bufio.Writer
		if TrType == "transactionsOUT" {
			TrFrom := usersTrunsactionsOUT.data[TranXon.Txid].TrFrom
			UserId := usersTrunsactionsOUT.data[TranXon.Txid].UserId

			b = make([]byte, 8)
			binary.LittleEndian.PutUint64(b, UserId)
			bufferedWriter = bufio.NewWriter(file)
			_, err = bufferedWriter.Write(b,)
			bufferedWriter.Flush()

			b = make([]byte, 34)
			copy(b[:], []byte(TrFrom))
			bufferedWriter = bufio.NewWriter(file)
			_, err = bufferedWriter.Write( b, )
			checkError(err)
			bufferedWriter.Flush()
		}
		b = make([]byte, 64)
		copy(b[:], []byte(TranXon.Txid))
		bufferedWriter = bufio.NewWriter(file)
		_, err = bufferedWriter.Write( b, )
		checkError(err)
		bufferedWriter.Flush()

	}
	writeData.mux.Unlock()
	file.Close()
}

/*
function SaveConfirmedTransactions
saved Confirmed transaction hash to the database
*/
func SaveConfirmedTransactions(transaction *btcjson.TxRawResult) {
	ConfirmedTransactions.mux.Lock()
	ConfirmedTransactions.data[transaction.Txid] = transaction
	ConfirmedTransactions.mux.Unlock()
}

func SaveConfirmedTransactionsOUT(transaction *btcjson.TxRawResult) {
	ConfirmedTransactionsOUT.mux.Lock()
	ConfirmedTransactionsOUT.data[transaction.Txid] = transaction
	ConfirmedTransactionsOUT.mux.Unlock()
}

func process_blocks(client *rpcclient.Client, hash string) {
	blockHash, err := chainhash.NewHashFromStr(hash)
	if err != nil {
		log.Fatal(err)
	}

	blockInfo, err := client.GetBlockVerbose(blockHash)
	if err != nil {
		log.Fatal(err)
	}
	getTransacionData(client, blockInfo.Tx)
}

func getRowTransaction(client *rpcclient.Client, TXHash string)  *btcjson.TxRawResult {
	blockHash, _ := chainhash.NewHashFromStr(TXHash)
	TxData, _ := client.GetRawTransactionVerbose(blockHash)
	return TxData
}

func LoadConfirmedTransactions(client *rpcclient.Client) {
	filename := "../Litecoin/ConfirmedTransactions.db"
	load_trunsuction_to(filename, "Confirmed", client)
}
func LoadConfirmedTransactionsOUT(client *rpcclient.Client) {
	filename := "../Litecoin/ConfirmedTransactionsOUT.db"
	load_trunsuction_to(filename, "Confirmed", client)
}
func load_Transactions(client *rpcclient.Client) {
	filename := "../Litecoin/transactions.db"
	load_trunsuction_to(filename, "Pending", client)
}
func load_TransactionsOUT(client *rpcclient.Client) {
	filename := "../Litecoin/transactionsOUT.db"
	load_trunsuction_to(filename, "PendingOUT", client)
}

func load_trunsuction_to(filename string, tr_type string, client *rpcclient.Client) {
	file, err := os.Open(filename)
	checkError(err)
	defer file.Close()
	fi, err := file.Stat()
	checkError(err)
	var i uint64
	var transuctionHash string
	var lenD int64
	if tr_type == "PendingOUT" {
		lenD = 106
	} else {
		lenD = 64
	}
	for i = 0; i < uint64(fi.Size()/(lenD)); i++ {
		var data []byte
		if tr_type == "PendingOUT" {
			data = readNextBytes(file, int(lenD))
			userId := binary.LittleEndian.Uint64(data[0:8])
			From := string(data[8:42])
			transuctionHash = string(data[42:106])

			transuction := getRowTransaction(client, transuctionHash)
			if transuction == nil {
				continue
			}
			for _, Tx2_ := range transuction.Vout {
				if len(Tx2_.ScriptPubKey.Addresses) >= 1 {
					UsersHasAddr.mux.Lock()
					//verification of customer address matches
					for _, userData := range UsersHasAddr.data {
						for _, keysData := range userData {
							for _, checkAddress := range Tx2_.ScriptPubKey.Addresses {
								if string(checkAddress) == string(keysData.PublicKey) {
									Value := uint64(Tx2_.Value*1000000000)
									save_TransactionOUT(transuction, checkAddress, From, Value, userId, 1)
								}
							}
						}
					}
					UsersHasAddr.mux.Unlock()
				}
			}
		} else {
			data = readNextBytes(file, int(lenD))
		}

		transuctionHash = string(data[0:64])
		transuction := getRowTransaction(client, transuctionHash)
		if transuction == nil {
			continue
		}
		if tr_type == "Confirmed" {
			ConfirmedTransactions.data[transuction.Txid] = transuction
		}
		if tr_type == "ConfirmedOUT" {
			ConfirmedTransactionsOUT.data[transuction.Txid] = transuction
		}
		if tr_type == "Pending" {
			doLoopIn(transuction, client, 1)
		}
	}
	file.Close()
}

func loadUsersWallets() {
	var coinType string
	coinType = "BTC"
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
		fileLen = 120
		PrivateKeyLen = 78
		PublicKeyLen = 	120
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
		data := readNextBytes(file, (int(fileLen)))
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
		UsersHasAddr.mux.Lock()
		if len(UsersHasAddr.data[UserId]) > 0 {
			for key, prevData := range UsersHasAddr.data[UserId] {
				keysData[key] = prevData
			}
			keysData[PrivateKey] = TMPstructData
		} else {
			keysData[PrivateKey] = TMPstructData
		}
		UsersHasAddr.data[UserId] = keysData

		UsersHasAddr.mux.Unlock()

		if CoinID == 3 {
			PendingAdderess.mux.Lock()
			PendingAdderess.data[UserId] = PublicKey
			PendingAdderess.mux.Unlock()
		}
	}
}

/*
function generateTrHbuff
converting transaction info to a binary type and sending it to Wallets server
*/
func generateTrHbuff(transaction UserTrunsaction, flag int) {
	buf := make([]byte, 208)
	bufOUT := make([]byte, 80)
	if flag == 1 {
		binary.LittleEndian.PutUint64(bufOUT[0:8], transaction.UserId)
		binary.LittleEndian.PutUint64(bufOUT[8:16], uint64(transaction.Vout))
		copy(bufOUT[16:80], []byte(transaction.TransactionHash))
	}
	binary.LittleEndian.PutUint64(buf[0:8], transaction.UserId)
	binary.LittleEndian.PutUint64(buf[8:16], uint64(transaction.InOut))
	binary.LittleEndian.PutUint64(buf[16:24], transaction.ConfirmedBlocks)
	binary.LittleEndian.PutUint64(buf[24:32], uint64(transaction.TrTime))
	copy(buf[32:96], []byte(transaction.TransactionHash))
	copy(buf[96:138], []byte(transaction.TrFrom))
	copy(buf[138:180], []byte(transaction.TrTo))
	copy(buf[180:200], []byte(transaction.TrValue))
	binary.LittleEndian.PutUint64(buf[200:208], uint64(transaction.CoinID))
	if len(transaction.TransactionHash) > 10 && len(transaction.TrValue) > 2 {
		if len(buf) == 208 {
			if len(sendBufSlice.data) <= 1 {
				sendBufSlice.mux.Lock()
				sendBufSlice.data = buf
				sendBufSlice.mux.Unlock()
			} else {
				sendBufSlice.mux.Lock()
				sendBufSlice.data = append(sendBufSlice.data, buf...)
				sendBufSlice.mux.Unlock()
			}
			globalCounter++
		}
		if len(bufOUT) == 80 && flag == 1 {
			sendBufSliceOUT.mux.Lock()
			sendBufSliceOUT.data = append(sendBufSliceOUT.data, bufOUT...)
			sendBufSliceOUT.mux.Unlock()
		}
	}
}

/*
function addTrOUTHistory
receiving transaction hash from the litecoin OUT server
*/
func addTrOUTHistory(client *rpcclient.Client) {
	poller := zmq.NewPoller()
	poller.Add(rcvOutTransaction, zmq.POLLIN)
	for {
		_, err := poller.Poll(-1)
		if err !=nil {
			fmt.Println(err)
		}
		msg,_ := rcvOutTransaction.RecvBytes(0)
		if len(msg) != 148 {
			continue
		}
		userId := binary.LittleEndian.Uint64(msg[0:8])
		Value := binary.LittleEndian.Uint64(msg[8:16])
		TxData := string(msg[16:80])
		checkAddress := string(msg[80:114])
		trFrom := string(msg[114:148])
		go waitForBlock(userId, TxData, Value, checkAddress, trFrom, client)
	}
}

/*
function waitForBlock
function for Tx that was sent from our servers
we will get an error and litecoin' API critical error if the transaction, not in the real block and has status = Unconfirmed
*/
func waitForBlock(userId uint64, TxData string, Value uint64, checkAddress string, trFrom string, client *rpcclient.Client) {
	blockHash, err := chainhash.NewHashFromStr(TxData)
	checkError(err)
	for {
		Tx, _ := client.GetRawTransactionVerbose(blockHash)
		if Tx.BlockHash == "" || len(Tx.BlockHash) < 5{
			time.Sleep(20 * time.Second)
			continue
		}
		save_TransactionOUT(Tx, checkAddress, trFrom, Value, userId, 0)
		break
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
	tmp.CoinID = 3
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
handling requests from the Wallet server for adding a  new user address to the address list
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
		if len(msg) != 94 {
			continue
		}
		userId := binary.LittleEndian.Uint64(msg[0:8])
		address := string(msg[8:42])
		privateKey := string(msg[42:94])
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
		}else{
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

	Uid := uint64(trData.UserId)
	if len(trData.TransactionHash) > 10 && len(trData.TrValue) > 2 {
		binary.LittleEndian.PutUint64(tmpBuf[0:8], Uid)
		binary.LittleEndian.PutUint16(tmpBuf[8:10], uint16(trData.InOut))
		binary.LittleEndian.PutUint16(tmpBuf[10:12], uint16(trData.CoinID))
		binary.LittleEndian.PutUint16(tmpBuf[12:14], uint16(trData.Vout))
		binary.LittleEndian.PutUint16(tmpBuf[14:16], uint16(trData.ConfirmedBlocks))
		binary.LittleEndian.PutUint64(tmpBuf[16:24], uint64(trData.TrTime))

		copy(tmpBuf[24:66], []byte(trData.TrTo))
		copy(tmpBuf[66:108], []byte(trData.TrFrom))
		copy(tmpBuf[108:126], []byte(trData.TrValue))
		copy(tmpBuf[126:192], []byte(trData.TransactionHash))

		return tmpBuf
	}
	return tmpBuf
}

/*
function writeAllData
write transaction history
*/
func writeAllData() {
	TMPTransactionData.mux.Lock()
	sliceCpy := make([]transactions, len(TMPTransactionData.data))
	TMPTransactionData.mux.Unlock()

	var flag bool
	flag = false
	TMPTransactionData.mux.Lock()
	for n:=0;n<len(sliceCpy);n++ {
		if len(TMPTransactionData.data[n].transactionsHistory) <= 1{
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
			_ = writeDataToFile(sliceCpy[n].transactionsHistory, sliceCpy[n].id, &TransactionsHistory)
			sliceCpy[n].transactionsHistory = make([]byte,1)
		}

	}
	writeLogTofile(&TransactionsHistory)
	TMPTransactionData.mux.Unlock()
}

/*
function writeLogTofile
write the data to the map database for transactions storage
*/
func writeLogTofile(dataStorage *Storage) {
	file, err := os.OpenFile(fmt.Sprintf("../Litecoin/%vMap.db",dataStorage.name), os.O_WRONLY|os.O_CREATE, 0666)
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
		} else {
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
func writeDataToFile(listTransactions []byte, userId uint16, dataFiles *Storage) (int) {
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
		if (recordData.file != 1){
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
		file, err := os.OpenFile(fmt.Sprintf("../Litecoin/%vU%d.db", dataFiles.name, userId), os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			log.Fatal(err)
		}
		recordData.file = 4
		file.Seek(0,2)
		_, err = file.Write(listTransactions[beginFile*192:endFile*192])
		if err != nil {
			log.Fatal(err)
		}
		file.Close()
	}
	recordData.records = recordData.records + uint64(len(listTransactions)/192)
	return 1
}

/*
function readMapData
args: storage
function reads data from the storage and writes it to the binary array
*/
func readMapData(dataStorage *Storage) {
	var userId uint64
	var userSliceId int

	file, err := os.Open(fmt.Sprintf("../Litecoin/%vMap.db",dataStorage.name))

	defer file.Close()

	if err != nil {
		os.Exit(1)
	}

	fi, err := file.Stat()
	if err != nil {
		log.Fatal("Failed to get stat: ", err) // Could not obtain stat, handle error
	}
	var i uint64

	for i = 0; i < uint64(fi.Size()/48); i++ {
		data := readNextBytes(file, 48)
		if (len(data)<48){
			fmt.Println(len(data))
			break
		}
		if (i == uint64(len(dataStorage.historyRegistry))){
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

		if dataStorage.name == "transactionsHistory" {
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
			userOrdersSlice[userSliceId].transactionsHistory = readFileData(userId,read_strings, new_read_offset, dataStorage)
		}
	}
}

/*
function readFileDataOffset
args: User ID, how many records will be read from storage, read offset, storage
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

	if (begin < 200) {
		readOffset200 = begin
		if(begin + amount < 200) {
			read200 = amount
			return readOffset200, read200, readOffset1000, read1000, readOffset10000, read10000, readOffsetFile, readFile
		} else {
			read200 = 200 - begin
			readOffset1000 = 0
		}
	}

	if (begin < 1200) {
		if (begin > 200){
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
function readFileData
args: User ID,  how many records will be read from storage, read offset, storage
function returns records from storage
*/
func readFileData(userId uint64, amount uint64, offset uint64, dataFiles *Storage) ([]byte) {
	buffer := make([]byte,1)
	recordData := &dataFiles.historyRegistry[userId]

	if (recordData.records < offset || recordData.records == 0) {
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
		fileX, err := os.Open(fmt.Sprintf("../Litecoin/%vU%d.db", dataFiles.name, userId))

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
function sendTrhistory
sending transaction history to the wallets server
*/
func sendTrhistory() {
	for{
		if len(sendBufSlice.data) >= 208 {
			sendBufSlice.mux.Lock()
			sndTransactionHistory.SendBytes(sendBufSlice.data, 0)
			sendBufSlice.data = []byte{}
			sendBufSlice.mux.Unlock()
		}
		if len(sendBufSliceOUT.data) >= 80 {
			sendBufSliceOUT.mux.Lock()
			sndTransactionHistoryOUT.SendBytes(sendBufSliceOUT.data, 0)
			sendBufSliceOUT.data = []byte{}
			sendBufSliceOUT.mux.Unlock()
		}
		time.Sleep(10*time.Second)
	}
}

/*
function saveTransactions
saving users' transactions
*/
func saveTransactions() {
	for {
		filename := "../Litecoin/transactions.db"
		PendingTransactions.mux.Lock()
		write_transaction_to(filename, PendingTransactions.data, "transactions")
		PendingTransactions.mux.Unlock()

		filename = "../Litecoin/transactionsOUT.db"
		PendingTransactionsOUT.mux.Lock()
		write_transaction_to(filename, PendingTransactionsOUT.data, "transactionsOUT")
		PendingTransactionsOUT.mux.Unlock()

		ConfirmedTransactions.mux.Lock()
		filename = "../Litecoin/ConfirmedTransactions.db"
		write_transaction_to(filename, ConfirmedTransactions.data, "ConfirmedTransactions")
		ConfirmedTransactions.mux.Unlock()

		ConfirmedTransactionsOUT.mux.Lock()
		filename = "../Litecoin/ConfirmedTransactionsOUT.db"
		write_transaction_to(filename, ConfirmedTransactionsOUT.data, "ConfirmedTransactionsOUT")
		ConfirmedTransactionsOUT.mux.Unlock()

		time.Sleep(10*time.Second)
	}
}

var globalCounter int

func main() {
	// make maps with without nil data
	usersTrunsactionsT.data = make(map[string]UserTrunsaction)
	ConfirmedTransactions.data = make(map[string] *btcjson.TxRawResult)
	ConfirmedTransactionsOUT.data = make(map[string] *btcjson.TxRawResult)
	PendingTransactions.data = make(map[string] *btcjson.TxRawResult)
	PendingTransactionsOUT.data = make(map[string] *btcjson.TxRawResult)
	PendingAdderess.data = make(map[uint64]string)
	UsersHasAddr.data = make(map[uint64]HasAddress)
	usersTrunsactionsOUT.data = make(map[string] UserTrunsaction)
	TMPTransactionData.data=  make([]transactions, 1, 1)
	ColdWalletTrunsactionsIN.data = make(map[string] UserTrunsaction)
	//-----
	globalCounter = 0
	StressTestFlag = false
	ColdWallet = LocalConfig.LitecoinColdWallet
	CustomNetwork.DefaultPort = LocalConfig.LitecoinDefaultPort

	//Creating connection with Balance server (send)
	//add founds
	sndBalance, _ := zmq.NewSocket(zmq.PUSH)
	defer sndBalance.Close()
	sndBalance.SetRcvhwm(1100000)
	err := sndBalance.Bind(fmt.Sprintf("%v%v", LocalConfig.LocalTcpIpAddr, LocalConfig.LTCBalancePort));
	if (err != nil){
		fmt.Println(err)
	}

	//Creating connection with Litecoin OUT server (receive)
	//receive new transaction
	rcvOutTransaction, _ = zmq.NewSocket(zmq.PULL)
	defer rcvOutTransaction.Close()
	rcvOutTransaction.SetRcvhwm(1100000)
	rcvOutTransactionAddress := fmt.Sprintf("%v%v", LocalConfig.LocalTcpIpAddr, LocalConfig.LTCOutTrnsctnPort)
	rcvOutTransaction.Connect(rcvOutTransactionAddress)

	//Creating connection with Wallets  server (receive)
	//receive new wallet address
	rcvNewUserAddr, _ = zmq.NewSocket(zmq.PULL)
	defer rcvNewUserAddr.Close()
	rcvNewUserAddr.SetRcvhwm(1100000)
	rcvAddress := fmt.Sprintf("%v%v", LocalConfig.LocalTcpIpAddr, LocalConfig.LTCrcvNewAddressPort)
	rcvNewUserAddr.Connect(rcvAddress)

	//Creating connection with Wallets  server (send)
	//send history
	sndTransactionHistory, _ = zmq.NewSocket(zmq.PUSH)
	defer sndTransactionHistory.Close()
	sndTransactionHistory.SetRcvhwm(1100000)
	sndTransactionHistoryAddress := fmt.Sprintf("%v%v", LocalConfig.LocalTcpIpAddr, LocalConfig.LTCsndTrnsctnHstrPort)
	err = sndTransactionHistory.Bind(sndTransactionHistoryAddress)
	checkError(err)

	//Creating connection with Litecoin OUT  server (send)
	sndTransactionHistoryOUT, _ = zmq.NewSocket(zmq.PUSH)
	defer sndTransactionHistoryOUT.Close()
	sndTransactionHistoryOUT.SetRcvhwm(1100002)
	sndTransactionHistoryAddressOUT := fmt.Sprintf("%v%v", LocalConfig.LocalTcpIpAddr, LocalConfig.LTCsndTrnsctnHstrOUTPORT)
	err = sndTransactionHistoryOUT.Bind(sndTransactionHistoryAddressOUT)
	checkError(err)

	//Creating connection with Litecoin OUT  server (send)
	//send cold waleet TxIn
	sndColdTxInOUT, _ = zmq.NewSocket(zmq.PUSH)
	defer sndColdTxInOUT.Close()
	sndColdTxInOUT.SetRcvhwm(1100001)
	sndColdTxInAddressOUT := fmt.Sprintf("%v%v", LocalConfig.LocalTcpIpAddr, LocalConfig.LTCsndColdTxInOUTPort)
	err = sndColdTxInOUT.Bind(sndColdTxInAddressOUT)
	checkError(err)

	connCfg := &rpcclient.ConnConfig{
		Host:         LocalConfig.LitecoinHost,
		User:         LocalConfig.LitecoinCoreUser,
		Pass:         LocalConfig.LitecoinCorePassw,
		HTTPPostMode: true, // Litecoin core only supports HTTP POST mode
		DisableTLS:   true, // Litecoin core does not provide TLS by default
	}

	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Shutdown()

	//Loading data from Litecoin Transactions Storage
	TransactionsHistory.file200, err = os.OpenFile("../Litecoin/transactionsHistory200.db", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer TransactionsHistory.file200.Close()

	TransactionsHistory.file1000, err = os.OpenFile("../Litecoin/transactionsHistory1000.db", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer TransactionsHistory.file1000.Close()

	TransactionsHistory.file10000, err = os.OpenFile("../Litecoin/transactionsHistory10000.db", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer TransactionsHistory.file10000.Close()
	TransactionsHistory.name = "transactionsHistory"

	//Read data from Litecoin Transactions Storage
	readMapData(&TransactionsHistory)

	loadUsersWallets()
	LoadConfirmedTransactions(client)
	LoadConfirmedTransactionsOUT(client)
	load_Transactions(client)
	load_TransactionsOUT(client)
	time.Sleep(1*time.Second)
	go addTrOUTHistory(client)
	go change_trnsctn_status(client)
	go change_trnsctn_statusOUT(client)
	go sendPing()
	go rcvnewAddr()
	go sendTrhistory()
	go saveTransactions()
	time.Sleep(1*time.Second)
	var lastProcessedBlockNumber int64
	//Stress test run
	//StressTest()
	//---------
	for {
		//load last processed block
		lastProcessedBlock := loadLastBlock()
		if len(lastProcessedBlock) <= 3 {
			lastProcessedBlockNumber = 0
		} else {
			lastProcessedBlockNumber = getblockHeight(client, lastProcessedBlock)
		}
		//get last block
		prevBlockHash := get_last_block(client)
		latestBlockNumber := getblockHeight(client, prevBlockHash)
		if (prevBlockHash != lastProcessedBlock) {
			saveLastBlock(prevBlockHash)
			for i:= latestBlockNumber; i > lastProcessedBlockNumber; i-- {
				prevBlockHash = get_block(client, i)
				//обработка блоков, проверка транзакций
				process_blocks(client, prevBlockHash)
				if i == lastProcessedBlockNumber {
					break
				}
			}
		}

		if len(TMPTransactionData.data) >= 1 {
			time.Sleep(1*time.Second)
			writeAllData()
		}

		buf := generateBalanceBuf()
		if len(buf) > 10 {
			sndBalance.SendBytes(buf,0)
		}
		time.Sleep(20*time.Second)
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
	}
}

func StressTest() {
	StressTestFlag = true

	time.Sleep(10*time.Second)

	var Tx *btcjson.TxRawResult
	var userID uint64
	var checkAddress string
	var from string

	var countEVEN int
	var countODD int

	countEVEN = 0
	countODD = 0

	for n:=0; n < 1000; n++ {
		if n%2 == 0 {
			countEVEN++
			userID = 3
			checkAddress = "mmFpQmxG1Qe2SvTe1toyeAvgVrVuVzT3vq"
			from = "mtoqkpTC13PawrXTxHcjD6Xu6MHQqTwA3Q"

			PendingAdderess.mux.Lock()
			PendingAdderess.data[userID] = checkAddress //add address to Pending map (user that will receive coins)
			PendingAdderess.mux.Unlock()
		} else {
			countODD++
			userID = 4
			checkAddress = "mtoqkpTC13PawrXTxHcjD6Xu6MHQqTwA3Q"
			from = "mmFpQmxG1Qe2SvTe1toyeAvgVrVuVzT3vq"
		}

		var Ending string
		var txIdtest string
		var blockHash string
		Ending = fmt.Sprintf("%v",n)
		if len(Ending) == 1{
			txIdtest = fmt.Sprintf("%vdbf9f87880211a12b3940b1de725df4e8b7ed7cf6c72821b4fb3f620f1c6c0a", n)
			blockHash = fmt.Sprintf("%v0000000000000cddd538c040e608704c30446b8cf5bea0834b0c688c4a5527a", n)
		}
		if len(Ending) == 2{
			txIdtest = fmt.Sprintf("%vdbf9f87880211a12b3940b1de725df4e8b7ed7cf6c72821b4fb3f620f1c6c0", n)
			blockHash = fmt.Sprintf("%v0000000000000cddd538c040e608704c30446b8cf5bea0834b0c688c4a5527", n)
		}
		if len(Ending) == 3{
			txIdtest = fmt.Sprintf("%vdbf9f87880211a12b3940b1de725df4e8b7ed7cf6c72821b4fb3f620f1c6c", n)
			blockHash = fmt.Sprintf("%v0000000000000cddd538c040e608704c30446b8cf5bea0834b0c688c4a552", n)
		}
		if len(Ending) == 4{
			txIdtest = fmt.Sprintf("%vdbf9f87880211a12b3940b1de725df4e8b7ed7cf6c72821b4fb3f620f1c6", n)
			blockHash = fmt.Sprintf("%v0000000000000cddd538c040e608704c30446b8cf5bea0834b0c688c4a55", n)
		}

		Tx = &btcjson.TxRawResult{
			"",
			txIdtest,
			blockHash,
			1,
			1,
			1,
			0,
			uint32(time.Now().Unix()),
			[]btcjson.Vin{},
			[]btcjson.Vout{},
			blockHash,
			1,
			time.Now().Unix(),
			time.Now().Unix()}

		if n%2 == 0 {
			save_Transaction(Tx, checkAddress, float64(n+100), userID, uint64(n), 0)
		} else {
			save_TransactionOUT(Tx, checkAddress, from, uint64(n+100), userID, 0)
		}
	}
	fmt.Println("globalCounter------ ", globalCounter)
	fmt.Println("globalCounter------ ", globalCounter)
	fmt.Println("globalCounter------ ", globalCounter)
	for {
		time.Sleep(30*time.Second)
		fmt.Println("countODD----- ", countODD)
		fmt.Println("countEVEN----- ", countEVEN)
		if len(TMPTransactionData.data) >= 1 {
			writeAllData()
		}

		fmt.Println("globalCounter------ ", globalCounter)
		fmt.Println("globalCounter------ ", globalCounter)
		fmt.Println("globalCounter------ ", globalCounter)
	}
}
func StressTransuction(TranXon *btcjson.TxRawResult) (bool, uint64){
	return true, 4
}