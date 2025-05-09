package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)



func HashPassword(password string) (string,error){
	response,err:=bcrypt.GenerateFromPassword([]byte(password),bcrypt.DefaultCost)
	return string(response),err
}

func CheckPasswordHash(hash,password string) error{
	err:=bcrypt.CompareHashAndPassword([]byte(hash),[]byte(password))
	return err
}

func MakeJWT(userId int,tokenSecret string,expiresIn time.Duration) (string,error) {
	claims:=jwt.RegisteredClaims{
		Issuer: "chirpy",
		IssuedAt: jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject: strconv.Itoa((userId)),
	}
	token:=jwt.NewWithClaims(jwt.SigningMethodHS256,claims)

	signedToken,err:=token.SignedString([]byte(tokenSecret))

	if err!=nil{
		return "",err
	}
	return signedToken,nil
	
}

func ValidateJWT(tokenstring,tokenSecret string) (int,error) {
	claims:=&jwt.RegisteredClaims{}
	token,err:=jwt.ParseWithClaims(tokenstring,claims,func(token *jwt.Token)(interface{},error){
		if token.Method!=jwt.SigningMethodHS256{
			return nil,errors.New("signing method is incorrect")
		}
		return []byte(tokenSecret),nil

	})
	if err!=nil{
		return 0,err
	}

	if !token.Valid{
		return 0,errors.New("invalid token")
	}
	userId,err:=strconv.Atoi(claims.Subject)

	if err!=nil{
		return 0,errors.New("invalid userid in token subject")
	}

	return userId,nil
}

func GetBearerToken(headers http.Header)(string,error){
	detail:=headers["Authorization"]
	if len(detail) == 0 {
		return "", errors.New("authorization header not found")
	}
	if !strings.HasPrefix(detail[0],"Bearer "){
		return "",errors.New("invalid Authorization format")
	}
	result:=strings.TrimPrefix(detail[0],"Bearer ")

	return result,nil
}

func MakeRefreshToken()(string){
	key:=make([]byte,32)
	rand.Read(key)
	str:=hex.EncodeToString(key)
	return str
}

