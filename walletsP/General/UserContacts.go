package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"wallets/walletsP/LocalConfig"
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var rcvWalletsAuth *zmq.Socket //ZMQ, socket variable //handle user login

var Wclients struct {
	data map[string]Session
	mux sync.Mutex // mutex for concurrent map
}

var WclientsId struct{
	data map[uint64]Session
	mux sync.Mutex // mutex for concurrent map
}

var UserContacts struct {
	data map[uint64]ContactStruct
	mux sync.Mutex // mutex for concurrent map
}

var tmpImgData struct {
	data map[int]struct{
		userID uint64
		name string
		ext string
	}
	mux sync.Mutex // mutex for concurrent map
}

type Session struct{
	userId uint64 //make 10 words??
	userSession string //TODO protect from oversize
	time int64
}

type ContactStruct map[int]struct{
	UserID uint64
	UserFIO string
	About string
	CoinId uint64
	WalletAddr string
	ImgUrl string
}

const uploadPath = "../WHTML/ConatctImages" //path to upload folder

func main() {
	// make maps with without nil data
	WclientsId.data = make(map[uint64]Session)
	UserContacts.data = make(map[uint64]ContactStruct)
	Wclients.data = make(map[string]Session)
	tmpImgData.data = make(map[int]struct{
			userID uint64
			name string
			ext string
		})
	//----
	fmt.Println("Starting User Contacts Server")
	loadContact()
	//create HTTP handle server
	Wservers := http.NewServeMux()
	Wservers.HandleFunc("/", handleWconn)
	port := LocalConfig.UcontactsHandlePort
	server_port := fmt.Sprintf(":%v", port)

	go func() {
		http.ListenAndServe(server_port, Wservers)
	}()

	rcvWalletsAuth, _ = zmq.NewSocket(zmq.PULL) //create new socket connection fo recive user auth...
	defer rcvWalletsAuth.Close()
	rcvWalletsAuth.SetRcvhwm(1100000)
	rcvWalletsAuthAddress := fmt.Sprintf("%v%v", LocalConfig.LocalTcpIpAddr, LocalConfig.UContactsAuthPort)
	rcvWalletsAuth.Connect(rcvWalletsAuthAddress)

	go checkWalletsAuth()

	//create file server for uploading images
	http.HandleFunc("/upload",  uploadFileHandler())
	fs := http.FileServer(http.Dir(uploadPath))
	http.Handle("/files/", http.StripPrefix("/files", fs))
	http.ListenAndServe(LocalConfig.UcontactsUploadPort, nil)
}

/*
func saveContact
args: user ID, user Full name, text about user, coin ID, Wallet address
function saved information about user // create new contact
 */
