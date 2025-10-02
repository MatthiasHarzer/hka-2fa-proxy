package otp

import (
	"time"

	"github.com/pquerna/otp/totp"
)

type Generator interface {
	Generate(time time.Time) string
}

type generator struct {
	secret string
}

func NewGenerator(secret string) (Generator, error) {
	_, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		return nil, err
	}

	return &generator{
		secret: secret,
	}, nil
}

func (g *generator) Generate(time time.Time) string {
	code, err := totp.GenerateCode(g.secret, time)
	if err != nil {
		return "" // This should never happen since we validate the secret on creation
	}
	return code
}
