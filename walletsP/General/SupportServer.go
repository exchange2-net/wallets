package main

import (
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"wallets/walletsP/LocalConfig"
)

var requestNumber uint64

var UserRequestNumber struct {
	data map [uint64]uint64
	mux sync.Mutex
}

var Mdata struct {
	data map[uint64]RecivedData
	mux sync.Mutex
}

type RecivedData struct {
	email string
	issue string
	wallet string
	userId uint64
	wallet_address string
	description string
}

func main() {
	requestNumber = 1
	UserRequestNumber.data = make(map[uint64]uint64)

	loadRequests()

	Mdata.data = make(map[uint64]RecivedData)
	httpServer := http.NewServeMux()
	httpServer.HandleFunc("/", handleRequest)
	server_port := fmt.Sprintf(":%v", 25001)
	go func() {
		http.ListenAndServe(server_port, httpServer)
	}()

	sendMailToUser()
}

func sendMailToUser() {
	for{
		Mdata.mux.Lock()
		if len(Mdata.data) >= 1 {
			send_mail_to_support(Mdata.data)
			send_mail_to_user(Mdata.data)

			MdataLen := len(Mdata.data)
			if MdataLen == 0 {
				Mdata.data = make(map[uint64]RecivedData)
			}
		}
		Mdata.mux.Unlock()

		saveRequests()
		time.Sleep(1*time.Minute)
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Connection", "close")

	cmd := r.FormValue("data")
	data := strings.Split(cmd, ",")

	if (len(data) < 6) {
		w.Write([]byte("nFAIL182"))
		return //wrong request
	}

	userId, _ := strconv.ParseUint(data[2], 10, 0)

	if (data[0] == "SupportRequest") {
		email := strings.TrimSpace(r.FormValue("email"))
		issue := strings.TrimSpace(r.FormValue("issue"))
		wallet := strings.TrimSpace(r.FormValue("wallet"))
		wallet_address := strings.TrimSpace(r.FormValue("wallet_address"))
		description := strings.TrimSpace(r.FormValue("description"))

		processMail(email, issue, wallet, wallet_address, description, userId)

		w.Write([]byte(fmt.Sprintf("%v", requestNumber)))
		return
	}
}

func processMail(email string, issue string, wallet string, wallet_address string, description string, userId uint64) int {
	var tmp = make(map[uint64]RecivedData)

	requestNumber++

	UserRequestNumber.mux.Lock()
	UserRequestNumber.data[userId] = requestNumber
	UserRequestNumber.mux.Unlock()
	tmp[userId] = RecivedData{
		email:email,
		issue:issue,
		wallet:wallet,
		wallet_address:wallet_address,
		description: description,
		userId: userId,
	}

	Mdata.mux.Lock()
	Mdata.data = tmp
	Mdata.mux.Unlock()

	return 1
}

func createMail(mData RecivedData, mailType string) (string, string) {
	mailText := ""
	//set mail texts

	mailText = replaceText(mData, mailType)
	subject := fmt.Sprintf("Support Request: %v", mData.issue)

	//set Headers
	headers := make(map[string]string)
	headers["From"] = LocalConfig.SendGridMailFrom
	if mailType == "ToSupport" {
		headers["To"] = LocalConfig.SupportMail
	} else {
		headers["To"] = mData.email
	}

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
function send_mail
args: message
function create connection to mail service and after that function send message
*/
func send_mail_to_support(mData map[uint64]RecivedData) {
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

	for _, data := range mDataCMP {

		var w io.WriteCloser

		_,message := createMail(data, "ToSupport")

		if err = connection.Mail(LocalConfig.SendGridMailFrom); err != nil {
			fmt.Println(err)
			return
		}

		if err = connection.Rcpt(LocalConfig.SupportMail); err != nil {
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
		//delete(mData, key)
		time.Sleep(250*time.Millisecond)
	}
	connection.Quit()
}

/*
function send_mail
args: message
function create connection to mail service and after that function send message
*/
func send_mail_to_user(mData map[uint64]RecivedData) {
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

		email,message := createMail(data, "ToUser")

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

/*
function replaceText
args: mail data, current time
function for replacing HTML template text by mail data
*/
func replaceText(mData RecivedData, mailType string) string{
	userId := fmt.Sprintf("%v", mData.userId)
	Urequest := fmt.Sprintf("%v", UserRequestNumber.data[mData.userId])

	template := loadMailTemplate(mailType)
	template = strings.Replace(template, "{{email}}", mData.email, -2)
	template = strings.Replace(template, "{{issue}}", mData.issue, -2)
	template = strings.Replace(template, "{{userId}}", userId, -2)
	template = strings.Replace(template, "{{description}}", mData.description, -2)
	template = strings.Replace(template, "{{wallet}}", mData.wallet, -2)
	template = strings.Replace(template, "{{wallet_address}}", mData.wallet_address, -2)
	template = strings.Replace(template, "{{Urequest}}", Urequest, -2)

	return template
}

/*
function loadMailTemplate
args: path to HTML template
fuction for loading HTML template
*/
func loadMailTemplate(file string) string{
	path := fmt.Sprintf("../mailTemplates/%v.html", file)
	template, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	return string(template)
}

func saveRequests() {
	file, err := os.OpenFile("../SupportRequests/countRequests.db", os.O_CREATE|os.O_WRONLY, 0666)
	defer file.Close()

	if err != nil {
		log.Fatal(err)
	}

	file.Truncate(0)
	file.Seek(0,0)

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, requestNumber)
	bufferedWriter := bufio.NewWriter(file)
	_, err = bufferedWriter.Write(b, )

	if err != nil {
		log.Fatal(err)
	}

	bufferedWriter.Flush()
	file.Close()
}

func loadRequests() {
	file, err := os.Open("../SupportRequests/countRequests.db")
	defer file.Close()

	if err != nil {
		log.Fatal(err)
	}

	fi, err := file.Stat()

	if err != nil {
		log.Fatal("Could not obtain stat", err)
	}

	if fi.Size() < 8 {
		requestNumber = 0
		file.Close()
		return
	}

	data := readNextBytes(file, 8)

	if len(data) < 8 {
		requestNumber = 0
	} else {
		requestNumber = binary.LittleEndian.Uint64(data[0:8])
	}

	file.Close()
}

/*
function readNextBytes
function return N bytes from file
*/
func readNextBytes(file *os.File, number int) []byte {
	bytes := make([]byte, number)
	size, _ := file.Read(bytes)

	if (size != number) {
		fmt.Println("size does not fit")
		return []byte{}
	}

	return bytes
}