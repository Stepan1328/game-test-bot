package game_logic

import (
	"math/rand"
)

type Situation struct {
	PlayField [3][3]int
}

type Action struct {
	Y int
	X int
}

func (s Situation) Analyze(player int, motion int) (Action, int) {
	winMoves := make([]Action, 0)
	drawMoves := make([]Action, 0)
	losingMoves := make([]Action, 0)

	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if s.PlayField[i][j] == 0 && (motion-1)%2+1 == player {
				s.PlayField[i][j] = (motion-1)%2 + 1

				win, _ := CheckingWinner(s.PlayField, motion)
				if win {
					winMoves = append(winMoves, Action{i, j})
				}

				s.PlayField[i][j] = 0
			}
		}
	}

	if len(winMoves) > 0 {
		return winMoves[rand.Intn(len(winMoves))], 2
	}

	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if s.PlayField[i][j] == 0 {
				s.PlayField[i][j] = (motion-1)%2 + 1

				win, winPlayer := CheckingWinner(s.PlayField, motion)
				if win && winPlayer == player {
					winMoves = append(winMoves, Action{i, j})
				} else if win {
					losingMoves = append(losingMoves, Action{i, j})
				} else if motion == 9 {
					drawMoves = append(drawMoves, Action{i, j})
				} else {
					move, result := s.Analyze(player, motion+1)

					switch result {
					case 2:
						winMoves = append(winMoves, move)
					case 1:
						drawMoves = append(drawMoves, move)
					case 0:
						losingMoves = append(losingMoves, move)
					}
				}

				s.PlayField[i][j] = 0
			}
		}
	}

	if player == (motion+1)%2+1 {
		if len(winMoves) > 0 {
			return winMoves[rand.Intn(len(winMoves))], 2
		}

		if len(drawMoves) > 0 {
			return drawMoves[rand.Intn(len(drawMoves))], 1
		}

		return losingMoves[rand.Intn(len(losingMoves))], 0
	} else {
		if len(losingMoves) > 0 {
			return losingMoves[rand.Intn(len(losingMoves))], 2
		}

		if len(drawMoves) > 0 {
			return drawMoves[rand.Intn(len(drawMoves))], 1
		}

		return winMoves[rand.Intn(len(winMoves))], 0
	}
}
