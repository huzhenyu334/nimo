package main

import (
	"fmt"
	"reflect"
	"github.com/golang-jwt/jwt/v5"
	"github.com/bitfantasy/nimo/internal/config"
	"github.com/bitfantasy/nimo/internal/middleware"
)

func main() {
	cfg, _ := config.Load()
	secret := cfg.JWT.Secret
	
	tokenString := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImFkbWluQHRlc3QuY29tIiwiZXhwIjoxNzcwODExMTc2LCJmZWlzaHVfdWlkIjoiIiwiaWF0IjoxNzcwODA3NTc2LCJpc3MiOiJuaW1vLXBsbSIsImp0aSI6InRlc3QtanRpLTEyMyIsIm5hbWUiOiJBZG1pbiIsInBlcm1zIjpbXSwicm9sZXMiOltdLCJzdWIiOiJ0ZXN0LWFkbWluIiwidWlkIjoidGVzdC1hZG1pbiJ9.CiPB7ufUz2_wPpG6SUu87O8EHct0gqbg2Fb0LZhjEFE"
	
	claims := &middleware.JWTClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	
	if err != nil {
		fmt.Println("Error:", err)
		fmt.Println("Error type:", reflect.TypeOf(err))
		return
	}
	
	if c, ok := token.Claims.(*middleware.JWTClaims); ok && token.Valid {
		fmt.Println("Valid! UserID:", c.UserID)
	} else {
		fmt.Println("Invalid claims")
	}
}
