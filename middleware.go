package main

import "net/http"

func (ac *apiConfig) middleWareMetricsInc(next http.Handler) http.Handler{
	return http.HandlerFunc(func (w http.ResponseWriter,req *http.Request){
		ac.fileServerHits.Add(1)
		next.ServeHTTP(w,req)
	})
}