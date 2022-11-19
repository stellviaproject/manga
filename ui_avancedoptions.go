package main

import (
	"fyne.io/fyne/v2/widget"
)

var genres = map[string][]string{
	"4-Koma": []string{"https://es.ninemanga.com/category/4-Koma.html"},
	"Acción": []string{"https://es.ninemanga.com/category/Acci%26oacute%3Bn.html", "https://es.ninemanga.com/category/Acci%C3%B3n.html"},
	"Adulto": []string{},
	"Aventura": []string{},
	"Animación": []string{},
	"Apocalítico": []string{},
	"Artes Marciales": []string{},
	"Boys Love (BL)": []string{},
	"Ciberpunk": []string{},
	"Ciencia Ficción": []string{},
	"Comedia": []string{},
	"Crimen": []string{},
	"Demonios": []string{},
	"Deporte": []string{},
	"Doujinshi": []string{},
	"Drama": []string{},
	"Ecchi": []string{},
	"Escolar": []string{},
	"Español": []string{},
	"Extranjero": []string{},
	"Familia": []string{},
	"Fantasía": []string{},
	"Bender": []string{},
	"Girs Love (GL)": []string{},
	"Gore": []string{},
	"Guerra": []string{},
	"Harem": []string{},
	"Hentai": []string{},
	"Historico": []string{},
	"Horror": []string{},
	"Isekai": []string{},
	"Josei": []string{},
	"Maduro": []string{},
	"Magia": []string{},
	"Magical Girls": []string{},
	"Mecha": []string{},
	"Medical": []string{},
	"Militar": []string{},
	"Misterio": []string{},
	"Música": []string{},
	"Niños": []string{},
	"Oeste": []string{},
	"One Shot": []string{},
	"Parodia": []string{},
	"Filosófico": []string{},
	"Policiaco": []string{},
	"Psicológico": []string{},
	"Realidad Virtual": []string{},
	"Reencarnación": []string{},
	"Samurai": []string{},
	"School Life": []string{},
	"Seinen": []string{},
	"Shoujo": []string{},
	"Shojo Ai": []string{},
	"Yuri Soft": []string{},
	"Shonen": []string{},
	"Shonen-Ai": []string{},
	"Yaoi": []string{},
	"Shota": []string{},
	"Slice Of Life": []string{},
	"Smut": []string{},
	"Sobrenatural": []string{},
	"Super Poderes": []string{},
	"Supervivencia": []string{},
	"Supenso": []string{},
	"Thiller": []string{},
	"Tragedia": []string{},
	"Vampiros": []string{},
	"Ver en Lectorma": []string{},
	"Webcomic": []string{},
	"Webtoon": []string{},
	"Wuxia": []string{},
	"Yaoi Soft": []string{},
	"Yuri": []string{},
}

type UIAvancedOptions struct {
}

func NewUIAvancedOptions() *UIAvancedOptions {
	return &UIAvancedOptions{}
}

func (ui *UIAvancedOptions) MakeUI() *widget.Accordion {
	items := make([]string, 0)
	for key, _ := range genres {
		items = append(items, key)
	}
	genres := widget.NewRadioGroup(items, ui.OnChangeForGenres)
	return widget.NewAccordion([]*widget.AccordionItem{
		widget.NewAccordionItem("Géneros", genres),
	}...)
}

func (ui *UIAvancedOptions) OnChangeForGenres(selected string) {
	controller.search.SetList(genres[selected][0])
}