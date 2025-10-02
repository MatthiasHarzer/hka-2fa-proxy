package otp

import (
	"time"

	"github.com/pquerna/otp/totp"
)

const CodeInterval = 30 * time.Second

type Generator interface {
	Generate(time time.Time) string
	WaitForNextInterval()
}

type generator struct {
	secret       string
	previousCode string
	previousTime time.Time
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

func (g *generator) WaitForNextInterval() {
	maxWaitTime := g.previousTime.Add(CodeInterval)

	for time.Now().Before(maxWaitTime) {
		code, err := totp.GenerateCode(g.secret, time.Now())
		if err != nil {
			return // This should never happen since we validate the secret on creation
		}
		if code != g.previousCode {
			g.previousCode = code
			return
		}

		time.Sleep(500 * time.Millisecond)
	}
	return
}

func (g *generator) Generate(time time.Time) string {
	code, err := totp.GenerateCode(g.secret, time)
	if err != nil {
		return "" // This should never happen since we validate the secret on creation
	}
	g.previousCode = code
	g.previousTime = time
	return code
}
