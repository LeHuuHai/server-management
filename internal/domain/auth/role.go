package authdomain

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
	RoleGuest Role = "guest"
)

func (r Role) String() string {
	return string(r)
}

func (r Role) Scopes() []Scope {
	switch r {
	case RoleAdmin:
		return []Scope{}

	case RoleUser:
		return []Scope{}

	case RoleGuest:
		return []Scope{}

	default:
		return []Scope{}
	}
}
