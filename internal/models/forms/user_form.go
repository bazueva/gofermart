package forms

type UserForm struct {
	Login    string `validate:"required,min=4,max=20"`
	Password string `validate:"required,min=8,max=32"`
}

type LoginForm struct {
	Login    string `validate:"required,min=4,max=20"`
	Password string `validate:"required,min=8,max=32"`
}
