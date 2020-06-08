package main

type Snail struct {
	Name     string
	Position int
	Score    int
	Adka     int
	Base     string
}

func (s *Snail) GetString() string {
	var out string
	if s.Position == winPos {
		out = "_________________________🐌🥇"
	} else {
		out = s.Base[:s.Position] + "🐌" + s.Base[s.Position:]

	}
	return out
}

func (s *Snail) Hodik() (bool, bool) {
	randomka := Random(0, 100)

	if randomka < changeSpeedProb {
		s.Adka = Random(1, 10)
	}

	s.Score += s.Adka
	//fmt.Println("Скоры "+s.Name+":", gary.Score)
	if s.Score > maxScore {
		s.Position++
		s.Score = 0

		if s.Position == winPos {
			return true, true
		}

		return true, false
	}
	return false, false
}
