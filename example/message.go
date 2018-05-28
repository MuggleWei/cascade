package example

type CommonMessage struct {
	Op   string      `json:"op"`
	Data interface{} `json:"data,omitempty"`
}

type LoginMessage struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

type LogoutMessage struct {
	User string `json"user"`
}

type GreetMessage struct {
	Greet string `json:"greet"`
}
