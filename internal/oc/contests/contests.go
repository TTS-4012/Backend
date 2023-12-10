package contests

import "ocontest/internal/db"

type ContestsHandler interface {
	CreateContest()
	GetContest()
	ListContests()
	UpdateContest()
	DeleteContest()
}

type ContestsHandlerImp struct {
	ContestRepo db.ContestsMetadataRepo
}
