package role

type Role string

const (
	SuperAdmin      Role = "super_admin"
	ClientAdmin     Role = "client_admin"
	TournamentStaff Role = "tournament_staff"
)
