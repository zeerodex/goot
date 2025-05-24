package gtasksapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/tasks/v1"
)

var (
	tokFile         = "token.json"
	credentialsFile = "credentials.json"
)

func getClient() (*http.Client, error) {
	config := getConfig()

	tok, err := tokenFromFile(tokFile)
	if err != nil {
		fmt.Println("No token found or token invalid. Please log in")
		tok, err := getTokenFromWeb(config)
		if err != nil {
			return nil, fmt.Errorf("unable to get token from web: %w", err)
		}
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok), nil
}

func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, fmt.Errorf("unable to read authorization code: %w", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve token from web: %w", err)
	}
	return tok, nil
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	if err = json.NewDecoder(f).Decode(tok); err != nil {
		return nil, fmt.Errorf("failed to decode token from file '%s': %w", tokFile, err)
	}

	if tok.Expiry.Before(time.Now()) {
		return nil, fmt.Errorf("token from file '%s' has expired", tokFile)
	}
	return tok, err
}

func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func getConfig() *oauth2.Config {
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, tasks.TasksScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	return config
}

func Logout() error {
	if err := os.Remove(tokFile); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("token file '%s' does not exist, nothing to log out", tokFile)
		}
		return fmt.Errorf("unable to remove token file '%s': %w", tokFile, err)
	}
	fmt.Printf("Logged out successfully. Removed token file '%s'", tokFile)
	return nil
}

func Login() error {
	config := getConfig()
	tok, err := getTokenFromWeb(config)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}
	saveToken(tokFile, tok) // saveToken already logs fatal on error, so no need to check here
	return nil
}

func GetService() (*tasks.Service, error) {
	client, err := getClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get authenticated client: %w", err)
	}
	srv, err := tasks.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve tasks service: %w", err)
	}
	return srv, nil
}
