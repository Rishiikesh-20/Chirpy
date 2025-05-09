package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Rishiikesh-20/Chirpy/internal/auth"
)

func (ac *apiConfig) GetNewToken(w http.ResponseWriter,req *http.Request){
	type returnError struct{
		Error string `json:"error"`
	}
	type responseBody struct{
		Token string `json:"token"`
	}
	respError:=returnError{
		Error: "Something went wrong",
	}

	token,err:=auth.GetBearerToken(req.Header)
	if err!=nil{
		respError.Error+=err.Error()
		dat,err:=json.Marshal(respError)
		if err!=nil{
			w.WriteHeader(500)
			log.Println("Error marshalling JSON: %s", err)
			return
		}
		w.WriteHeader(401)
		w.Write(dat)
		return
	}

	res,err:=ac.dbQueries.GetUserFromRefreshToken(req.Context(),token)

	if err!=nil{
		log.Println(err)
		respError.Error+=": Not found"
		dat,err:=json.Marshal(respError)
		if err!=nil{
			w.WriteHeader(500)
			log.Println("Error marshalling JSON: %s", err)
			return
		}
		w.WriteHeader(401)
		w.Write(dat)
		return
	}

	if res.Token=="" {
		respError.Error+=": Token is not there"
		dat,err:=json.Marshal(respError)
		if err!=nil{
			w.WriteHeader(500)
			log.Println("Error marshalling JSON: %s", err)
			return
		}
		w.WriteHeader(401)
		w.Write(dat)
		return
	}
	log.Println(res.RevokedAt)
	if  res.RevokedAt.Valid{
		respError.Error+=": Revoked Refresh Token"
		dat,err:=json.Marshal(respError)
		if err!=nil{
			w.WriteHeader(500)
			log.Println("Error marshalling JSON: %s", err)
			return
		}
		w.WriteHeader(401)
		w.Write(dat)
		return
	}

	if  res.ExpiresAt.Before(time.Now()) {
		respError.Error+=": Expired Refresh token"
		dat,err:=json.Marshal(respError)
		if err!=nil{
			w.WriteHeader(500)
			log.Println("Error marshalling JSON: %s", err)
			return
		}
		w.WriteHeader(401)
		w.Write(dat)
		return
	}

	tokenJWT,err:=auth.MakeJWT(int(res.UserID),ac.JWT_SECRET,1*time.Hour)

	if err!=nil{
		log.Println(err)
		dat,err:=json.Marshal(respError)
		if err!=nil{
			w.WriteHeader(500)
			log.Println("Error marshalling JSON: %s", err)
			return
		}
		w.WriteHeader(401)
		w.Write(dat)
		return
	}

	tokenStruct:=responseBody{
		Token:tokenJWT,
	}

	dat,err:=json.Marshal(tokenStruct)
	if err!=nil{
		w.WriteHeader(500)
		log.Println("Error marshalling JSON: %s", err)
		return
	}

	w.WriteHeader(200)
	w.Write(dat)

}

func (ac *apiConfig) revokeTheToken(w http.ResponseWriter,req *http.Request){
	type returnError struct{
		Error string `json:"error"`
	}
	respError:=returnError{
		Error: "Something went wrong",
	}

	token,err:=auth.GetBearerToken(req.Header)
	if err!=nil{
		respError.Error+=err.Error()
		dat,err:=json.Marshal(respError)
		if err!=nil{
			w.WriteHeader(500)
			log.Println("Error marshalling JSON: %s", err)
			return
		}
		w.WriteHeader(401)
		w.Write(dat)
		return
	}

	err=ac.dbQueries.RevokeRefreshToken(req.Context(),token)

	if err!=nil{
		log.Println(err)
		respError.Error+=": Not found"
		dat,err:=json.Marshal(respError)
		if err!=nil{
			w.WriteHeader(500)
			log.Println("Error marshalling JSON: %s", err)
			return
		}
		w.WriteHeader(401)
		w.Write(dat)
		return
	}

	w.WriteHeader(204)
	return

}