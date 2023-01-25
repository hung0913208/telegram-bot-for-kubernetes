package telegram

type Update struct {
	UpdateID int     `json:"update_id"`
	Message  Message `json:"message"`
}

type Message struct {
	Text     string   `json:"text"`
	Chat     Chat     `json:"chat"`
	Audio    Audio    `json:"audio"`
	Voice    Voice    `json:"voice"`
	Document Document `json:"document"`
}

// Audio message has extra attributes
type Audio struct {
	FileId   string `json:"file_id"`
	Duration int    `json:"duration"`
}

type Voice Audio

type Document struct {
	FileId   string `json:"file_id"`
	FileName string `json:"file_name"`
}

type Chat struct {
	ID int `json:"id"`
}
