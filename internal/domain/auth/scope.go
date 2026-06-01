package authdomain

type Scope string

const (
	ScopeServerRead      Scope = "server:read"
	ScopeServerCreate    Scope = "server:create"
	ScopeServerUpdate    Scope = "server:update"
	ScopeServerDelete    Scope = "server:delete"
	ScopeServerImport    Scope = "server:import"
	ScopeServerExport    Scope = "server:export"
	ScopeServerReport    Scope = "server:report"
	ScopreReportDownload Scope = "report:download"

	ScopeUserRead Scope = "user:read"
)
