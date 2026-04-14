package domain

type UserAgentGenerator interface {
	Generate() string
}

type TokenStorage interface {
	SaveTokens(xsrf, hhtoken string) error
	LoadTokens() (xsrf, hhtoken string, err error)
}

type HH struct {
	phone    string
	password string
	uaGen    UserAgentGenerator
	storage  TokenStorage
	xsrf     string
	token    string
}

func NewHH(phone string, password string, uaGen UserAgentGenerator, storage TokenStorage) *HH {
	return &HH{
		phone:    phone,
		password: password,
		uaGen:    uaGen,
		storage:  storage,
	}
}

func (h *HH) IsAuthorized() bool {
	return h.token != "" && h.xsrf != ""
}
