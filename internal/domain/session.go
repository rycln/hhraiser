package domain

type Session struct {
	xsrf    string
	hhtoken string
}

func NewSession(xsrf, token string) *Session {
	return &Session{
		xsrf:    xsrf,
		hhtoken: token,
	}
}

func (s *Session) IsAuthenticated() bool {
	return s.xsrf != "" && s.hhtoken != ""
}

func (s *Session) GetXSRF() string {
	return s.xsrf
}

func (s *Session) GetToken() string {
	return s.hhtoken
}
