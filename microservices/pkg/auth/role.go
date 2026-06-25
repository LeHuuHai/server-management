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
		return []Scope{
			ScopeServerRead,
			ScopeServerCreate,
			ScopeServerUpdate,
			ScopeServerDelete,
			ScopeServerImport,
			ScopeServerExport,
			ScopeServerReport,
			ScopeReportDownload,
			ScopeUserRead,
		}

	case RoleUser:
		return []Scope{
			ScopeServerRead,
			ScopeServerCreate,
			ScopeServerUpdate,
			ScopeServerDelete,
			ScopeServerImport,
			ScopeServerExport,
			ScopeServerReport,
			ScopeReportDownload,
		}

	case RoleGuest:
		return []Scope{
			ScopeServerRead,
		}

	default:
		return []Scope{}
	}
}
