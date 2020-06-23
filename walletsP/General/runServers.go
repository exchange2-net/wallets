package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

//Path to go files folder
var ServersPATH = "" //full path to current folder
var servers = map[int] string{}
var serversArgs = map[int] int{}
var haveArgs = map[string] int{} //programs that need args
var dynamicHaveArgs = map[string] int{}
var ServersProcces = map[string] string{}

func main() {
	os.Chdir(ServersPATH) // change execute dir
	var argFlag bool
	ServersList := ServersName() // get list of Go files
	ServersListArgs := ServersArgs()  // get ARGS list
	ServersListHaveArgs := ServersHaveArgs() // get list of programs that need args

	//starting Go files
	for _, serverName := range ServersList {
		time.Sleep(5 * time.Second)
		argFlag = false
		for argsServer := range ServersListHaveArgs {
			if serverName == argsServer {
				for _, args := range ServersListArgs {
					fmt.Println("Starting :",serverName, args)
					dynamicHaveArgs[serverName] = 1
					time.Sleep(5 * time.Second)
					CurrentArgs := fmt.Sprintf("%v", args)
					go run_apps(1, serverName, CurrentArgs)
					argFlag = true
				}
			} else {
				argFlag = false
			}
		}
		if argFlag == false {
			if dynamicHaveArgs[serverName] == 1 {
				fmt.Println("continue...")
			} else {
				fmt.Println("Starting :",serverName)
				go run_apps(0, serverName, "")
			}
		}
	}
	time.Sleep(5 * time.Second)
	fmt.Println("All servers stared...")
	var argsFlag int
	for { // loop for check stopped programs and rerun it
		time.Sleep(5 * time.Second)
		fmt.Println(ServersProcces)
		for name, args := range ServersProcces {
			if args == "" {
				argsFlag = 0
			} else {
				argsFlag = 1
			}
			go rerun_server(argsFlag, name, args)
			delete(ServersProcces, name)
			fmt.Println(ServersProcces)
			time.Sleep(5 * time.Second)
		}
	}
	os.Exit(0)

}

/*
function rerun_server
function rerun stopped programs
 */
func rerun_server(flag int, serverName string, CurrentArgs string) {
	fmt.Println("Re RUN SERVER - ", serverName, CurrentArgs)
	if flag == 1 {
		cmd := exec.Command("go", "run", serverName, CurrentArgs)
		cmd.Dir = ServersPATH
		cmd.Stdout = nil
		cmd.Run()
		ServersProcces[serverName] = CurrentArgs
		cmd.Process.Kill()
		fmt.Println(cmd.Process.Pid)
	} else {
		cmd := exec.Command("go", "run", serverName)
		cmd.Dir = ServersPATH
		cmd.Stdout = nil
		cmd.Run()
		ServersProcces[serverName] = CurrentArgs
		cmd.Process.Kill()
		fmt.Println(cmd.Process.Pid)
	}
}

/*
function ServersName
function initialize list of Go programs
 */
func ServersName() map[int]string{
	servers[0] = "SupportServer.go"
	servers[1] = "authorizationServer.go"
	servers[2] = "balanceServer.go"
	servers[3] = "mailServer.go"
	servers[4] = "UserContacts.go"
	servers[5] = "Wallets.go"
	return servers
}

/*
function ServersArgs
function initialize list of ARGS
*/
func ServersArgs() map[int]int {
	return serversArgs
}

/*
function ServersArgs
function initialize list of programs that need args
*/
func ServersHaveArgs() map[string]int{
	return haveArgs
}

/*
function run_apps
args: 1 or 0, program name, program args
function that run programs
 */
func run_apps(flag int, serverName string, CurrentArgs string) {
	if flag == 1 {
		cmd := exec.Command("go", "run", serverName, CurrentArgs)
		cmd.Dir = ServersPATH
		cmd.Stdout = nil
		cmd.Run()
		ServersProcces[serverName] = CurrentArgs
		cmd.Process.Kill()
		fmt.Println(cmd.Process.Pid)
	} else {
		cmd := exec.Command("go", "run", serverName)
		cmd.Dir = ServersPATH
		cmd.Stdout = nil
		cmd.Run()
		ServersProcces[serverName] = CurrentArgs
		cmd.Process.Kill()
		fmt.Println(cmd.Process.Pid)
	}
}