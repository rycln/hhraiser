package domain

type Session struct {
	xsrf  string
	token string
}

func NewSession(xsrf, token string) *Session {
	return &Session{
		xsrf:  xsrf,
		token: token,
	}
}

func (s *Session) IsAuthenticated() bool {
	return s.xsrf != "" && s.token != ""
}

func (s *Session) GetXSRF() string {
	return s.xsrf
}

func (s *Session) GetToken() string {
	return s.token
}
