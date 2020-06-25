package main

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/smtp"
	"os"
	"strings"
	"sync"
	"time"
	"wallets/walletsP/LocalConfig"
	"wallets/walletsP/Pairs"
)

var handleAUTH *zmq.Socket // hendle registration, login or logout 3701
var handleWallets *zmq.Socket //hendle reseve or send coins 3702

var PairData[1000]Pairs.WalletType
var delimiter[1000]uint64
var tcpPairs[1000]Pairs.Pair

var MailData struct {
	data map[string]struct {
		subject string
		file string
	}
	mux sync.Mutex
}

var Mdata struct {
	data map[uint64]RecivedData
	mux sync.Mutex
}

type RecivedData struct {
	email string
	password string
	ipAddr string
	Value uint64
	CoinID uint64
	FromAddr string
	ToAddr string
	subject string
	UserAgent string
}

func main() {
	//
	MailData.data = make(map[string]struct {
		subject string
		file string
	})
	Mdata.data = make(map[uint64]RecivedData)
	//------
	fmt.Println("Starting mail server")
	PairData = Pairs.WalletList
	delimiter = Pairs.Delimiter
	tcpPairs = Pairs.PairList
	loadSubjects()

	load_testdata()

	//listen Auth and Wallets
	handleAUTH, _ = zmq.NewSocket(zmq.PULL)
	defer handleAUTH.Close()
	handleAUTH.SetRcvhwm(1100000)
	handleAUTH.Connect(fmt.Sprintf("%v%v", LocalConfig.LocalTcpIpAddr, LocalConfig.MailhandleAUTHPort))

	handleWallets, _ = zmq.NewSocket(zmq.PULL)
	defer handleWallets.Close()
	handleWallets.SetRcvhwm(1100000)
	handleWallets.Connect(fmt.Sprintf("%v%v", LocalConfig.LocalTcpIpAddr, LocalConfig.MailhandleWallets))

	go listenAuthServer()
	go sendMailToUser()
	listenWalletsServer()
}

func load_testdata() {
	Mdata.data[1] = RecivedData{
		email:     "dovbysh-2000@ukr.net",
		//email:     "1silentforce@ukr.net",
		password:  "password",
		ipAddr:    "ipAddr",
		Value:     0,
		CoinID:    0,
		FromAddr:  "",
		ToAddr:    "",
		subject:   "registartion",
		UserAgent: "UserAgent",
	}
}

func sendMailToUser() {
	for{
		Mdata.mux.Lock()
		if len(Mdata.data) >= 1 {
			sand_mail(Mdata.data)
			MdataLen := len(Mdata.data)
			if MdataLen == 0 {
				Mdata.data = make(map[uint64]RecivedData)
			}
			//Mdata.data = make(map[uint64]RecivedData) //TODO - delete
		}
		Mdata.mux.Unlock()
		time.Sleep(1*time.Minute)
	}

}

/*
function listenAuthServer
function listens a request from the authorization server.
Request types:
- send Email about registration
- send Email about log out
- send Email about log in
 */
func listenAuthServer() {
	poller := zmq.NewPoller()
	poller.Add(handleAUTH, zmq.POLLIN)
	for {
		_, err := poller.Poll(-1)
		if err != nil {
			eer2 := fmt.Sprintf("Socket error: %v", err)
			fmt.Println(eer2)
			continue //  Interrupted
		}
		msg, _ := handleAUTH.RecvBytes(0)
		if len(msg) < 350 {
			continue
		}
		go processAUTData(msg)
	}
}

