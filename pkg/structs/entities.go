package structs

type User struct {
	ID                int64
	Username          string
	EncryptedPassword string
	Email             string
	Verified          bool
}

type ProblemDescription struct {
	ID          string
	Description string
}
