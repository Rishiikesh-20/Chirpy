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

	mux.Handle("/app/",http.StripPrefix("/app",http.FileServer(http.Dir("."))))

	mux.HandleFunc("/healthz",func(w http.ResponseWriter,req *http.Request){
		w.Header().Set("Content-Type","text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("OK"))
		
	})



	log.Println("Server started at port 8080")

	err:=http.ListenAndServe(server.Addr,server.Handler)

	if(err!=nil){
		log.Println("Server Closed",err)
	}

}