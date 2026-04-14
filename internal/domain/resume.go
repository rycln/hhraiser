package domain

type Resume struct {
	ID    string
	Title string
}

func NewResume(id, title string) *Resume {
	return &Resume{ID: id, Title: title}
}
