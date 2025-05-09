package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Rishiikesh-20/Chirpy/internal/auth"
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



func (ac *apiConfig) validate_chirp(w http.ResponseWriter,req *http.Request){
	type returnError struct{
		Error string `json:"error"`
	}
	respError:=returnError{
		Error: "Something went wrong",
	}
	anotherRespError:=returnError{
		Error: "Chirp is too long",
	}
	decoder:=json.NewDecoder(req.Body)
	params:=database.CreateChirpParams{}
	err:=decoder.Decode(&params)

	if err!=nil{
		respError.Error+=err.Error()
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
	params.Body=strings.Join(arrBody," ")

	bearerToken,err:=auth.GetBearerToken(req.Header)

	if err!=nil{
		log.Println(err)
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

	_,err1:=auth.ValidateJWT(bearerToken,ac.JWT_SECRET)

	if err1!=nil{
		log.Println(err1)
		respError.Error+=": Invalid token"
		dat,err:=json.Marshal(respError)
		if err!=nil{
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(401)
		w.Write([]byte(dat))
		return 
	}

	dat,err:=ac.dbQueries.CreateChirp(req.Context(),params)

	if err!=nil{
		log.Println("Error: ",err)
		w.WriteHeader(500)
		return 
	}

	res,err:=json.Marshal(dat)

	if err!=nil{
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
	w.Write([]byte(res))

}

func (acg *apiConfig) createUser(w http.ResponseWriter,req *http.Request){
	type reqBody struct{
		Email string `json:"email"`
		Password string `json:"password"`
	}
	type returnError struct{
		Error string `json:"error"`
	}
	respError:=returnError{
		Error: "Something went wrong",
	}

	r:=reqBody{}
	decoder:=json.NewDecoder(req.Body)
	err:=decoder.Decode(&r)

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

	rs:=database.CreateUserParams{}
	rs.Email=r.Email
	rs.HashedPassword,err=auth.HashPassword(r.Password)
	if err!=nil{
		respError.Error+=err.Error()
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

	users,err:=acg.dbQueries.CreateUser(req.Context(),rs)

	type response struct{
		ID             int32 `json:"id"`
		Email          string `json:"email"`
		CreatedAt      time.Time `json:"created_at"`
		UpdatedAt      time.Time `json:"updated_at"`
		hashedPassword string 
	}

	userDetails:=response{}

	userDetails.ID=users.ID
	userDetails.Email=users.Email
	userDetails.CreatedAt=users.CreatedAt
	userDetails.UpdatedAt=users.UpdatedAt
	userDetails.hashedPassword=users.HashedPassword

	if err!=nil{
		log.Println("Error in Creating user:",err)
		w.WriteHeader(500)
		return 
	}

	dat,err:=json.Marshal(userDetails)
	if err!=nil{
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(201)
	w.Write([]byte(dat))

}

func (ac *apiConfig) getOneChirp(w http.ResponseWriter,req *http.Request){

	type returnError struct{
		Error string `json:"error"`
	}
	respError:=returnError{
		Error: "Chirp not found",
	}
	
	id:=req.PathValue("chirpId")

	idInt, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid chirpId", http.StatusBadRequest)
		return
	}
	response,err:=ac.dbQueries.GetOneUser(req.Context(),int32(idInt))

	w.Header().Set("Content-Type","application/json; charset=utf-8")
	if err!=nil{
		respError.Error+=err.Error()
		dat,err:=json.Marshal(respError)
		if err!=nil{
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(404)
		w.Write([]byte(dat))
		return 
	}

	w.WriteHeader(200)

	dat,err:=json.Marshal(response)

	if err!=nil{
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	w.Write([]byte(dat))


}

func (ac *apiConfig) getAllChirps(w http.ResponseWriter,req *http.Request){

	type returnError struct{
		Error string `json:"error"`
	}
	respError:=returnError{
		Error: "Something went wrong",
	}

	response,err:=ac.dbQueries.GetAllChirps(req.Context())
	w.Header().Set("Content-Type","application/json; charset=utf-8")
	if err!=nil{
		respError.Error+=err.Error()
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

	dat,err:=json.Marshal(response)

	if err!=nil{
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte(dat))
}

func (ac *apiConfig) loginFunc(w http.ResponseWriter,req *http.Request){
	type requestBody struct{
		Email string `json:"email"`
		Password string `json:"password"`
		Expires_in_seconds int `json:"expires_in_second"`
	}
	type returnError struct{
		Error string `json:"error"`
	}
	type returnIncorrectEmail struct{
		Error string `json:"error"`
	}
	type returnIncorrectPassword struct{
		Error string `json:"error"`
	}
	respError:=returnError{
		Error: "Something went wrong",
	}
	respEmail:=returnIncorrectEmail{
		Error: "Incorrect Email",
	}
	respPassword:=returnIncorrectPassword{
		Error: "Incorrect Password",
	}
	
	params:=requestBody{}
	decoder:=json.NewDecoder(req.Body)
	err:=decoder.Decode(&params)

	if err!=nil{
		respError.Error+=err.Error()
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

	if params.Expires_in_seconds==0 || params.Expires_in_seconds>3600{
		params.Expires_in_seconds=3600 
	}


	users,err:=ac.dbQueries.GetOneUserByEmail(req.Context(),params.Email)

	if err!=nil{
		log.Println(err)
		dat,err:=json.Marshal(respEmail)
		if err!=nil{
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(401)
		w.Write([]byte(dat))
		return 
	}


	err=auth.CheckPasswordHash(users.HashedPassword,params.Password)

	if err!=nil{
		dat,err:=json.Marshal(respPassword)
		if err!=nil{
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(401)
		w.Write([]byte(dat))
		return
	}

	token,err:=auth.MakeJWT(int(users.ID),ac.JWT_SECRET,time.Duration(params.Expires_in_seconds)*time.Second)

	if err!=nil{
		respError.Error+=err.Error()
		dat,err:=json.Marshal(respError)
		if err!=nil{
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return	
		}
		w.WriteHeader(500)
		w.Write(dat)
	}
	w.WriteHeader(200)
	type response2 struct{
		ID             int32 `json:"id"`
		Email          string `json:"email"`
		CreatedAt      time.Time `json:"created_at"`
		UpdatedAt      time.Time `json:"updated_at"`
		hashedPassword string 
		Token string `json:"token"`
	}

	userDetails:=response2{}

	userDetails.ID=users.ID
	userDetails.Email=users.Email
	userDetails.CreatedAt=users.CreatedAt
	userDetails.UpdatedAt=users.UpdatedAt
	userDetails.hashedPassword=users.HashedPassword
	userDetails.Token=token

	dat,err:=json.Marshal(userDetails)

	if err!=nil{
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	w.Write(dat)
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

	



	log.Println("Server started at port 8080")

	err=http.ListenAndServe(server.Addr,server.Handler)

	if(err!=nil){
		log.Println("Server Closed",err)
	}

}