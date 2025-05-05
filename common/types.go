package common

type Payload struct {
	UserID    int     `json:"user_id"`
	Total     float64 `json:"total"`
	Title     string  `json:"title"`
	Meta      Meta    `json:"meta"`
	Completed bool    `json:"completed"`
}

type Meta struct {
	Logins       []Login           `json:"logins"`
	PhoneNumbers map[string]string `json:"phone_numbers"`
}

type Login struct {
	Time string `json:"time"`
	IP   string `json:"ip"`
}
