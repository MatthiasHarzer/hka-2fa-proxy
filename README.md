# HKA 2FA Proxy (owa)

A simple proxy to internal access of Outlook Web Access (OWA) by providing the OTP secret. This can be used to integrate the Outlook calendar into other applications, which usually requires providing the OTP next to calendar-share URL. 

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
<br>

## Setup
Download the [latest release](https://github.com/MatthiasHarzer/hka-2fa-proxy/releases) and add the executable to your `PATH`.

## Usage
1. Start the proxy with `hka-2fa-proxy -u <rz-username> -s <otp-secret> [-p <port>]`.
	 - The `-u` / `--username` flag is used to specify the RZ username.
	 - The `-s` / `--secret` flag is used to specify the OTP secret (Base32 encoded).
	 - The `-p` / `--port` flag is optional and specifies the port to listen on (default is 8080).
2. To use the proxy, replace the host of the URL with the host of the proxy. For example, if the original URL is `https://owa.h-ka.de/owa/calendar/...` and you proxy is running on `localhost:8080`, the proxied URL would be `http://localhost:8080/owa/calendar/...`.


