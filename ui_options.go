package main

type UIOptions struct {
}

func NewUIOptions() *UIOptions {
	return &UIOptions{}
}

func (ui *UIOptions) OnSelect(option string) {
	//"Todos", "Populares", "Nuevos", "Actualizados", "Generos", "Años", "Alfabético", "Estado"
	switch option {
	case "Todos":
		controller.search.SetList(CATEGORY)
	case "Populares":
		controller.search.SetList(POPULAR)
	case "Nuevos":
		controller.search.SetList(NUEVOS)
	case "Recientes":
		controller.search.SetList(RECIENTES)
	}
}
