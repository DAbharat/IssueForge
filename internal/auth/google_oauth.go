package auth

import (
	"context"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type GoogleOIDC struct {
	Config   *oauth2.Config
	Verifier *oidc.IDTokenVerifier
}

func NewGoogleOauthConfig(clientID, clientSecret, redirectURL string, ctx context.Context) (*GoogleOIDC, error) {
	provider, err := oidc.NewProvider(ctx, "https://accounts.google.com")
	if err != nil {
		return nil, err
	}

	return &GoogleOIDC{
		Config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Endpoint:     provider.Endpoint(),
			Scopes: []string{
				"openid", "email", "profile",
			},
		},
		Verifier: provider.Verifier(&oidc.Config{
			ClientID: clientID,
		}),
	}, nil
}
