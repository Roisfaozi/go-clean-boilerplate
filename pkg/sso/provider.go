package sso

import (
	"context"

	"golang.org/x/oauth2"
)

type ProviderConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

type UserInfo struct {
	Email      string
	ProviderID string
	Name       string
	AvatarURL  string
}

type Provider interface {
	GetLoginURL(state string) string
	ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error)
	GetUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error)
}
