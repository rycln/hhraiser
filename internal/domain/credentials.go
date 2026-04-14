package domain

type Credentials struct {
	phone    string
	password string
}

func NewCredentials(phone, password string) *Credentials {
	return &Credentials{phone: phone, password: password}
}

func (c *Credentials) GetPhone() string {
	return c.phone
}

func (c *Credentials) GetPassword() string {
	return c.password
}