func processAUTData(msg []byte) {
	//var Mdata RecivedData
	var Mtype uint64
	var ii uint64

	/*RecivedData
	0 - type (data[0:8] 1 - registration, 2 - login, 3 - logout) // uint64
	1 - mail (data[8:58]) // string //login
	2 - ipAddr (data[58:100])
	3 - password (data[100:200])
	*/

	Mdata.mux.Lock()
	var offset int
	offset = 0
	for ij := 0; ij < (len(msg) / 350); ij++ {
		Mtype = binary.LittleEndian.Uint64(msg[0+offset:8+offset])

		var subject string
		ii = uint64(len(Mdata.data))

		if Mtype == 1 {
			subject = "registartion"
		}
		if Mtype == 2 {
			subject = "login"
		}
		if Mtype == 3 {
			subject = "logout"
		}

		email := string(bytes.Trim(msg[8+offset:58+offset], "\x00"))
		ipAddr := string(bytes.Trim(msg[58+offset:100+offset], "\x00"))
		password := string(bytes.Trim(msg[100+offset:200+offset], "\x00"))
		UserAgent := string(bytes.Trim(msg[200+offset:350+offset], "\x00"))

		Mdata.data[ii] = RecivedData{
			email:     email,
			password:  password,
			ipAddr:    ipAddr,
			Value:     0,
			CoinID:    0,
			FromAddr:  "",
			ToAddr:    "",
			subject:   subject,
			UserAgent: UserAgent,
		}
		offset = offset + 350
	}
	Mdata.mux.Unlock()
}
/*
function listenWalletsServer
function listens a  request from the Wallets server.
Request types:
- send Email about receiving coins
- send Email about sending coins
*/
func listenWalletsServer() {
	poller := zmq.NewPoller()
	poller.Add(handleWallets, zmq.POLLIN)
	for {
		_, err := poller.Poll(-1)
		if err != nil {
			eer2 := fmt.Sprintf("Socket error: %v", err)
			fmt.Println(eer2)
			continue //  Interrupted
		}
		msg, _ := handleWallets.RecvBytes(0)

		if len(msg) < 158 {
			continue
		}
		go procwssWData(msg)
	}
}

func procwssWData(msg []byte) {
	//var Mdata RecivedData
	var Mtype uint64

	/*RecivedData
	0 - type (data[0:8] 1 - reciveCoins, 2 - sendCoins) // uint64
	1 - Value (data[8:16])
	2 - email (data[16:66])
	3 - CoinShort (data[66:70])
	4 - FromAddr (data[70:122])
	5 - ToAddr (data[122:154])
	*/
	var ii uint64

	Mdata.mux.Lock()
	var offset int
	offset = 0
	for ij := 0; ij < (len(msg) / 158); ij++ {
		Mtype = uint64(binary.LittleEndian.Uint16(msg[0+offset:2+offset]))
		var subject string
		ii = uint64(len(Mdata.data))
		if Mtype == 1 {
			subject= "reciveCoins"
		}
		if Mtype == 2 {
			subject = "sendCoins"
		}

		Value := binary.LittleEndian.Uint64(msg[2+offset:10+offset])
		CoinID := binary.LittleEndian.Uint64(msg[10+offset:18+offset])

		email := string(bytes.Trim(msg[18+offset:74+offset], "\x00"))
		FromAddr := string(bytes.Trim(msg[74+offset:116+offset], "\x00"))
		ToAddr := string(bytes.Trim(msg[116+offset:158+offset], "\x00"))

		Mdata.data[ii] = RecivedData{
			email:     email,
			password:  "",
			ipAddr:    "",
			Value:     Value,
			CoinID:    CoinID,
			FromAddr:  FromAddr,
			ToAddr:    ToAddr,
			subject:   subject,
			UserAgent: "",
		}
		offset = offset + 158
	}
	Mdata.mux.Unlock()
}
/*
function loadSubjects
function to initialize data:
- path to HTML templates
- Email subjects
 */
func loadSubjects() {
	var tmp struct {
		subject string
		file string
	}

	MailData.mux.Lock()
	tmp.subject = "Login to exchange2.net"
	tmp.file = "Login.html"
	MailData.data["login"] = tmp

	tmp.subject = "Registartion to exchange2.net"
	tmp.file = "Registration.html"
	MailData.data["registartion"] = tmp

	tmp.subject = "Logout from exchange2.net"
	tmp.file = "Logout.html"
	MailData.data["logout"] = tmp

	tmp.subject = "Send coins"
	tmp.file = "SendCoins.html"
	MailData.data["sendCoins"] = tmp

	tmp.subject = "Recive coins"
	tmp.file = "ReciveCoins.html"
	MailData.data["reciveCoins"] = tmp
	MailData.mux.Unlock()
}

