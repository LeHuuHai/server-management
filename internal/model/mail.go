package model

type Attachment struct {
	Filename string
	Data     []byte
}

type Mail struct {
	From        string
	To          []string
	Subject     string
	Body        string
	Attachments []Attachment
}
