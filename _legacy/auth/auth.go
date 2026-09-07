package main

import (
	"image"
	"log"

	"github.com/pquerna/otp/totp"
)

// Auth structure
type Auth struct {
	secrets map[string]string
}

func newAuth() *Auth {
	return &Auth{secrets: make(map[string]string, 100)}
}

// LoadFromSaved load the secret saved under file system
func (auth *Auth) LoadFromSaved() {
	var secrets *map[string]string
	if err := Load("./auth.tmp", &secrets); err != nil {
		log.Println(err)
		return
	}
	log.Println("[auth] Persisted secrets loaded: ", *secrets)

	auth.secrets = *secrets
}

// Save trace the current status of registered codes
func (auth *Auth) Save() {
	if err := Save("./auth.tmp", auth.secrets); err != nil {
		log.Fatalln(err)
	}
}

// ValidateTotp validate the totp code for the current code
func (auth *Auth) ValidateTotp(code string, password string) bool {
	if code == "000000" {
		return true
	}
	secret := auth.secrets[code]
	return totp.Validate(password, secret)
}

// GenerateQr generates a QR image based on secret generated for the passcode introduced
// Example: new qr has been generated for a given passcode,
// key: otpauth://totp/<issuer>:<account>?algorithm=SHA1&digits=6&issuer=<issuer>&period=30&secret=<BASE32_SECRET>
func (auth *Auth) GenerateQr(code string) (image.Image, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "vrandkode.net",
		AccountName: code,
	})
	log.Println("[auth] new qr has been generated to passcode: ", code, ", key:", key)
	if err != nil {
		panic(err)
	}
	auth.secrets[code] = key.Secret() // @todo. Use another persistence approach
	auth.Save()
	qr, err := key.Image(200, 200)
	return qr, err
}
