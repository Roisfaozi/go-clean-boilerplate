package sso

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/microsoft"
)

type MicrosoftProvider struct {
	config *oauth2.Config
}

func NewMicrosoftProvider(cfg ProviderConfig) *MicrosoftProvider {
	scopes := cfg.Scopes
	if len(scopes) == 0 {
		scopes = []string{"User.Read"}
	}

	conf := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Scopes:       scopes,
		Endpoint:     microsoft.AzureADEndpoint("common"),
	}

	return &MicrosoftProvider{
		config: conf,
	}
}

func (p *MicrosoftProvider) GetLoginURL(state string) string {
	return p.config.AuthCodeURL(state)
}

func (p *MicrosoftProvider) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	return p.config.Exchange(ctx, code)
}

func (p *MicrosoftProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
	client := p.config.Client(ctx, token)
	req, err := http.NewRequest("GET", "https://graph.microsoft.com/v1.0/me", nil)
	if err != nil {
		return nil, fmt.Errorf("failed creating request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed getting user info, status code %d", resp.StatusCode)
	}

	var data struct {
		Id                string `json:"id"`
		UserPrincipalName string `json:"userPrincipalName"`
		Mail              string `json:"mail"`
		DisplayName       string `json:"displayName"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed decoding user info: %w", err)
	}

	email := data.Mail
	if email == "" {
		email = data.UserPrincipalName
	}

	return &UserInfo{
		Email:      email,
		ProviderID: data.Id,
		Name:       data.DisplayName,
		AvatarURL:  "", // Avatar skipped for simplicity, requires another API call.
	}, nil
}
