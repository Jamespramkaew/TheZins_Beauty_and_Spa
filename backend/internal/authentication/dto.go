package authentication

type LoginRequest struct {
	Email		string	`json:"username"`
	Password	string	`json:"password"`
}
