package main

import (
	"bytes"
	"encoding/json"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"io/ioutil"
	"regexp"
	"strconv"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"math/rand"
	"net"
	"bufio"
	"time"
	"os"
	"encoding/binary"
	"log"
	"net/http"
	"sync"
	"wallets/walletsP/LocalConfig"
)
 
var SendToServers = make(map[int]*zmq.Socket) //ZMQ, socket variable  // send auth message
var mailAUTH *zmq.Socket //ZMQ, socket variable // send message to the mail server

var lastId uint64
var realIP string

var TMPMailData struct {
	data []byte
	mux sync.Mutex
}

var logIN_users struct {
	data map[uint64]string
	mux sync.Mutex //mutex for the concurrent map
}

//TODO converting to a binary number map key
var users struct {
	data map[string] User
	mux sync.Mutex // mutex for the concurrent map
}

var LoginRedirect struct {
	data map[string]string
	mux sync.Mutex // mutex for the concurrent map
}

var UsersLoginIn struct {
	data map[uint64]string
	mux sync.Mutex // mutex for the concurrent map
}

var WaitingForGoogle struct {
	data map[string] User
	mux sync.Mutex // mutex for the concurrent map
}

var ErrorTryingLogin struct{
	data map[uint64]ErrorTrying
	mux sync.Mutex // mutex for the concurrent map
}

var BannedUsers struct {
	data map[uint64]string
	mux sync.Mutex // mutex for the concurrent map
}

var sendAuth sync.Mutex
var GetFreeWalletMutex sync.Mutex

var mailChennel  *MailChennel
var google_chennel *GoogleChennel

type MailChennel struct{
	sendToUser chan *Mail_data
}

type GoogleChennel struct{
	HandleGoogle chan *GoogleAuth
}

type ErrorTrying struct {
	countTrys uint64
	Email string
}

type UserRegData struct{
	user *User
	coin string
	address string
}

type User struct {
	Id uint64
	LastLogin uint64
	Session string
	IpAddr string
	Email string
	Key string
	UserAgent []string
}

type Mail_data struct {
	Id uint64
	LastLogin uint64
	Session string
	IpAddr string
	Email string
	password string
	mail_type string
}

type Users struct{
	ChWritefile chan *User
}

type GoogleAuth struct{
	Email string
	Verified_email bool
}

type googleResponse struct {
	Id string	`json:"id"`
	Email string	`json:"email"`
	Verified_email bool	`json:"verified_email"`
	Picture string	`json:"picture"`
}

var (
	googleOauthConfig *oauth2.Config
	// TODO: randomize it
	oauthStateString = "pseudo-random"
)
//Google console API Key
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)
/*
function init
init google AUTH data
 */
func init() {
	googleOauthConfig = &oauth2.Config{
		RedirectURL:  "",
		ClientID:     "",
		ClientSecret: "",
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}
}

/*
read more in google Documentation
https://godoc.org/golang.org/x/oauth2
https://cloud.google.com/go/getting-started/authenticate-users-with-iap
 */
func handleGoogleLogin(w http.ResponseWriter, r *http.Request) string {
	url := googleOauthConfig.AuthCodeURL(oauthStateString)
	return url
}

