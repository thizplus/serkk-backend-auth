package dto

type OAuthURLResponse struct {
	AuthURL string `json:"auth_url"`
}

type OAuthCallbackRequest struct {
	Code  string `json:"code" validate:"required"`
	State string `json:"state"`
}

type OAuthLoginResponse struct {
	AccessToken string       `json:"access_token"`
	TokenType   string       `json:"token_type"`
	ExpiresIn   int          `json:"expires_in"`
	User        UserResponse `json:"user"`
	IsNewUser   bool         `json:"is_new_user"`
}

// Provider-specific user info structures

type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

type FacebookUserInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Picture struct {
		Data struct {
			URL string `json:"url"`
		} `json:"data"`
	} `json:"picture"`
}

type LINEUserInfo struct {
	UserID        string `json:"userId"`
	DisplayName   string `json:"displayName"`
	PictureURL    string `json:"pictureUrl"`
	StatusMessage string `json:"statusMessage"`
}

type LINEIDToken struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}
