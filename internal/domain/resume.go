package domain

type Resume struct {
	id    string
	title string
}

func NewResume(id, title string) *Resume {
	return &Resume{id: id, title: title}
}

func (r *Resume) GetID() string {
	return r.id
}

func (r *Resume) GetTitle() string {
	return r.title
}
