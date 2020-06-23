package main

import (
	"bufio"
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"exchange2/LocalConfig"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

var dirtySalt cryptosalt
var keysMap = make(map[string]string)

type cryptosalt struct {
	Crypto struct {
		Kdfparams struct{
			Salt string `json:"salt"`
		} `json:"kdfparams"`
	} `json:"crypto"`
}

//first step - run geth
func main() {
	createETHAddr()
	writeAddrData("Ethereum")
}

/*
function writeAddrData
write keys from map to databese
 */
func writeAddrData(folder string) {
	Path := fmt.Sprintf("../users/FreeKeys%v.db", folder)
	file, err := os.OpenFile(Path, os.O_APPEND|os.O_WRONLY, 0666)
	defer file.Close()

	if err != nil {
		log.Fatal(err)
	}

	for PrivateKey, PublicKey := range keysMap {
		b := make([]byte, 64)
		copy(b[:], []byte(PrivateKey))
		bufferedWriter := bufio.NewWriter(file)
		_, err = bufferedWriter.Write( b, )
		if err != nil {
			log.Fatal("writeFile 103 ",err)
		}

		b = make([]byte, 42)
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
function createETHAddr
loop for creating address
 */
func createETHAddr() {
	i := 1
	for i <= 100 {
		account := create_account()
		salt := getSalt(account.KeyPath)
		fmt.Println(len(salt))
		decodeEthereum(salt)
		i = i + 1
	}
}

func decodeEthereum(salt string) {
	privateKey, err := crypto.HexToECDSA(salt)
	if err != nil {
		log.Fatal(err)
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	fmt.Println(fromAddress.String())
	keysMap[salt] = fromAddress.String()
}

/*
function getSalt
get private key (salt) from file
 */
func getSalt(name string) string{
	var key_PATH = "../../../Library/Ethereum/keystore/"
	path := fmt.Sprintf("%v%v", key_PATH, name)
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		log.Fatal("loadUserAddresses",err)
	}
	fi, err := file.Stat()
	if err != nil {
		log.Fatal("Could not obtain stat", err)
	}
	data := readNextBytes(file, int(fi.Size()))
	result := &dirtySalt
	err = json.Unmarshal(data, result)
	if err != nil {
		panic(err)
	}

	return result.Crypto.Kdfparams.Salt
}

/*
function create_account
send request to Ethereum core
Create New User account and get ETH address.
 */
func create_account() LocalConfig.New_account{
	var account LocalConfig.New_account

	url := ""
	password := ""

	str := fmt.Sprintf("{\"jsonrpc\":\"2.0\",\"method\":\"personal_newAccount\",\"params\":[\"%s\"],\"id\":1}",password)
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
	result := &LocalConfig.Account_result{}
	err = json.Unmarshal([]byte(body), result)
	if err != nil {
		panic(err)
	}
	account.Address = result.Result

	reg_expr := regexp.MustCompile(".+x")
	reg_result := reg_expr.ReplaceAllString(result.Result, "")
	account.KeyPath = get_key_file(reg_result)
	return account
}

/*
function get_key_file
return full path to kestore file
 */
func get_key_file(address string) string {
	var key_PATH = "../../../Library/Ethereum/keystore/"
	var create_regex string
	var key string
	create_regex = fmt.Sprintf("%v",address)
	err := filepath.Walk(key_PATH, func(key_PATH string, info os.FileInfo, err error) error {
		found,err := regexp.MatchString(create_regex,info.Name())
		if found != false {
			key = info.Name()
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	return key
}

/*
function readNextBytes
function return N bytes from file
*/
func readNextBytes(file *os.File, number int) []byte {
	bytes := make([]byte, number)
	_, err := file.Read(bytes)
	if err != nil {
		log.Fatal(err)
	}
	return bytes
}