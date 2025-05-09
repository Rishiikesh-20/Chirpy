package main

import (
	"fmt"
	"log"
	"net/http"
)

func (ac *apiConfig) getCount (w http.ResponseWriter,req *http.Request){
	w.Header().Set("Content-Type","text/html; charset=utf-8")
	w.WriteHeader(200)
	htmlFile:=fmt.Sprintf(`<html>
		<body>
		  <h1>Welcome, Chirpy Admin</h1>
		  <p>Chirpy has been visited %d times!</p>
		</body>
	  </html>`,ac.fileServerHits.Load())
	w.Write([]byte(htmlFile))
}

func (ac *apiConfig) resetUsers (w http.ResponseWriter,req *http.Request){
	if ac.PLATFORM!="dev"{
		w.WriteHeader(403)
		return 
	}
	w.Header().Set("Content-Type","text/plain; charset=utf-8")
	err:=ac.dbQueries.DeleteUser(req.Context())
	if err!=nil{
		log.Println("Error in Deleting user:",err)
		w.WriteHeader(500)
		return 
	}
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}