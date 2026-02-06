package domain

type User struct {
	ID           int64
	Username     string
	PasswordHash string
	IsEnabled    bool
	Roles        []string
	Permissions  []string
}