/*
function loadMailTemplate
args: path to HTML template
fuction for loading HTML template
 */
func loadMailTemplate(file string) string{
	path := fmt.Sprintf("../mailTemplates/%v", file)
	template, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	return string(template)
}

/*
function replaceText
args: mail data, current time
function for replacing HTML template text by mail data
 */
func replaceText(mData RecivedData, current_time string) string{
	var value float64
	var valueSTR string
	if mData.Value != 0 {
		value = float64(mData.Value)/float64(delimiter[tcpPairs[mData.CoinID].Coin])
	} else {
		value = 0
	}

	valueSTR = fmt.Sprintf("%v",value)

	template := loadMailTemplate(MailData.data[mData.subject].file)
	template = strings.Replace(template, "{{email}}", mData.email, -2)
	template = strings.Replace(template, "{{password}}", mData.password, -2)
	template = strings.Replace(template, "{{loginTime}}", current_time, -2)
	template = strings.Replace(template, "{{ipAddr}}", mData.ipAddr, -2)
	template = strings.Replace(template, "{{UserAgent}}", mData.UserAgent, -2)
	template = strings.Replace(template, "{{Value}}", valueSTR, -2)
	template = strings.Replace(template, "{{CoinShort}}", PairData[mData.CoinID].WalletShort, -2)
	template = strings.Replace(template, "{{FromAddr}}", mData.FromAddr, -2)
	template = strings.Replace(template, "{{ToAddr}}", mData.ToAddr, -2)

	return template
}

func createMail(mData RecivedData) (string, string) {
	current_time := time.Now().String()
	mailText := ""
	//set mail texts

	mailText = replaceText(mData, current_time)
	subject := MailData.data[mData.subject].subject

	//set Headers
	headers := make(map[string]string)
	headers["From"] = LocalConfig.SendGridMailFrom
	headers["To"] = mData.email
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=\"UTF-8\""
	headers["Content-Transfer-Encoding"] = "base64"
	//set Headers END------
	// Setup message
	var message = ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}

	message += "\r\n" + base64.StdEncoding.EncodeToString([]byte(mailText))

	return mData.email, message
}

/*
function sand_mail
args: message
function to create connection to mail service and after that it sends a message
 */
func sand_mail(mData map[uint64]RecivedData) {
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	conn, err := net.DialTimeout("tcp", LocalConfig.SendGridHostAndTLSPort, 10*time.Minute)
	if err != nil {
		fmt.Println(err)
	}

	connection, err := smtp.NewClient(conn, LocalConfig.SendGridHost)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	connection.StartTLS(tlsconfig)

	auth := smtp.PlainAuth("", LocalConfig.SendGridLogin, LocalConfig.SendGridPassw, LocalConfig.SendGridHost)

	if err = connection.Auth(auth); err != nil {
		fmt.Println(err)
		return
	}

	connection.Hello("hello")

	mDataCMP := mData

	for key, data := range mDataCMP {

		var w io.WriteCloser

		email,message := createMail(data)

		if err = connection.Mail(LocalConfig.SendGridMailFrom); err != nil {
			fmt.Println(err)
			return
		}

		if err = connection.Rcpt(email); err != nil {
			fmt.Println(err)
			return
		}

		w, err = connection.Data()
		if err != nil {
			fmt.Println(err)
			return
		}

		_, err = w.Write([]byte(message))
		if err != nil {
			fmt.Println(err)
			return
		}

		err = w.Close()
		if err != nil {
			fmt.Println(err)
			return
		}
		delete(mData, key)
		time.Sleep(250*time.Millisecond)
	}
	connection.Quit()
}