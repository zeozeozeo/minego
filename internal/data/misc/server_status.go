package misc

type ServerStatusResponse struct {
	Description string              `json:"description"`
	Players     ServerStatusPlayers `json:"players"`
	Version     ServerStatusVersion `json:"version"`
}

type ServerStatusPlayers struct {
	Max    int `json:"max"`
	Online int `json:"online"`
}

type ServerStatusVersion struct {
	Name     string `json:"name"`
	Protocol int    `json:"protocol"`
}
