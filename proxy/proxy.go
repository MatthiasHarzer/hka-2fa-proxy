package proxy

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/MatthiasHarzer/hka-2fa-proxy/otp"
)

const (
	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36"
)

func isLoginSuccessful(response *http.Response) bool {
	return response.StatusCode == http.StatusFound // For some reason
}

func isLoginPage(body string) bool {
	return strings.Contains(body, "Welcome to HKA MFA-protected Services.")
}

func isRedirectToLogin(response *http.Response) bool {
	location := response.Header.Get("Location")
	return strings.Contains(location, "lm_auth_proxy")
}

type server struct {
	otpGenerator otp.Generator
	targetHost   string
	username     string
	client       *http.Client
	mutex        *sync.Mutex
}

func NewServer(targetHost, username string, otpGenerator otp.Generator) (http.Handler, error) {
	sv := &server{targetHost: targetHost, otpGenerator: otpGenerator, username: username, mutex: &sync.Mutex{}}
	err := sv.authenticateClient()
	if err != nil {
		return nil, fmt.Errorf("could not authenticate client: %w", err)
	}
	return sv, nil
}

// getLoginParameters performs the initial request to get session cookies and login form parameters.
func (s *server) getLoginParameters(client *http.Client) (url.Values, string, error) {
	// Create a request so we can set headers
	initialURL := s.targetHost + "/"
	req, err := http.NewRequest("GET", initialURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("could not create initial request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	// The client's CheckRedirect is configured to stop redirects
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("initial GET request failed: %w", err)
	}
	defer resp.Body.Close()

	// We expect a 302 Found status code
	if resp.StatusCode != http.StatusFound {
		return nil, "", fmt.Errorf("expected a 302 redirect, but got status %s", resp.Status)
	}

	locationHeader := resp.Header.Get("Location")
	if locationHeader == "" {
		return nil, "", fmt.Errorf("'Location' header not found in the response")
	}

	// Parse parameters from the unusual URL format (split by '?')
	parts := strings.Split(locationHeader, "?")
	if len(parts) < 2 {
		return nil, "", fmt.Errorf("could not parse query string from location: %s", locationHeader)
	}
	parsedParams, err := url.ParseQuery(parts[len(parts)-1])
	if err != nil {
		return nil, "", fmt.Errorf("could not parse query parameters: %w", err)
	}

	refererURL := s.targetHost + locationHeader
	return parsedParams, refererURL, nil
}

// submitLogin prompts for credentials and submits the login form.
func (s *server) submitLogin(client *http.Client, params url.Values, refererURL, username, password string) (*http.Response, error) {
	// Prepare form data for the POST request
	formData := url.Values{}
	formData.Set("curl", params.Get("curl"))
	formData.Set("curlid", params.Get("curlid"))
	formData.Set("curlmode", params.Get("curlmode"))
	formData.Set("username", strings.TrimSpace(username))
	formData.Set("password", password)

	postURL := fmt.Sprintf("%s/lm_auth_proxy?LMLogon", s.targetHost)

	req, err := http.NewRequest("POST", postURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("could not create POST request: %w", err)
	}

	// Set necessary headers
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Referer", refererURL)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Execute the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("login POST request failed: %w", err)
	}

	return resp, nil
}

func (s *server) authenticateClient() error {
	log.Println("authenticating client")

	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalf("failed to create cookie jar: %v", err)
	}

	// Create a custom HTTP client
	client := &http.Client{
		Jar: jar,
	}

	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		// This stops the client from following any redirects
		return http.ErrUseLastResponse
	}

	loginParams, refererURL, err := s.getLoginParameters(client)
	if err != nil {
		log.Fatalf("could not get login parameters: %v", err)
	}

	// This is required, as one OTP can only be used once. To prevent timing issues, we wait for the next interval.
	s.otpGenerator.WaitForNextInterval()
	password := s.otpGenerator.Generate(time.Now())

	loginResp, err := s.submitLogin(client, loginParams, refererURL, s.username, password)
	if err != nil {
		log.Fatalf("could not submit login: %v", err)
	}
	defer loginResp.Body.Close()

	if !isLoginSuccessful(loginResp) {
		return fmt.Errorf("login failed")
	}

	s.client = client

	log.Println("client authenticated successfully")

	return nil
}

func (s *server) proxyRequest(w http.ResponseWriter, r *http.Request, client *http.Client) error {
	// Create a new request based on the original one
	proxyReq, err := http.NewRequest(r.Method, s.targetHost+r.RequestURI, r.Body)
	if err != nil {
		return err
	}

	// Copy headers from the original request
	for name, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(name, value)
		}
	}
	proxyReq.Header.Set("User-Agent", userAgent)

	// Perform the request
	resp, err := client.Do(proxyReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	responseBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}
	responseBody := string(responseBodyBytes)

	if isLoginPage(responseBody) || isRedirectToLogin(resp) {
		return fmt.Errorf("not logged in anymore")
	}

	// Copy response headers
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	// Write the status code
	w.WriteHeader(resp.StatusCode)

	// Copy the response body
	_, err = w.Write(responseBodyBytes)
	if err != nil {
		log.Printf("error copying response body: %v", err)
	}
	return nil
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	log.Printf("proxying request: %s %s", r.Method, r.URL.String())

	err := s.proxyRequest(w, r, s.client)
	if err != nil {
		err = s.authenticateClient()
		if err != nil {
			http.Error(w, "re-authentication failed: "+err.Error(), http.StatusBadGateway)
			return
		}
		err = s.proxyRequest(w, r, s.client)
		if err != nil {
			http.Error(w, "proxy error after re-authentication: "+err.Error(), http.StatusBadGateway)
		}
	}
}
