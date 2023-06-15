package pt_af_logic

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
)

func generatePassword(passwordLength int) (string, error) {
	var password strings.Builder

	for i := 0; i < passwordLength; i++ {
		random, err := rand.Int(rand.Reader, big.NewInt(int64(len(allCharSet))))
		if err != nil {
			return "", fmt.Errorf("rand.Int: %w", err)
		}
		password.WriteString(string(allCharSet[random.Int64()]))
	}

	return password.String(), nil
}