func saveContact(UserID uint64, UserFIO string, About string, CoinId uint64, WalletAddr string) string{
	file, err := os.OpenFile("../Contacts/Contacts.db", os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()
	//UserID uint64, UserFIO string, About string, CoinId uint64, WalletAddr string
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, UserID)
	bufferedWriter := bufio.NewWriter(file)
	_, err = bufferedWriter.Write(b,)
	if err != nil {
		log.Fatal("writeFileWallet 229",err)
	}
	bufferedWriter.Flush()
	b = make([]byte, 8)
	binary.LittleEndian.PutUint64(b, CoinId)
	bufferedWriter = bufio.NewWriter(file)
	_, err = bufferedWriter.Write(b,)
	if err != nil {
		log.Fatal("writeFileWallet 229",err)
	}
	bufferedWriter.Flush()
	b = make([]byte, 200)
	copy(b[:], []byte(UserFIO))
	_, err = bufferedWriter.Write( b, )
	if err != nil {
		log.Fatal("writeFile 110", err)
	}
	bufferedWriter.Flush()
	b = make([]byte, 200)
	copy(b[:], []byte(About))
	_, err = bufferedWriter.Write( b, )
	if err != nil {
		log.Fatal("writeFile 110", err)
	}
	bufferedWriter.Flush()
	b = make([]byte, 42)
	copy(b[:], []byte(WalletAddr))
	_, err = bufferedWriter.Write( b, )
	if err != nil {
		log.Fatal("writeFile 110", err)
	}
	bufferedWriter.Flush()

	UserContacts.mux.Lock()
	//create variable for TMP data
	var tmpData = make(map[int]struct{
		UserID uint64
		UserFIO string
		About string
		CoinId uint64
		WalletAddr string
		ImgUrl string
	})
	var tmp struct{
		UserID uint64
		UserFIO string
		About string
		CoinId uint64
		WalletAddr string
		ImgUrl string
	}
	tmp.UserID = UserID
	tmp.CoinId = CoinId
	tmp.About = About
	tmp.UserFIO = UserFIO
	tmp.WalletAddr = WalletAddr
	tmpImgData.mux.Lock()
	var ij int
	ij = 0
	for _, imgData := range tmpImgData.data {
		if UserID == imgData.userID {
			tmp.ImgUrl = fmt.Sprintf("%v%v", imgData.name, imgData.ext)
		}
		ij++
	}
	delete(tmpImgData.data, ij)
	tmpImgData.mux.Unlock()

	b = make([]byte, 42)
	copy(b[:], []byte(tmp.ImgUrl))
	_, err = bufferedWriter.Write( b, )
	if err != nil {
		log.Fatal("writeFile 110", err)
	}
	bufferedWriter.Flush()

	var i int
	if len(UserContacts.data[UserID]) > 0 {
		i = len(UserContacts.data[UserID])
		tmpData = UserContacts.data[UserID]
	} else {
		i = 0
	}
	tmpData[i] = tmp
	UserContacts.data[UserID] = tmpData
	UserContacts.mux.Unlock()

	file.Close()
	return tmp.ImgUrl
}

/*
function loadContact
function loaded user contacts from file
 */
func loadContact() {
	file, err := os.Open("../Contacts/Contacts.db")
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		log.Fatal("Could not obtain stat", err)
	}

	var tmpData = make(map[int]struct{
		UserID uint64
		UserFIO string
		About string
		CoinId uint64
		WalletAddr string
		ImgUrl string
	})
	var tmp struct{
		UserID uint64
		UserFIO string
		About string
		CoinId uint64
		WalletAddr string
		ImgUrl string
	}
	var i int
	var j int
	j = 1
	for i = 0; i < int(fi.Size()/(500)); i++ {

		data := readNextBytes(file, 500)

		UserId := binary.LittleEndian.Uint64(data[0:8])
		CoinID :=  binary.LittleEndian.Uint64(data[8:16])

		UserFIO := string(bytes.Trim(data[16:216], "\x00"))
		About := string(bytes.Trim(data[216:416], "\x00"))
		WalletAddr := string(bytes.Trim(data[416:458], "\x00"))
		ImgUrl := string(bytes.Trim(data[458:500], "\x00"))
		tmp.CoinId = CoinID
		tmp.About = About
		tmp.UserID = UserId
		tmp.UserFIO = UserFIO
		tmp.WalletAddr = WalletAddr
		tmp.ImgUrl = ImgUrl
		tmpData[i] = tmp

		UserContacts.data[UserId] = tmpData
		j++
	}

}

/*
function uploadFileHandler
handler for uploading images
 */
func uploadFileHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Connection", "close")
		file, _, err := r.FormFile("uploadFile")
		var tmp struct{
			userID uint64
			name string
			ext string
		}
		data := r.FormValue("data")
		userId, _ := strconv.ParseUint(data, 10, 0)
		fmt.Println(userId)
		if err != nil {
			fmt.Println("Error: ", err)
			w.Write([]byte("Error"))
			return
		}
		if file == nil {
			w.Write([]byte("Error"))
			return
		}
		defer file.Close()
		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			fmt.Println("Error: ", err)
			w.Write([]byte("Error"))
			return
		}
		detectedFileType := http.DetectContentType(fileBytes)
		switch detectedFileType {
			case "image/jpeg", "image/jpg":
			case "image/gif", "image/png":
			break
		default:
			fmt.Println("Error: ", err)
			w.Write([]byte("Error"))
			return
		}

		fileName := randToken(12) //rename files
		fileEndings, err := mime.ExtensionsByType(detectedFileType)
		if err != nil {
			fmt.Println("Error: ", err)
			w.Write([]byte("Error"))
			return
		}

		ext := fileEndings[0]
		newPath := filepath.Join(uploadPath, fileName+fileEndings[0])
		newFile, err := os.Create(newPath)
		if err != nil {
			fmt.Println("Error: ", err)
			w.Write([]byte("Error"))
			return
		}
		defer newFile.Close()
		if _, err := newFile.Write(fileBytes); err != nil || newFile.Close() != nil {
			fmt.Println("Error: ", err)
			w.Write([]byte("Error"))
			return
		}

		tmpImgData.mux.Lock()
		tmp.userID = userId
		tmp.name = fileName
		tmp.ext = ext
		i := len(tmpImgData.data) + 1
		tmpImgData.data[i] = tmp
		tmpImgData.mux.Unlock()
		w.Write([]byte("Success"))
		return
	})
}

/*
function handleWconn
handler for http requests
 */
func handleWconn(w http.ResponseWriter, r *http.Request)  {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Connection", "close")

	cmd := r.FormValue("data")
	data := strings.Split(cmd, ",")

	if (len(data) < 6) {
		w.Write([]byte("nFAIL182"))
		return //wrong request
	}
	/*
		request struct
		data[0] - type of request //get balance, create wallet, send money
		data[1] - user session
		data[2] - user id
		data[3-9] - request's params
	*/
	userId, _ := strconv.ParseUint(data[2], 10, 0)

	if (Wclients.data[data[1]].userId != userId){
		w.Write([]byte("FAIL197"))
		return //wrong request
	}

	if (data[0] == "addNewContact") {
		UserFIO := data[3]
		About := data[4]
		CoinId , _ := strconv.ParseUint(data[5], 10, 64)
		WalletAddr := data[6]
		imgUrl := saveContact(userId, UserFIO, About, CoinId, WalletAddr)
		response := fmt.Sprintf("%v,%v,%v,%v,%v", UserFIO, About, CoinId, WalletAddr,imgUrl)
		//-------
		w.Write([]byte(response))
		return
	}
	if (data[0] == "DisplayAllContacts") {
		var walletAddress string
		dataOfImg, is_ok := UserContacts.data[userId]
		if is_ok {
			for _, Imgdata := range dataOfImg {
				walletAddress = fmt.Sprintf("%v,%v,%v,%v,%v|%v", Imgdata.UserFIO, Imgdata.About, Imgdata.CoinId, Imgdata.WalletAddr, Imgdata.ImgUrl, walletAddress)
			}
		} else {
			walletAddress = ""
		}

		w.Write([]byte(walletAddress))
		return
	}
}

/*
function checkWalletsAuth
listening socket for receive user authorization.
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
		fmt.Println("Received Auth", string(msg))
		s := strings.Split(string(msg), ",")

		if len(s) < 3 {
			//wrong request
			continue
		}
		if (s[2] != "" && s[1] != "") { //handle login
			userId, _ = strconv.ParseUint(s[2], 10, 64)
			userSession = s[1]
			Wclients.mux.Lock()
			Wclients.data[userSession] = Session{userId, userSession, time.Now().Unix()}
			Wclients.mux.Unlock()

			WclientsId.mux.Lock()
			WclientsId.data[userId] = Session{userId, userSession, time.Now().Unix()}
			WclientsId.mux.Unlock()
		} else if(s[0] != "" && s[2] != "") { // handle logout
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
function return N bytes from file
*/
func readNextBytes(file *os.File, number int) []byte {
	bytes := make([]byte, number)

	size, err := file.Read(bytes)
	if err != nil {
		log.Fatal("467",err, size)
	}
	if (size != number){
		fmt.Println("size does not fit")
		return []byte{}
	}
	return bytes
}

/*
function randToken
function for generating random file name
 */
func randToken(len int) string {
	b := make([]byte, len)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}