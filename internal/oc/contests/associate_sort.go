package contests

import "github.com/ocontest/backend/pkg/structs"

type SortByOther struct {
	mainSlice  []int
	otherSlice []structs.ScoreboardUserStanding
}

func (sbo SortByOther) Len() int {
	return len(sbo.mainSlice)
}

func (sbo SortByOther) Swap(i, j int) {
	sbo.mainSlice[i], sbo.mainSlice[j] = sbo.mainSlice[j], sbo.mainSlice[i]
	sbo.otherSlice[i], sbo.otherSlice[j] = sbo.otherSlice[j], sbo.otherSlice[i]
}

func (sbo SortByOther) Less(i, j int) bool {
	return sbo.mainSlice[i] < sbo.mainSlice[j]
}