/*
func handleGoogleCallback
handle google auth answer
*/
func handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	content, err := getUserInfo(r.FormValue("state"), r.FormValue("code"))
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	result := googleResponse{}
	content2 := string(content[:])
	space := regexp.MustCompile(`\s+`)
	content2 = space.ReplaceAllString(content2, "")

	err = json.Unmarshal([]byte(content2), &result)
	if err != nil {
		fmt.Println(err.Error())
	}

	if result.Verified_email == true { //check result
		WaitingForGoogle.mux.Lock()
		user, is_ok := WaitingForGoogle.data[result.Email] // search user, who is waiting for google response
		WaitingForGoogle.mux.Unlock()
		var UserAgent []string
		if is_ok{
			WaitingForGoogle.mux.Lock()
			UserAgent = WaitingForGoogle.data[result.Email].UserAgent
			delete(WaitingForGoogle.data, result.Email) //delete user from the map
			WaitingForGoogle.mux.Unlock()

			users.mux.Lock()
			checkUser, mail_ok := users.data[result.Email]; // if isset user
			users.mux.Unlock()

			if mail_ok{ // create new session
				checkUser.Session = RandStringBytesMaskImprSrc(25)
				checkUser.LastLogin = uint64(time.Now().UnixNano())
				checkUser.IpAddr = realIP
				checkUser.UserAgent = UserAgent

				users.mux.Lock()
				users.data[result.Email] = checkUser
				users.mux.Unlock()

				logIN_users.mux.Lock()
				logIN_users.data[user.Id] = result.Email
				logIN_users.mux.Unlock()

				//send login mail
				mailChennel.sendToUser <- &Mail_data{user.Id, user.LastLogin, user.Session, realIP, user.Email, "", "login"}
				w = handleGoogleAuth(result.Email, w)

				UsersLoginIn.mux.Lock()
				authIn := UsersLoginIn.data[user.Id] // redirect to...
				UsersLoginIn.mux.Unlock()

				LoginRedirect.mux.Lock()
				url := LoginRedirect.data[authIn]
				LoginRedirect.mux.Unlock()

				//url := "http://127.0.0.1:8001/"
				http.Redirect(w, r, url, http.StatusSeeOther)
				return
			} else {
				//handle error
				ErrorTryingLogin.mux.Lock()
				ErrorTryingLogin.data[user.Id] = ErrorTrying{ countTrys: (ErrorTryingLogin.data[user.Id].countTrys+1), Email:result.Email}
				ErrorTryingLogin.mux.Unlock()

				code := checkUserForBan() // add 1 login mistake // 3 mistakes - ban

				WaitingForGoogle.mux.Lock()
				delete(WaitingForGoogle.data, result.Email)
				WaitingForGoogle.mux.Unlock()

				url := fmt.Sprintf("http://127.0.0.1:8001/?error=%v", code)
				http.Redirect(w, r, url, http.StatusSeeOther)
				return
			}
		} else {
			//handle error
			ErrorTryingLogin.mux.Lock()
			ErrorTryingLogin.data[user.Id] = ErrorTrying{ countTrys: (ErrorTryingLogin.data[user.Id].countTrys+1), Email:result.Email}
			ErrorTryingLogin.mux.Unlock()

			code := checkUserForBan()

			WaitingForGoogle.mux.Lock()
			delete(WaitingForGoogle.data, result.Email)
			WaitingForGoogle.mux.Unlock()

			//error redirect
			url := fmt.Sprintf("http://127.0.0.1:8001/?error=%v", code)
			http.Redirect(w, r, url, http.StatusSeeOther)
			return
		}
	}
}

/*
func handleGoogleAuth
function sets  cookie and says to other servers about user login
 */
func handleGoogleAuth(Email string, w http.ResponseWriter) http.ResponseWriter {
	users.mux.Lock()
	user, mail_ok := users.data[Email];
	users.mux.Unlock()

	if mail_ok {
		user.Session = RandStringBytesMaskImprSrc(25)
		user.LastLogin = uint64(time.Now().UnixNano())
		user.IpAddr = realIP

		logIN_users.mux.Lock()
		logIN_users.data[user.Id] = Email
		logIN_users.mux.Unlock()

		//send email about login
		mailChennel.sendToUser <- &Mail_data{user.Id, user.LastLogin, user.Session, realIP, user.Email, "", "login"}

		sendAuth.Lock()
		for key,_ := range SendToServers{
			SendToServers[key].Send(fmt.Sprintf("l,%v,%v,%v", user.Session, user.Id, user.Email), 0)
		}
		sendAuth.Unlock()
		expire := time.Now().AddDate(0, 0, 1)
		userId := strconv.FormatUint(user.Id, 10)

		UsersLoginIn.mux.Lock()
		authIn := UsersLoginIn.data[user.Id]
		UsersLoginIn.mux.Unlock()

		if authIn == "wallet" {
			cookie := http.Cookie{
				Name:    "wallets_u_s",
				Value:   user.Session,
				Expires: expire,
				Path: "/",
			}
			http.SetCookie(w, &cookie)

			cookie = http.Cookie{
				Name:    "wallets_u_id",
				Value:   userId,
				Expires: expire,
				Path: "/",
			}
			http.SetCookie(w, &cookie)
			cookie = http.Cookie{
				Name:    "wallets_u_log",
				Value:   user.Email,
				Expires: expire,
				Path: "/",
			}
			http.SetCookie(w, &cookie)
		} else {
			cookie := http.Cookie{
				Name:    "exc_u_s",
				Value:   user.Session,
				Expires: expire,
				Path: "/",
			}
			http.SetCookie(w, &cookie)

			cookie = http.Cookie{
				Name:    "exc_u_id",
				Value:   userId,
				Expires: expire,
				Path: "/",
			}
			http.SetCookie(w, &cookie)
			cookie = http.Cookie{
				Name:    "exc_u_log",
				Value:   user.Email,
				Expires: expire,
				Path: "/",
			}
			http.SetCookie(w, &cookie)
		}
		return w
	}
	return w
}

