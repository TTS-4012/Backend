package structs

type User struct {
	ID                int
	Username          string
	EncryptedPassword string
	Email             string
	Verified          bool
}
