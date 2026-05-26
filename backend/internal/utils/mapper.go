package utils

var teamTranslations = map[string]string{
	"Brazil":             "Brasil",
	"Argentina":          "Argentina",
	"Algeria":            "Argélia",
	"Australia":          "Austrália",
	"Austria":            "Áustria",
	"Belgium":            "Bélgica",
	"Bosnia-Herzegovina": "Bósnia e Herzegovina",
	"Canada":             "Canadá",
	"Cape Verde Islands": "Cabo Verde",
	"Colombia":           "Colômbia",
	"Congo DR":           "Rep. Demoncratica do Congo",
	"Croatia":            "Croácia",
	"Curaçao":            "Curação",
	"Czechia":            "República Tcheca",
	"Ecuador":            "Equador",
	"Egypt":              "Egito",
	"England":            "Inglaterra",
	"France":             "França",
	"Germany":            "Alemanha",
	"Ghana":              "Gana",
	"Haiti":              "Haiti",
	"Iran":               "Irã",
	"Iraq":               "Iraque",
	"Ivory Coast":        "Costa do Marfim",
	"Japan":              "Japão",
	"Jordan":             "Jordânia",
	"Mexico":             "México",
	"Morocco":            "Marrocos",
	"Netherlands":        "Holanda",
	"New Zealand":        "Nova Zelândia",
	"Noway":              "Noruega",
	"Norway":             "Noruega",
	"Panama":             "Panamá",
	"Paraguay":           "Paraguai",
	"Portugal":           "Portugal",
	"Qatar":              "Catar",
	"Saudi Arabia":       "Arábia Saudita",
	"Scotland":           "Escócia",
	"Senegal":            "Senegal",
	"South Africa":       "África do Sul",
	"South Korea":        "Coreia do Sul",
	"Spain":              "Espanha",
	"Switzerland":        "Suíça",
	"Sweden":             "Suécia",
	"Tunisia":            "Tunísia",
	"Turkey":             "Turquia",
	"United States":      "Estados Unidos",
	"Uruguay":            "Uruguai",
	"Uzbekistan":         "Uzbequistão",
}

func TranslateTeam(name string) string {
	if translated, ok := teamTranslations[name]; ok {
		return translated
	}

	return name
}