/*
read more in google Documentation
https://godoc.org/golang.org/x/oauth2
https://cloud.google.com/go/getting-started/authenticate-users-with-iap
*/
func getUserInfo(state string, code string) ([]byte, error) {
	if state != oauthStateString {
		return nil, fmt.Errorf("invalid oauth state")
	}

	token, err := googleOauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %s", err.Error())
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}

	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading response body: %s", err.Error())
	}

	return contents, nil
}

/*
func listen_mail
function for handling mail Channel and sending data to mail server
 */
func listen_mail() {
	for {
		buff := <- mailChennel.sendToUser
		mailbuf := make([]byte, 350)

		var Mtype uint64

		if buff.mail_type == "login" {
			Mtype = 2
		}
		if buff.mail_type == "registartion" {
			Mtype = 1
		}
		if buff.mail_type == "logout" {
			Mtype = 3
		}

		users.mux.Lock()
		user, mail_ok := users.data[buff.Email];
		users.mux.Unlock()

		var UserAgent string
		if mail_ok {
			if len(user.UserAgent) > 0 {
				UserAgent = user.UserAgent[0]
			} else {
				UserAgent = ""
			}

		}
		binary.LittleEndian.PutUint64(mailbuf[0:8], Mtype)
		copy(mailbuf[8:58], []byte(buff.Email))
		copy(mailbuf[58:100], []byte(buff.IpAddr))
		copy(mailbuf[100:200], []byte(buff.password))
		copy(mailbuf[200:350], []byte(UserAgent))

		TMPMailData.mux.Lock()
		if len(TMPMailData.data) <= 10 {
			TMPMailData.data = mailbuf
		} else {
			TMPMailData.data = append(TMPMailData.data, mailbuf...)
		}
		TMPMailData.mux.Unlock()

		//mailAUTH.SendBytes(mailbuf, 0)
	}
}

/*
function get_rial_ip
get real User IP addr
 */
func get_rial_ip(w http.ResponseWriter, r *http.Request) string {
	realIP = r.Header.Get("X-Real-Ip")
	real2IP := r.Header.Get("X-Forwarded-For")
	real3IP,_,_ := net.SplitHostPort(r.RemoteAddr)
	if len(realIP) == 0 {
		if len(real2IP) == 0 {
			realIP = real3IP
		}else{
			realIP = real2IP
			if real2IP != real3IP {
				//...
			}
		}
	}
	return realIP
}
/*
function clear_mail
check user mail. Return True or False
 */
func clear_mail(email string) bool{
	expression := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	result := expression.MatchString(email)
	return result
}

/*
function handleLogout
handle user loguot
delete user data from maps
 */
