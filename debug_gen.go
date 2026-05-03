package main

import (
	"fmt"

	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/password"
)

func main() {
	// Генерируем хеш именно твоей функцией с твоими DefaultParams
	pass := "owner12345"
	hash, err := password.HashPassword(pass, password.DefaultParams)
	if err != nil {
		panic(err)
	}
	fmt.Printf("\nПароль: %s\n", pass)
	fmt.Printf("Email: admin@foodcourt.fun\n")
	fmt.Printf("Хеш для SQL: %s\n\n", hash)
}
