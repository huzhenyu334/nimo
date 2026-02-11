package main

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/bitfantasy/nimo/internal/config"
	"github.com/bitfantasy/nimo/internal/middleware"
	"reflect"
)

func main() {
	cfg, _ := config.Load()
	secret := cfg.JWT.Secret
	fmt.Println("Secret:", secret)
	
	tokenString := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImFkbWluQHRlc3QuY29tIiwiZXhwIjoxNzcwODExMTc2LCJmZWlzaHVfdWlkIjoiIiwiaWF0IjoxNzcwODA3NTc2LCJpc3MiOiJuaW1vLXBsbSIsImp0aSI6InRlc3QtanRpLTEyMyIsIm5hbWUiOiJBZG1pbiIsInBlcm1zIjpbXSwicm9sZXMiOltdLCJzdWIiOiJ0ZXN0LWFkbWluIiwidWlkIjoidGVzdC1hZG1pbiJ9.CiPB7ufUz2_wPpG6SUu87O8EHct0gqbg2Fb0LZhjEFE"
	
	// Use middleware's JWTClaims type
	_ = middleware.JWTAuth // just to confirm access
	
	type JWTClaims struct {
		UserID      string   `json:"uid"`
		Name        string   `json:"name"`
		Email       string   `json:"email"`
		FeishuUID   string   `json:"feishu_uid"`
		Roles       []string `json:"roles"`
		Permissions []string `json:"perms"`
		jwt.RegisteredClaims
	}
	
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	
	if err != nil {
		fmt.Println("Error:", err)
		fmt.Println("Error type:", reflect.TypeOf(err))
		return
	}
	
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		fmt.Println("Valid! UserID:", claims.UserID)
	} else {
		fmt.Println("Invalid claims or token not valid")
	}
}