func handleLogout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Connection", "close")

	switch r.Method {
	case "GET":
		http.Error(w, "404 not found.", http.StatusNotFound)
	case "POST":
		log_userId := r.FormValue("userId")

		userId, _ := strconv.ParseUint(log_userId, 10, 64)

		auth := logIN_users.data[userId]

		users.mux.Lock()
		user, ok := users.data[auth];
		users.mux.Unlock()
		if ok {
			user.Session = ""

			users.mux.Lock()
			users.data[auth] = user
			users.mux.Unlock()

		}else{
			//TODO write error logOut log
			fmt.Fprintf(w, "%v", "-1")
		}
		sendAuth.Lock()
		for key,_ := range SendToServers{
			SendToServers[key].Send(fmt.Sprintf("l,%v,%v", user.Session, user.Id),0)
		}
		sendAuth.Unlock()

		logIN_users.mux.Lock()
		delete(logIN_users.data, userId)
		logIN_users.mux.Unlock()

		fmt.Fprintf(w, "%v", "1")

		mailChennel.sendToUser <-  &Mail_data{user.Id,user.LastLogin,user.Session,realIP, user.Email, "","logout"}
		return
	}
}

/*
func autoLogOut
Delete user data from maps
 */
func autoLogOut() {
	twoHours := uint64(2*time.Hour)
	for {
		logIN_users.mux.Lock()
		for userID, mail := range logIN_users.data {
			currentTime := uint64(time.Now().UnixNano())

			users.mux.Lock()
			user, ok := users.data[mail];
			users.mux.Unlock()

			if ok {
				if (currentTime - user.LastLogin) >= twoHours { // 7200 -  2 hours
					user.Session = ""

					users.mux.Lock()
					users.data[mail] = user
					users.mux.Unlock()
					sendAuth.Lock()
					for key,_ := range SendToServers{
						SendToServers[key].Send(fmt.Sprintf("l,%v,%v", user.Session, user.Id),0)
					}
					sendAuth.Unlock()
					delete(logIN_users.data, userID)
					mailChennel.sendToUser <-  &Mail_data{user.Id,user.LastLogin,user.Session,"", user.Email, "","logout"}
				}
			}
		}
		logIN_users.mux.Unlock()
		time.Sleep(10*time.Second)
	}
}

/*
func handleAuth
Handle user Auth
 */
