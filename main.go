package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"github.com/Rishiikesh-20/Chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct{
	fileServerHits atomic.Int32
	dbQueries *database.Queries
	PLATFORM string
	JWT_SECRET string
}

func main(){
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	jwtSecret:=os.Getenv("JWT_SECRET")
	db,err:=sql.Open("postgres",dbURL)
	dbQueries:=database.New(db)
	mux:=http.NewServeMux()
	server:=&http.Server{
		Handler: mux,
		Addr: ":8080",
	}
	ac:=apiConfig{}
	ac.dbQueries=dbQueries
	ac.JWT_SECRET=jwtSecret
	ac.PLATFORM=os.Getenv("PLATFORM")
	mux.Handle("/app/",ac.middleWareMetricsInc(http.StripPrefix("/app",http.FileServer(http.Dir(".")))))

	mux.HandleFunc("GET /api/healthz",func(w http.ResponseWriter,req *http.Request){
		w.Header().Set("Content-Type","text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("OK"))

		
	})

	mux.HandleFunc("GET /admin/metrics",ac.getCount)
	mux.HandleFunc("GET /api/chirps",ac.getAllChirps)
	mux.HandleFunc("GET /api/chirps/{chirpId}",ac.getOneChirp)

	mux.HandleFunc("POST /admin/reset",ac.resetUsers)
	mux.HandleFunc("POST /api/chirps",ac.validate_chirp)
	mux.HandleFunc("POST /api/users",ac.createUser)
	mux.HandleFunc("POST /api/login",ac.loginFunc)
	mux.HandleFunc("POST /api/refresh",ac.GetNewToken)
	mux.HandleFunc("POST /api/revoke",ac.revokeTheToken)

	mux.HandleFunc("PUT /api/users",ac.updateUserInfo)

	log.Println("Server started at port 8080")

	err=http.ListenAndServe(server.Addr,server.Handler)

	if(err!=nil){
		log.Println("Server Closed",err)
	}

}