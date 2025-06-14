package apis

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2"
)

const redirectURL = "http://localhost:8080/oauth/callback"

type OAuthHandler struct {
	config         *oauth2.Config
	state          string
	tokFile        string
	tokenChan      chan *oauth2.Token
	errChan        chan error
	callbackServer *http.Server
	mu             sync.Mutex
}

func NewOAuthHandler(clientID, clientSecret, authURL, tokenURL, tokFile string, scopes []string) *OAuthHandler {
	return &OAuthHandler{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  authURL,
				TokenURL: tokenURL,
			},
		},
		state:     generateRandomState(),
		tokFile:   tokFile,
		tokenChan: make(chan *oauth2.Token),
		errChan:   make(chan error),
	}
}

func generateRandomState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func (h *OAuthHandler) GetClient() (*http.Client, error) {
	tok, err := h.tokenFromFile(h.tokFile)
	if err != nil {
		apiName, _ := strings.CutSuffix(h.tokFile, "_token.json")
		fmt.Printf("No token found or token invalid for %s api.\n", apiName)
		tok, err = h.getTokenFromWeb(h.config)
		if err != nil {
			return nil, fmt.Errorf("unable to get token from web: %w", err)
		}
		h.saveToken(h.tokFile, tok)
	}
	return h.config.Client(context.Background(), tok), nil
}

func (h *OAuthHandler) Logout() error {
	if err := os.Remove(h.tokFile); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("token file '%s' does not exist, nothing to log out", h.tokFile)
		}
		return fmt.Errorf("unable to remove token file '%s': %w", h.tokFile, err)
	}
	fmt.Printf("Logged out successfully. Removed token file '%s'", h.tokFile)
	return nil
}

func (h *OAuthHandler) Login() error {
	tok, err := h.getTokenFromWeb(h.config)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}
	h.saveToken(h.tokFile, tok) // saveToken already logs fatal on error, so no need to check here
	return nil
}

func (h *OAuthHandler) getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL(h.state, oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser to log in:\n%s\n", authURL)

	mux := http.NewServeMux()
	mux.HandleFunc("/oauth/callback", h.oauthCallbackHandler)

	h.mu.Lock()
	h.callbackServer = &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	h.mu.Unlock()

	go func() {
		if err := h.callbackServer.ListenAndServe(); err != nil {
			h.errChan <- fmt.Errorf("callback server error: %w", err)
		}
	}()

	select {
	case tok := <-h.tokenChan:
		h.shutdownCallbackServer()
		return tok, nil
	case err := <-h.errChan:
		h.shutdownCallbackServer()
		return nil, err
	case <-time.After(5 * time.Minute):
		h.shutdownCallbackServer()
		return nil, fmt.Errorf("OAuth callback timed out after 5 minutes")
	}
}

func (h *OAuthHandler) oauthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	queryState := r.FormValue("state")
	if queryState != h.state {
		err := fmt.Errorf("invalid state parameter: expected '%s', got '%s'", h.state, queryState)
		http.Error(w, err.Error(), http.StatusBadRequest)
		h.errChan <- err
		return
	}

	code := r.FormValue("code")
	if code == "" {
		err := fmt.Errorf("no authorization code found in callback")
		http.Error(w, err.Error(), http.StatusBadRequest)
		h.errChan <- err
		return
	}

	tok, err := h.config.Exchange(context.Background(), code)
	if err != nil {
		err = fmt.Errorf("unable to retrieve token from web: %w", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		h.errChan <- err
		return
	}

	// Send a success message back to the browser
	fmt.Fprintf(w, "Authentication successful! You can close this window.")
	h.tokenChan <- tok
}

func (h *OAuthHandler) shutdownCallbackServer() {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.callbackServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		h.callbackServer.Shutdown(ctx)
		h.callbackServer = nil
	}
}

func (h *OAuthHandler) tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	if err = json.NewDecoder(f).Decode(tok); err != nil {
		return nil, fmt.Errorf("failed to decode token from file '%s': %w", h.tokFile, err)
	}

	return tok, err
}

func (h *OAuthHandler) saveToken(path string, token *oauth2.Token) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