func handleAuth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Connection", "close")
	var loginType string

	switch r.Method {
		case "GET":
			http.Error(w, "404 not found.", http.StatusNotFound)
		case "POST":
			cmd := r.FormValue("cmd")
			if (cmd == "add_funds"){
				userId := r.FormValue("user")
				value := r.FormValue("add")
				coin := r.FormValue("coin")
				//TODO проверки на баланс пользователя
				sendAuth.Lock()
				for key,_ := range SendToServers{
					SendToServers[key].Send(fmt.Sprintf("a,%v,%v,%v", value, userId, coin),0)
				}
				sendAuth.Unlock()
				fmt.Fprintf(w, "%v", "1")
				return
			}
			if (cmd == "remove_funds"){
				userId := r.FormValue("user")
				value := r.FormValue("remove")
				coin := r.FormValue("coin")
				//TODO проверки на баланс пользователя
				sendAuth.Lock()
				for key,_ := range SendToServers{
					SendToServers[key].Send(fmt.Sprintf("r,%v,%v,%v", value, userId, coin),0)
				}
				sendAuth.Unlock()
				fmt.Fprintf(w, "%v", "1")
				return
			}
			// read data from Json Vars
			log_userId := r.FormValue("userId")
			log_session := r.FormValue("session_c")

			var user User
			var ok bool
			var mail_ok bool
			var auth string
			var email string
			var UserAgent []string
			UserAgent = r.Header["User-Agent"]

			//check for logged users // if user logged and has session
			if len(log_userId) > 0 && len(log_session) > 5 {
				userId, _ := strconv.ParseUint(log_userId, 10, 64)
				auth := logIN_users.data[userId]

				users.mux.Lock()
				user, ok = users.data[auth];
				users.mux.Unlock()
				if ok {
					fmt.Println("session user")
				} else {
					ErrorTryingLogin.mux.Lock()
					ErrorTryingLogin.data[user.Id] = ErrorTrying{ countTrys: (ErrorTryingLogin.data[user.Id].countTrys+1), Email:user.Email}
					ErrorTryingLogin.mux.Unlock()

					code := checkUserForBan()
					fmt.Fprintf(w, "%v", code)
				}
			} else {
				//New Login
				//TODO limit size of auth to 256  chars
				auth = r.FormValue("auth")
				email = r.FormValue("email")
				loginType = r.FormValue("AuthType")
				//TODO make one more auth level to add phone verification
				realIP = get_rial_ip(w, r)

				users.mux.Lock()
				user, mail_ok = users.data[email];
				users.mux.Unlock()

				if mail_ok {
					BannedUsers.mux.Lock()
					_, banned := BannedUsers.data[user.Id]
					BannedUsers.mux.Unlock()

					if banned {
						fmt.Fprintf(w, "%v", "-2") // User in Ban list
						return
					}
					if len(user.Session) > 0 { // bad session
						ErrorTryingLogin.mux.Lock()
						ErrorTryingLogin.data[user.Id] = ErrorTrying{ countTrys: (ErrorTryingLogin.data[user.Id].countTrys+1), Email:user.Email}
						ErrorTryingLogin.mux.Unlock()

						code := checkUserForBan()
						//TODO if last_login * 0.2 hour < current time
						fmt.Fprintf(w, "%v", code)
						return
					}
					if user.Key != auth { // Wrong password
						ErrorTryingLogin.mux.Lock()
						ErrorTryingLogin.data[user.Id] = ErrorTrying{ countTrys: (ErrorTryingLogin.data[user.Id].countTrys+1), Email:user.Email}
						ErrorTryingLogin.mux.Unlock()

						code := checkUserForBan()
						fmt.Fprintf(w, "%v", code)
						return
					}
				} else {
					//User Doesn't Exist
					ErrorTryingLogin.mux.Lock()
					ErrorTryingLogin.data[user.Id] = ErrorTrying{ countTrys: (ErrorTryingLogin.data[user.Id].countTrys+1), Email:user.Email}
					ErrorTryingLogin.mux.Unlock()

					code := checkUserForBan()
					fmt.Fprintf(w, "%v", code)
					return
				}
				google_auth := false // Enable or Disable google AUTH

				if google_auth {
					//add a parallel process to make backgrounds of function and pass data
					redirect_url := handleGoogleLogin(w, r)
					user.UserAgent = UserAgent

					WaitingForGoogle.mux.Lock()
					WaitingForGoogle.data[user.Email] = user
					WaitingForGoogle.mux.Unlock()

					UsersLoginIn.mux.Lock()
					UsersLoginIn.data[user.Id] = loginType
					UsersLoginIn.mux.Unlock()

					fmt.Fprintf(w, "%v %v", "googleAuth", redirect_url)
					//return
				} else {
					user.Session = RandStringBytesMaskImprSrc(25)
					user.LastLogin = uint64(time.Now().UnixNano())
					user.IpAddr = realIP
					user.UserAgent = UserAgent

					users.mux.Lock()
					users.data[email] = user
					users.mux.Unlock()

					logIN_users.mux.Lock()
					logIN_users.data[user.Id] = email
					logIN_users.mux.Unlock()

					mailChennel.sendToUser <- &Mail_data{user.Id, user.LastLogin, user.Session, realIP, user.Email, "", "login"}

					UsersLoginIn.mux.Lock()
					UsersLoginIn.data[user.Id] = loginType
					UsersLoginIn.mux.Unlock()

					sendAuth.Lock()
					for key,_ := range SendToServers{
						SendToServers[key].Send(fmt.Sprintf("l,%v,%v,%v", user.Session, user.Id, user.Email),0)
					}
					sendAuth.Unlock()

					GetFreeWalletMutex.Lock()
					fmt.Fprintf(w, "%v,%v,%v", user.Session, user.Id, "") //remove - add only if needed in the future
					GetFreeWalletMutex.Unlock()
				}
			}
		default:
			http.Error(w, "404 not found.", http.StatusNotFound)
	}
}
/*
func handleReg
register a new user
 */
