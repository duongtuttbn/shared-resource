package dt

type AuthParam struct {
	AuthUserID UserID `form:"-" json:"-" swag:"-"`
}
