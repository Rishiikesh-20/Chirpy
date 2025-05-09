package auth

import "golang.org/x/crypto/bcrypt"



func HashPassword(password string) (string,error){
	response,err:=bcrypt.GenerateFromPassword([]byte(password),bcrypt.DefaultCost)
	return string(response),err
}

func CheckPasswordHash(hash,password string) error{
	err:=bcrypt.CompareHashAndPassword([]byte(hash),[]byte(password))
	return err
}