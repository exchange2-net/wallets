// 3000000000000000000 - 3 eth
package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/binary"
	"encoding/hex"
	"wallets/walletsP/LocalConfig"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
	zmq "github.com/pebbe/zmq4"
	"log"
	"math/big"
	"os"
	"time"
)

var sndTransactionAnswer *zmq.Socket //ZMQ, socket variable / send the transaction result to the Wallets Server
var sndBalance *zmq.Socket //ZMQ, socket variable // if we have an error in transaction we return coins
var rcvTransaction *zmq.Socket //ZMQ, socket variable // receive a request form the Wallets server

/*
function sendRawTransaction
args: amount coins, Gas Limit, Gas Price, Sender Private key, reciver addr
function for creating transactions and sending it to blockchain
 */
func sendRawTransaction(value *big.Int, gasLimit uint64, gasPrice *big.Int, privateKeyStr string, sndTo string) string {
	//connecting to infura services
	url := fmt.Sprintf("%v%v", LocalConfig.InfuraRinkebyApiURL, LocalConfig.InfuraProjectID)
	client, err := ethclient.Dial(url)
	if err != nil {
		return ""
		log.Fatal(err)
	}
	privateKey, err := crypto.HexToECDSA(privateKeyStr)
	if err != nil {
		return ""
		log.Fatal(err)
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey) // convert private key
	if !ok {
		return ""
		log.Fatal("error casting public key to ECDSA")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return ""
	}
	toAddress := common.HexToAddress(sndTo) //convert receiver address
	var data []byte
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data) // Create new Transaction
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return ""
	}
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey) //signed Transaction
	if err != nil {
		return ""
	}
	ts := types.Transactions{signedTx}
	rawTx := hex.EncodeToString(ts.GetRlp(0))
	rawTxBytes, err := hex.DecodeString(rawTx)
	rlp.DecodeBytes(rawTxBytes, &tx)
	err = client.SendTransaction(context.Background(), tx) // send to blockchain
	if err == nil {
		fmt.Println(rawTx)
		return tx.Hash().Hex()
	} else {
		fmt.Println(err)
		return ""
	}
	return ""
}

/*
function listen_wallet
function listening socket for request to make a new transaction
*/
func listen_wallet() {
	poller := zmq.NewPoller()
	poller.Add(rcvTransaction, zmq.POLLIN)
	for {
		_, err := poller.Poll(-1)
		if err !=nil {
			fmt.Println("error, receve addr: ", err)
		}
		msg,_ := rcvTransaction.RecvBytes(0)
		if len(msg) != 122 {
			continue
		}
		userId := binary.LittleEndian.Uint64(msg[0:8])
		value := binary.LittleEndian.Uint64(msg[8:16])
		privateKeyStr := string(msg[16:80])
		sndTo := string(msg[80:122])
		valueCPM := value
		value = value - 420000 // - Fee 42

		gasLimit := uint64(21000)
		gasPrice := big.NewInt(20000000000)
		amount := new(big.Int).SetUint64(value*1000000000)

		Tx := sendRawTransaction(amount, gasLimit, gasPrice, privateKeyStr, sndTo)
		if len(Tx) < 3 {
			buf := make([]byte, 3)
			response := "0x0"
			copy(buf[0:3], []byte(response))
			sndTransactionAnswer.SendBytes(buf, 0)
			//send error
			lineBuf := make([]byte, 14)

			binary.LittleEndian.PutUint16(lineBuf[0:2], 4)
			binary.LittleEndian.PutUint32(lineBuf[2:6], uint32(userId))
			binary.LittleEndian.PutUint64(lineBuf[6:14],valueCPM)

			sndBalance.SendBytes(lineBuf,0)
			time.Sleep(1*time.Second)
		} else {
			//send TX
			buf := make([]byte, 66)
			copy(buf[0:66], []byte(Tx))
			sndTransactionAnswer.SendBytes(buf, 0)
			time.Sleep(1*time.Second)
		}
	}
}

func main() {
	sndTransactionAnswer, _ = zmq.NewSocket(zmq.PUSH) //create new socket for sending Transaction result
	defer sndTransactionAnswer.Close()
	sndTransactionAnswer.SetRcvhwm(1100000)
	err := sndTransactionAnswer.Bind(fmt.Sprintf("%v%v", LocalConfig.LocalTcpIpAddr, 7414))
	checkError(err)

	rcvTransaction, _ = zmq.NewSocket(zmq.PULL)  //create new socket for receiving requests from the  Wallet server
	defer rcvTransaction.Close()
	rcvTransaction.SetRcvhwm(1100000)
	rcvAddress := fmt.Sprintf("%v%v", LocalConfig.LocalTcpIpAddr, 7204)
	err = rcvTransaction.Connect(rcvAddress)
	checkError(err)

	sndBalance, _ = zmq.NewSocket(zmq.PUSH) //create new socket for return coins if we have some error
	defer sndBalance.Close()
	sndBalance.SetRcvhwm(1100000)
	rcvBalanceAddress := fmt.Sprintf("%v%v", LocalConfig.LocalTcpIpAddr, 9094)
	err = sndBalance.Bind(rcvBalanceAddress)
	checkError(err)

	listen_wallet()

	os.Exit(1) //exit with error
}

func checkError(err error) {
    if err != nil {
        fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
        os.Exit(1)
    }
}
