package domain

type Resume struct {
	id    string
	title string
}

type HH struct {
	phone     string
	password  string
	userAgent string
	xsrf      string
	token     string
	resume    *Resume
}
