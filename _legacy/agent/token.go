package main

import (
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
)

// GenerateToken returns a generated token based on expected size
func GenerateToken(n int) string {
	var letter = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	return string(b)
}

// AgentToken provides the saved or generated a new token if requires
// if token was not saved it will be generated
func AgentToken() string {
	token, err := GetSavedToken()
	if err == nil {
		return token
	}
	token = GenerateToken(8)
	SaveToken(token)
	return token
}

// SaveToken save the generated token under file system
func SaveToken(token string) {
	f, err := os.Create(".stk")
	check(err)
	defer f.Close()
	f.WriteString(token + "\n")
}

// GetSavedToken restore the token from file system
func GetSavedToken() (string, error) {
	//return "dcb6888d", nil
	data, err := ioutil.ReadFile(".stk")
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(string(data), "\n"), nil
}
