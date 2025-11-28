package domain

type User struct {
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	Name         string `json:"name"`
}
