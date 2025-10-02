package run

import (
	"net/http"

	"github.com/MatthiasHarzer/hka-2fa-proxy/otp"
	"github.com/MatthiasHarzer/hka-2fa-proxy/proxy"
	"github.com/spf13/cobra"
)

var username string
var otpSecret string

func init() {
	Command.Flags().StringVarP(&username, "username", "u", "", "The username to use for authentication")
	Command.Flags().StringVarP(&otpSecret, "secret", "s", "", "The OTP-secret to use for generating the OTPs")
}

var Command = &cobra.Command{
	Use:   "run",
	Short: "Runs the proxy",
	Long:  "Runs the proxy",
	Run: func(c *cobra.Command, args []string) {
		generator, err := otp.NewGenerator(otpSecret)
		if err != nil {
			panic(err)
		}
		server := proxy.NewServer("https://owa.h-ka.de", username, generator)

		err = http.ListenAndServe(":8080", server)
		if err != nil {
			panic(err)
		}
	},
}
