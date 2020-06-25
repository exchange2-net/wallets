package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	customBTC "github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"log"
	"os"
)

type Network struct {
	name        string
	symbol      string
	xpubkey     byte
	xprivatekey byte
}

//Setup params
var network = map[string]Network{
	"rdd": {name: "reddcoin", symbol: "rdd", xpubkey: 0x3d, xprivatekey: 0xbd},
	"dgb": {name: "digibyte", symbol: "dgb", xpubkey: 0x1e, xprivatekey: 0x80},
	"btc": {name: "bitcoin",  symbol: "btc", xpubkey: 0x00, xprivatekey: 0x80},
	"ltc": {name: "litecoin", symbol: "ltc", xpubkey: 0x30, xprivatekey: 0xb0},
}

// string a function name
var funcMap = map[string]interface{}{
	"Bitcoin" : DecodeBitcoin,
}

var PrivateKeys = make(map[string]string)
var PublicKeys = make(map[string]string)
var NetPorts = make(map[string]string)


//Setup params
func (network Network) GetNetworkParams() *customBTC.Params {
	networkParams := &customBTC.TestNet3Params
	networkParams.DefaultPort = "18443"
	return networkParams
}


//Setup params
func (network Network) CreatePrivateKey() (*btcutil.WIF, error) {
	secret, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return nil, err
	}
	return btcutil.NewWIF(secret, network.GetNetworkParams(), true)
}


//Setup params
func (network Network) ImportWIF(wifStr string) (*btcutil.WIF, error) {
	wif, err := btcutil.DecodeWIF(wifStr)
	if err != nil {
		return nil, err
	}
	if !wif.IsForNet(network.GetNetworkParams()) {
		return nil, errors.New("The WIF string is not valid for the `" + network.name + "` network")
	}
	return wif, nil
}


//Setup params
func (network Network) GetAddress(wif *btcutil.WIF) (*btcutil.AddressPubKey, error) {
	return btcutil.NewAddressPubKey(wif.PrivKey.PubKey().SerializeCompressed(), network.GetNetworkParams())
}

func main() {
	//set your CORES' ports
	NetPorts["Bitcoin"] = "18443"

	create_wallets()//Create Private key and Encode Address
	CreatePublicAddr("Bitcoin") //Create Real public BTC addres
	writeAddrData("Bitcoin")  // Write keys to database
}

/*
function writeAddrData
write keys from the map to database
*/
func writeAddrData(folder string) {
	Path := fmt.Sprintf("../users/FreeKeys%v.db", folder)
	file, err := os.OpenFile(Path, os.O_APPEND|os.O_WRONLY, 0666)
	defer file.Close()

	if err != nil {
		log.Fatal(err)
	}

	for PrivateKey, PublicKey := range PublicKeys {
		//fmt.Println(PublicKey)
		b := make([]byte, 52)
		copy(b[:], []byte(PrivateKey))
		bufferedWriter := bufio.NewWriter(file)
		_, err = bufferedWriter.Write( b, )
		if err != nil {
			log.Fatal("writeFile 103 ",err)
		}

		b = make([]byte, 34)
		copy(b[:], []byte(PublicKey))
		_, err = bufferedWriter.Write( b, )
		if err != nil {
			log.Fatal("writeFile 110", err)
		}

		bufferedWriter.Flush()
	}
	file.Close()
}
/*
call the function by her string name
 */
func CreatePublicAddr(Coin string) {
	funcMap[Coin].(func())()
}

/*
function DecodeBitcoin
function for decoding address (create public address)
 */
func DecodeBitcoin() {
	for Private, keys := range PrivateKeys {
		PublicKeys[Private] = keys
	}
}


func create_wallets() {
	PrivateKeys = make(map[string]string)
	i := 1
	for i <= 100 {
		wif, _ := network["btc"].CreatePrivateKey()

		if (len([]byte(wif.String())) < 52 ) {
			continue
		}
		address, _ := network["btc"].GetAddress(wif)
		PrivateKeys[wif.String()] = string(address.EncodeAddress())
		fmt.Println("%s - %s", wif.String(), address.EncodeAddress())
		i = i + 1
	}
}