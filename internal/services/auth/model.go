package auth

// Register DTOs
type RegisterResult struct {
	Account      *Account
	AccessToken  string
	RefreshToken string
}

type AccountDTO struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
}

type RegisterRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type RegisterResponse struct {
	Account     AccountDTO `json:"account"`
	AccessToken string     `json:"access_token"`
}
