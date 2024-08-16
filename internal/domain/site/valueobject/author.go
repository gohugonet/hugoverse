package valueobject

type Author struct {
	name  string
	email string
}

func NewAuthor(name, email string) *Author {
	return &Author{
		name:  name,
		email: email,
	}
}

func (a *Author) Email() string {
	return a.email
}

func (a *Author) Name() string {
	return a.name
}
