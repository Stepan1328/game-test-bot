package game_logic

func CheckingWinner(playField [3][3]int, move int) (bool, int) {
	victory := false

	player := (move-1)%2 + 1

	for i := 0; i < 3; i++ {
		if playField[i][0] == player && playField[i][1] == player && playField[i][2] == player {
			victory = true
			return victory, player
		} else if playField[0][i] == player && playField[1][i] == player && playField[2][i] == player {
			victory = true
			return victory, player
		}
	}

	if playField[0][0] == player && playField[1][1] == player && playField[2][2] == player {
		victory = true
		return victory, player
	}

	if playField[0][2] == player && playField[1][1] == player && playField[2][0] == player {
		victory = true
		return victory, player
	}

	return victory, player
}
