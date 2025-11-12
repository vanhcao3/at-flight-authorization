package types

type UsersQuery struct {
	Exact     string `json:"exact"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Search    string `json:"search"`
	First     string `json:"first"`
	Max       string `json:"max"`
	Enabled   string `json:"enabled"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type UserMetadata struct {
	Username string `json:"username"`
	Id       string `json:"id"`
}
