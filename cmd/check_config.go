package main

import (
	"fmt"
	"github.com/bitfantasy/nimo/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("JWT Secret:", cfg.JWT.Secret)
	fmt.Println("JWT Issuer:", cfg.JWT.Issuer)
}
