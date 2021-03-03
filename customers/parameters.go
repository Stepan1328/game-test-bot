package customers

type UsersStatistic struct {
	RunGame      bool
	PlayingField [][]int
}

var (
	NilPlayField = [][]int{{0, 0, 0}, {0, 0, 0}, {0, 0, 0}}

	Defuser = UsersStatistic{
		RunGame:      false,
		PlayingField: NilPlayField,
	}
)
