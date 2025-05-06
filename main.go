package main

import (
	"log"
	"net/http"
)

func main(){
	mux:=http.NewServeMux()
	server:=&http.Server{
		Handler: mux,
		Addr: ":8080",
	}

	mux.Handle("/",http.FileServer(http.Dir(".")))

	log.Println("Server started at port 8080")

	err:=http.ListenAndServe(server.Addr,server.Handler)

	if(err!=nil){
		log.Println("Server Closed")
	}

}