func (usersStr *Users) handleReg(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Connection", "close")

	switch r.Method {
		case "GET":
			http.Error(w, "404 not found.", http.StatusNotFound)
		case "POST":
			//TODO limit size of auth to 256  chars
			login := r.FormValue("login")
			check_mail := clear_mail(login)
			if check_mail == false {
				fmt.Fprintf(w, "%v", "-1")
				return
			}
			Email := login
			password := r.FormValue("password")
			//TODO make one more auth level to add phone verification
			h := sha256.New()
			h.Write([]byte(fmt.Sprintf("%v %v", login, password)))
			key := hex.EncodeToString(h.Sum(nil))

			users.mux.Lock()
			lastId=lastId+1
			lastUserId:=lastId
			_, ok := users.data[login];
			users.mux.Unlock()

			var UserAgent []string
			UserAgent = r.Header["User-Agent"]
			if ok {
				fmt.Fprintf(w, "%v", "-1")
			} else {
				realIP = get_rial_ip(w, r)
				Session := RandStringBytesMaskImprSrc(25)
				user := User{lastUserId,0,"",realIP ,Email,key, UserAgent}
				user.Session = Session
				user.LastLogin = uint64(time.Now().UnixNano())

				fmt.Fprintf(w, "%v,%v,%v", user.Session, user.Id, "")
				for key,_ := range SendToServers{
					SendToServers[key].Send(fmt.Sprintf("n,%v,%v,%v", user.Session, user.Id, user.Email),0)
				}
				user.Key = key

				logIN_users.mux.Lock()
				logIN_users.data[lastUserId] = login
				logIN_users.mux.Unlock()

				users.mux.Lock()
				users.data[login] = user
				users.mux.Unlock()

				usersStr.ChWritefile <- &user
				mailChennel.sendToUser <-  &Mail_data{lastUserId, user.LastLogin, user.Session, realIP, Email, password,"registartion"}
			}
		default:
			http.Error(w, "404 not found.", http.StatusNotFound)
	}
}

func RandStringBytesMaskImprSrc(n int) string {
    var src = rand.NewSource(time.Now().UnixNano())
    b := make([]byte, n)
    // A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
    for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
        if remain == 0 {
            cache, remain = src.Int63(), letterIdxMax
        }
        if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
            b[i] = letterBytes[idx]
            i--
        }
        cache >>= letterIdxBits
        remain--
    }

    return string(b)
}

/*
function writeFile
Write user info to database
 */
func (usersStr *Users) writeFile() {

	file, err := os.OpenFile("../users/users.db", os.O_APPEND|os.O_WRONLY, 0666)
	defer file.Close()
	if err != nil {
		log.Fatal("writeFile 366",err)
	}

	for {
		usr, ok := <- usersStr.ChWritefile

		if ok==false{
			fmt.Println("WriteFileChannel is closed")
			return
		}

		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, usr.Id)
		bufferedWriter := bufio.NewWriter(file)

		_, err = bufferedWriter.Write( b, )
		if err != nil {
			log.Fatal("writeFile 391 ",err)
		}

		b = make([]byte, 50)
		binary.LittleEndian.PutUint64(b, usr.LastLogin)

		_, err = bufferedWriter.Write( b, )
		if err != nil {
			log.Fatal("writeFile 404", err)
		}
		//----
		b = make([]byte, 100)
		copy(b[:], []byte(usr.Email))
		_, err = bufferedWriter.Write( b, )
		if err != nil {
			log.Fatal("writeFile 404", err)
		}
		b = make([]byte, 20)
		copy(b[:], []byte(usr.IpAddr))
		_, err = bufferedWriter.Write( b, )
		if err != nil {
			log.Fatal("writeFile 404", err)
		}
		b = make([]byte, 30)
		_, err = bufferedWriter.Write( b, )
		if err != nil {
			log.Fatal("writeFile 404", err)
		}
		b = make([]byte, 100)
		copy(b[:], []byte(usr.Key))
		_, err = bufferedWriter.Write( b, )
		if err != nil {
			log.Fatal("writeFile 404", err)
		}
		//----
		_, err = bufferedWriter.Write([]byte(usr.Key), )

		bufferedWriter.Flush()
	}
	file.Close()

}

/*
function loadUsers
Load users from database
User id > 2
 */
