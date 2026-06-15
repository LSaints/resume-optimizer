package auth

type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expiresAt"`
}
