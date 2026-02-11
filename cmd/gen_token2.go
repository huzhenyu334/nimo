package main

import (
	"fmt"
	"os"
	"time"
	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	UserID      string   `json:"uid"`
	Name        string   `json:"name"`
	Email       string   `json:"email"`
	FeishuUID   string   `json:"feishu_uid"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"perms"`
	jwt.RegisteredClaims
}

func main() {
	secret := "your-jwt-secret-key-change-in-production"
	now := time.Now()
	
	claims := jwt.MapClaims{
		"sub":        "test-admin",
		"uid":        "test-admin",
		"name":       "Admin",
		"email":      "admin@test.com",
		"feishu_uid": "",
		"roles":      []string{},
		"perms":      []string{},
		"iss":        "nimo-plm",
		"iat":        now.Unix(),
		"exp":        now.Add(time.Hour).Unix(),
		"jti":        "test-jti-123",
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		fmt.Fprintf(os.Stderr, "sign error: %v\n", err)
		os.Exit(1)
	}
	
	// Verify by parsing
	parsed, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
		os.Exit(1)
	}
	if c, ok := parsed.Claims.(*JWTClaims); ok && parsed.Valid {
		fmt.Fprintf(os.Stderr, "Token verified: uid=%s\n", c.UserID)
	}
	
	fmt.Print(tokenString)
}
