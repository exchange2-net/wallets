package main

import (
	"fmt"
	"log"
	"net/http"
)
/*
Create File server to display HTML files
 */
func main(){
	fmt.Println("Starting File server")
	fs := http.FileServer(http.Dir("../WHTML/"))
	http.Handle("/", fs)
	err := http.ListenAndServe(fmt.Sprintf("localhost:%v", 80), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}