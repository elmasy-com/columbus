package frontend

type heroData struct {
	Title    string
	Subtitle string
}

func getHeroData(title string, subtitle string) heroData {
	return heroData{title, subtitle}
}
