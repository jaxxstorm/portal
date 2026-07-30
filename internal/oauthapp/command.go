// Package oauthapp provides local bootstrap commands for Portal's Tailscale
// OAuth App integration.
package oauthapp

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const (
	defaultTokenURL     = "https://api.tailscale.com/api/v2/oauth/token"
	defaultOAuthAppsURL = "https://api.tailscale.com/api/v2/tailnet/-/oauth-apps"
	callbackPath        = "/_portal/auth/tailscale/callback"
)

// Run executes Portal's OAuth App command when args starts with "oauth-app".
// It returns handled=false for normal Portal serving commands.
func Run(ctx context.Context, args []string, stdout, stderr io.Writer) (handled bool, err error) {
	if len(args) == 0 || args[0] != "oauth-app" {
		return false, nil
	}
	cmd := newCommand(stdout, stderr, http.DefaultClient, defaultTokenURL, defaultOAuthAppsURL)
	cmd.SetArgs(args[1:])
	return true, cmd.ExecuteContext(ctx)
}

func newCommand(stdout, stderr io.Writer, httpClient *http.Client, tokenURL, oauthAppsURL string) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "oauth-app",
		Short:         "Create and configure Portal Tailscale OAuth Apps",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.AddCommand(newCreateCommand(stdout, httpClient, tokenURL, oauthAppsURL))
	return cmd
}

func newCreateCommand(stdout io.Writer, httpClient *http.Client, tokenURL, oauthAppsURL string) *cobra.Command {
	var clientID, clientSecret, hostname, name, description string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Tailscale OAuth App for Portal login",
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := Create(cmd.Context(), httpClient, tokenURL, oauthAppsURL, CreateRequest{
				ManagementClientID:     clientID,
				ManagementClientSecret: clientSecret,
				Name:                   name,
				Description:            description,
				Hostname:               hostname,
			})
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(stdout, "OAuth App created. Store the two secrets securely; they are not written to disk.\n\nCallback URL:\n%s\n\nClient ID:\n%s\n\nOAuth App client secret:\n%s\n\nPortal session secret:\n%s\n\nPortal authorization flags:\n--auth-mode=tailscale\n--tailscale-oauth-client-id=%q\n--tailscale-oauth-client-secret=%q\n--tailscale-session-secret=%q\n", result.CallbackURL, result.ClientID, result.ClientSecret, result.SessionSecret, result.ClientID, result.ClientSecret, result.SessionSecret)
			return err
		},
	}
	cmd.Flags().StringVar(&clientID, "oauth-client-id", "", "Management OAuth client ID with OAuth Apps permission")
	cmd.Flags().StringVar(&clientSecret, "oauth-client-secret", "", "Management OAuth client secret with OAuth Apps permission")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Public hostname for Portal (for example portal.tailnet.ts.net)")
	cmd.Flags().StringVar(&name, "name", "portal", "OAuth App name")
	cmd.Flags().StringVar(&description, "description", "Portal browser login", "OAuth App description")
	_ = cmd.MarkFlagRequired("oauth-client-id")
	_ = cmd.MarkFlagRequired("oauth-client-secret")
	_ = cmd.MarkFlagRequired("hostname")
	return cmd
}

func callbackURLForHostname(hostname string) (string, error) {
	hostname = strings.TrimSpace(hostname)
	if hostname == "" || strings.Contains(hostname, "://") {
		return "", fmt.Errorf("invalid hostname %q", hostname)
	}

	parsed, err := url.Parse("https://" + hostname)
	if err != nil || parsed.Host == "" || parsed.Hostname() == "" || parsed.User != nil || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", fmt.Errorf("invalid hostname %q: provide a hostname without a scheme, path, query, or fragment", hostname)
	}
	return "https://" + parsed.Host + callbackPath, nil
}

// CreateRequest contains the management OAuth client and Portal app details.
type CreateRequest struct {
	ManagementClientID     string
	ManagementClientSecret string
	Name                   string
	Description            string
	Hostname               string
}

// CreateResult holds the generated OAuth App credentials and Portal session secret.
type CreateResult struct {
	CallbackURL   string
	ClientID      string
	ClientSecret  string
	SessionSecret string
}

// Create exchanges management OAuth client credentials and creates an OAuth App.
func Create(ctx context.Context, httpClient *http.Client, tokenURL, oauthAppsURL string, request CreateRequest) (CreateResult, error) {
	if strings.TrimSpace(request.ManagementClientID) == "" || request.ManagementClientSecret == "" {
		return CreateResult{}, fmt.Errorf("management OAuth client ID and secret are required")
	}
	if strings.TrimSpace(request.Name) == "" {
		return CreateResult{}, fmt.Errorf("OAuth App name is required")
	}
	callbackURL, err := callbackURLForHostname(request.Hostname)
	if err != nil {
		return CreateResult{}, err
	}
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	tokenConfig := clientcredentials.Config{
		ClientID:     request.ManagementClientID,
		ClientSecret: request.ManagementClientSecret,
		TokenURL:     tokenURL,
		AuthStyle:    oauth2.AuthStyleInHeader,
	}
	token, err := tokenConfig.Token(ctx)
	if err != nil {
		return CreateResult{}, fmt.Errorf("obtaining management OAuth token: %w", err)
	}

	payload, err := json.Marshal(map[string]any{
		"name":         request.Name,
		"description":  request.Description,
		"redirectURIs": []string{callbackURL},
		"scopes":       []string{"openid"},
	})
	if err != nil {
		return CreateResult{}, err
	}
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, oauthAppsURL, strings.NewReader(string(payload)))
	if err != nil {
		return CreateResult{}, err
	}
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.SetBasicAuth(token.AccessToken, "")
	response, err := httpClient.Do(httpRequest)
	if err != nil {
		return CreateResult{}, fmt.Errorf("creating OAuth App: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 1<<20))
		return CreateResult{}, fmt.Errorf("creating OAuth App: %s: %s", response.Status, strings.TrimSpace(string(body)))
	}
	var created map[string]any
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&created); err != nil {
		return CreateResult{}, fmt.Errorf("decoding OAuth App response: %w", err)
	}
	clientID := responseString(created, "id", "clientId", "client_id", "clientID")
	clientSecret := responseString(created, "clientSecret", "client_secret", "clientSecretValue")
	if clientID == "" || clientSecret == "" {
		return CreateResult{}, fmt.Errorf("OAuth App response did not include client credentials")
	}
	sessionSecret, err := newSessionSecret()
	if err != nil {
		return CreateResult{}, err
	}
	return CreateResult{CallbackURL: callbackURL, ClientID: clientID, ClientSecret: clientSecret, SessionSecret: sessionSecret}, nil
}

func responseString(response map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := response[key].(string); ok {
			return value
		}
	}
	return ""
}

func newSessionSecret() (string, error) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return "", fmt.Errorf("generating session secret: %w", err)
	}
	return hex.EncodeToString(secret), nil
}