func loadUsers(database_file string) {
// TODO load user login and connect it to usersdb (store id/user separately)
	var UsrlasTId uint64
	UsrlasTId = 2
	if database_file != "users_test.db" {
		database_file = "../users/users.db"
	}
	file, err := os.Open(database_file)

	defer file.Close()

	if err != nil {
		log.Fatal("loadUsers 474", err)
	}

	fi, err := file.Stat()
	if err != nil {
		// Could not obtain stat, handle error
		log.Fatal("Could not obtain stat", err)
	}

	var i uint64

	for i = 0; i < uint64(fi.Size()/(64+308)); i++ {
		var offset = int(i) + 1
		if offset == 0 {
			offset = 1
		}
		data := readNextBytes(file, (64+308))
		users.mux.Lock()

		users.data[string(bytes.Trim(data[58:158], "\x00"))]=User{
			Id:binary.LittleEndian.Uint64(data[0:8]),
			LastLogin:binary.LittleEndian.Uint64(data[8:58]),
			Email:string(bytes.Trim(data[58:158], "\x00")),
			IpAddr:string(bytes.Trim(data[158:178], "\x00")),
			Session:string(bytes.Trim(data[178:208], "\x00")),
			Key:string(bytes.Trim(data[208:308], "\x00")),
		}
		users.mux.Unlock()
		if UsrlasTId < binary.LittleEndian.Uint64(data[0:8]) {
			UsrlasTId = binary.LittleEndian.Uint64(data[0:8])
		}
	}

	lastId=UsrlasTId
	if lastId == 0 || lastId == 1 || lastId == 2 {
		lastId = 2
	}
}

func readNextBytes(file *os.File, number int) []byte {
	bytes := make([]byte, number)

	_, err := file.Read(bytes)
	if err != nil {
		log.Fatal("readNextBytes 531", err)
	}

	return bytes
}
/*
function checkUserForBan
count wrong login tries and add user to ban list
 */
func checkUserForBan() string{
	var code string
	code = "-1"
	ErrorTryingLogin.mux.Lock()
	for _, data := range ErrorTryingLogin.data {
		if data.countTrys >= 3 {
			code = "-2" //user Banned

			users.mux.Lock()
			user, mail_ok := users.data[data.Email];
			users.mux.Unlock()

			if mail_ok {
				BannedUsers.mux.Lock()
				BannedUsers.data[user.Id] = user.Email
				BannedUsers.mux.Unlock()

				addBann(user.Id, user.Email)
			}

		}
	}
	ErrorTryingLogin.mux.Unlock()
	return code
}
/*
function addBann
write user to bans database
 */
func addBann(usrId uint64, email string) {
	file, err := os.OpenFile("../users/bans.db", os.O_APPEND|os.O_WRONLY, 0666)
	defer file.Close()
	if err != nil {
		log.Fatal("writeFile 366",err)
	}

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, usrId)
	bufferedWriter := bufio.NewWriter(file)
	_, err = bufferedWriter.Write( b, )
	if err != nil {
		log.Fatal("writeFile 1019 ",err)
	}

	b = make([]byte, 100)
	copy(b[:], []byte(email))
	_, err = bufferedWriter.Write( b, )
	if err != nil {
		log.Fatal("writeFile 1026", err)
	}

	bufferedWriter.Flush()
	file.Close()
}

/*
function updateBanns
Delete user from Ban list
 */
func updateBanns(userID uint64) {
	file, err := os.OpenFile("../users/bans.db", os.O_CREATE|os.O_WRONLY, 0666)
	defer file.Close()
	if err != nil {
		log.Fatal("read file bans 1072",err)
	}
	file.Truncate(0)
	file.Seek(0,0)

	BannedUsers.mux.Lock()
	_, is_ok := BannedUsers.data[userID]
	if is_ok{
		delete(BannedUsers.data, userID)
	}

	for usrId, email := range BannedUsers.data {
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, usrId)
		bufferedWriter := bufio.NewWriter(file)
		_, err = bufferedWriter.Write( b, )
		if err != nil {
			log.Fatal("writeFile 1019 ",err)
		}

		b = make([]byte, 100)
		copy(b[:], []byte(email))
		_, err = bufferedWriter.Write( b, )
		if err != nil {
			log.Fatal("writeFile 1026", err)
		}

		bufferedWriter.Flush()
	}
	BannedUsers.mux.Unlock()

	file.Close()
}

/*
function loadBannedUsers
load ban list from Database
 */
