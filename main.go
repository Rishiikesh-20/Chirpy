package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
)

type apiConfig struct{
	fileServerHits atomic.Int32
}

func (ac *apiConfig) middleWareMetricsInc(next http.Handler) http.Handler{
	return http.HandlerFunc(func (w http.ResponseWriter,req *http.Request){
		ac.fileServerHits.Add(1)
		next.ServeHTTP(w,req)
	})
}

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

func (ac *apiConfig) resetCount (w http.ResponseWriter,req *http.Request){
	w.Header().Set("Content-Type","text/plain; charset=utf-8")
	w.WriteHeader(200)
	ac.fileServerHits.Swap(0)
}



func validate_chirp(w http.ResponseWriter,req *http.Request){
	type returnError struct{
		Error string `json:"error"`
	}
	type returnVal struct{
		CleanedBody string `json:"cleaned_body"`
	}
	type parameters struct{
		Body string `json:"body"`
	}
	respError:=returnError{
		Error: "Something went wrong",
	}
	anotherRespError:=returnError{
		Error: "Chirp is too long",
	}
	respBody:=returnVal{}

	decoder:=json.NewDecoder(req.Body)
	params:=parameters{}
	err:=decoder.Decode(&params)

	if err!=nil{
		dat,err:=json.Marshal(respError)
		if err!=nil{
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(500)
		w.Write([]byte(dat))
		return 
	}

	if len(params.Body)>140{
		dat,err:=json.Marshal(anotherRespError)
		if err!=nil{
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(400)
		w.Write([]byte(dat))
		return 
	}

	arrBody:=strings.Split(params.Body, " ")

	for i,val :=range arrBody{
		comp:=strings.ToLower(val)
		if(comp=="kerfuffle" || comp=="sharbert" || comp=="fornax"){
			arrBody[i]=strings.Repeat("*",4)
		}
	}

	respBody.CleanedBody=strings.Join(arrBody, " ")

	dat,err:=json.Marshal(respBody)

	if err!=nil{
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
	w.Write([]byte(dat))
	

}

func main(){
	mux:=http.NewServeMux()
	server:=&http.Server{
		Handler: mux,
		Addr: ":8080",
	}
	ac:=apiConfig{}
	mux.Handle("/app/",ac.middleWareMetricsInc(http.StripPrefix("/app",http.FileServer(http.Dir(".")))))

	mux.HandleFunc("GET /api/healthz",func(w http.ResponseWriter,req *http.Request){
		w.Header().Set("Content-Type","text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("OK"))

		
	})

	mux.HandleFunc("GET /admin/metrics",ac.getCount)
	mux.HandleFunc("POST /admin/reset",ac.resetCount)
	mux.HandleFunc("POST /api/validate_chirp",validate_chirp)



	log.Println("Server started at port 8080")

	err:=http.ListenAndServe(server.Addr,server.Handler)

	if(err!=nil){
		log.Println("Server Closed",err)
	}

}