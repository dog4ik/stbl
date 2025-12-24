package connect

type Settings struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Sandbox  bool   `json:"sandbox"`
}
