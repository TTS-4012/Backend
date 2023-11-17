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

type Problem struct {
	CreatedBy   int64
	ID          int64
	Title       string
	DocumentID  string
	SolvedCount int64
	Hardness    int64
}
