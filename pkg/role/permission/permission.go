package permission

const (
	// User Permissions
	CreateUser = "create.user"
	UpdateUser = "update.user"
	ViewUser   = "view.user"
	DeleteUser = "delete.user"

	// Client Permissions
	CreateClient = "create.client"
	UpdateClient = "update.client"
	ViewClient   = "view.client"
	DeleteClient = "delete.client"

	// Player Permissions
	UpdatePlayer = "update.player"
	ViewPlayer   = "view.player"
	DeletePlayer = "delete.player"

	// Sport Permissions
	CreateSport = "create.sport"
	UpdateSport = "update.sport"
	ViewSport   = "view.sport"
	DeleteSport = "delete.sport"

	// Division Permissions
	CreateDivision = "create.division"
	UpdateDivision = "update.division"
	ViewDivision   = "view.division"
	DeleteDivision = "delete.division"

	// Registration Permissions
	CreateRegistration  = "create.registration"
	ViewRegistration    = "view.registration"
	ApproveRegistration = "approve.registration"
	CancelRegistration  = "cancel.registration"

	// Payment Submission Permissions
	CreatePaymentSubmission = "create.payment_submission"
	VerifyPaymentSubmission = "verify.payment_submission"
	RejectPaymentSubmission = "reject.payment_submission"

	// Team Permissions
	CreateTeam       = "create.team"
	ViewTeam         = "view.team"
	DisbandTeam      = "disband.team"
	AddTeamMember    = "add.team_member"
	RemoveTeamMember = "remove.team_member"

	// Court Permissions
	CreateCourt    = "create.court"
	UpdateCourt    = "update.court"
	ViewCourt      = "view.court"
	SetCourtStatus = "set_status.court"

	// Seeding Permissions
	AssignSeed      = "assign.seed"
	GenerateSeeding = "generate.seeding"
	ViewSeeding     = "view.seeding"

	// Bracket Permissions
	GenerateBracket = "generate.bracket"
	ViewBracket     = "view.bracket"
	ResetBracket    = "reset.bracket"

	// Match Permissions
	GenerateMatches = "generate.matches"
	ViewMatch       = "view.match"
	ScheduleMatch   = "schedule.match"
	StartMatch      = "start.match"
	RecordMatchGame = "record.match_game"
	FinalizeMatch   = "finalize.match"
	CancelMatch     = "cancel.match"
)
