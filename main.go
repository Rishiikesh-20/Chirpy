package main

import (
	"fmt"
	"log"
	"net/http"
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



	log.Println("Server started at port 8080")

	err:=http.ListenAndServe(server.Addr,server.Handler)

	if(err!=nil){
		log.Println("Server Closed",err)
	}

}