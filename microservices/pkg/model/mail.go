package model

type Attachment struct {
	Filename string `json:"filename"`
	Data     []byte `json:"data"`
}

type Mail struct {
	From        string       `json:"from"`
	To          []string     `json:"to"`
	Subject     string       `json:"subject"`
	Body        string       `json:"body"`
	Attachments []Attachment `json:"attachments"`
}