func loadBannedUsers() {
	file, err := os.Open("../users/bans.db")
	defer file.Close()
	if err != nil {
		log.Fatal("loadUserBANS",err)
	}

	fi, err := file.Stat()
	if err != nil {
		log.Fatal("Could not obtain stat", err)
	}

	var i uint64
	BannedUsers.mux.Lock()
	for i = 0; i < uint64(fi.Size()/(108)); i++ {
		data := readNextBytes(file, (108))
		BannedUsers.data[binary.LittleEndian.Uint64(data[0:8])] = string(bytes.Trim(data[8:108], "\x00"))
	}
	BannedUsers.mux.Unlock()
}

func sendToMailServer() {
	for {
		TMPMailData.mux.Lock()
	//	fmt.Println(TMPMailData.data)
		if len(TMPMailData.data) >= 350 {
			mailAUTH.SendBytes(TMPMailData.data, 0)
			TMPMailData.data = []byte{}
		}
		TMPMailData.mux.Unlock()

		time.Sleep(1*time.Minute)
	}
}

func main() {
	// make maps with no nil data
	logIN_users.data = make(map[uint64]string)
	users.data = make(map[string] User)
	LoginRedirect.data = make(map[string]string)
	UsersLoginIn.data = make(map[uint64]string)
	WaitingForGoogle.data = make(map[string] User)
	ErrorTryingLogin.data = make(map[uint64]ErrorTrying)
	BannedUsers.data = make(map[uint64]string)
	//----

	LoginRedirect.mux.Lock()
	LoginRedirect.data["wallet"] = "http://localhost/"
	LoginRedirect.data["exchange"] = "http://127.0.0.1:8001/"
	LoginRedirect.mux.Unlock()

	loadBannedUsers()

	mailChennel = &MailChennel{
		sendToUser: make(chan *Mail_data, 10000),
	}

	google_chennel = &GoogleChennel{
		HandleGoogle: make(chan *GoogleAuth, 10000),
	}

	users := &Users{
		ChWritefile: make(chan *User, 10000),
	}

    loadUsers("")
    go listen_mail()
	go autoLogOut()

	var sndAddresses = make(map[int]string)
	var serversPorts = make(map[int]int)
	serversPorts[3] = LocalConfig.AuthSendToBalancePort // listening port
	serversPorts[1] = LocalConfig.AuthSendToWalletPort //Wallets port
	serversPorts[2] = LocalConfig.AuthSendToUContPort // mail port
	for i := 1; i <= len(serversPorts); i++  {
		SendToServers[i],_ = zmq.NewSocket(zmq.PUSH)
		defer SendToServers[i].Close()
		SendToServers[i].SetRcvhwm(1100000)
		sndAddresses[i] = fmt.Sprintf("%v%v", LocalConfig.LocalTcpIpAddr, serversPorts[i])
		SendToServers[i].Bind(sndAddresses[i])
		fmt.Println("Bind Wallet port: ", serversPorts[i])
		time.Sleep(10*time.Millisecond)
	}

	mailAUTH ,_ = zmq.NewSocket(zmq.PUSH)
	defer mailAUTH.Close()
	mailAUTH.SetRcvhwm(1100000)
	mailAUTHAddress := fmt.Sprintf("%v%v", LocalConfig.LocalTcpIpAddr, LocalConfig.AuthmailPort)
	mailAUTH.Bind(mailAUTHAddress);

	go sendToMailServer()

	mux := http.NewServeMux()
	mux.HandleFunc("/auth", handleAuth)
	mux.HandleFunc("/reg", users.handleReg)
	mux.HandleFunc("/logout", handleLogout)
	mux.HandleFunc("/callback", handleGoogleCallback)

	srv := &http.Server{
	    ReadTimeout:  10 * time.Second,
	    WriteTimeout: 10 * time.Second,
	    IdleTimeout:  1 * time.Second,
			Addr:           ":8008",
	    Handler:      mux,
	}

	//updateBanns(4)
	go users.writeFile()
	err2 := srv.ListenAndServe()
	if err2 != nil {
		log.Fatal("ListenAndServe error: ", err2)
	}
}
