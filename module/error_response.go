package module

type Error struct {
	Code    string `json:"error"`
	Message string `json:"message"`
	Help    string `json:"help"`
}

type Errors []Error
