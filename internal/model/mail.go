package model

type Attachment struct {
	Filename string
	Path     string
}

type Mail struct {
	From        string
	To          []string
	Subject     string
	Body        string
	Attachments []Attachment
}
