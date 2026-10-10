package permission

import "sport-grid-be/internal/role"

func ForRole(userRole role.Role) []string {
	switch userRole {
	case role.SuperAdmin:
		return All()
	case role.ClientAdmin:
		return []string{CreateUser, ViewUser}
	default:
		return nil
	}
}

func All() []string {
	return []string{
		CreateUser, UpdateUser, ViewUser, DeleteUser,
		CreateClient, UpdateClient, ViewClient, DeleteClient, ActivateClient, SuspendClient,
		UpdatePlayer, ViewPlayer, DeletePlayer,
		CreateSport, UpdateSport, ViewSport, DeleteSport, SetSportStatus,
		CreateTournament, UpdateTournament, ViewTournament, PublishTournament,
		OpenTournamentRegistration, CloseTournamentRegistration, StartTournament,
		CompleteTournament, CancelTournament,
		CreateDivision, UpdateDivision, ViewDivision, DeleteDivision,
		ActivateDivision, DeactivateDivision,
		CreateRegistration, ViewRegistration, ApproveRegistration, CancelRegistration,
		CreatePaymentSubmission, ViewPaymentSubmission, VerifyPaymentSubmission,
		RejectPaymentSubmission,
		CreateTeam, ViewTeam, DisbandTeam, AddTeamMember, RemoveTeamMember,
		CreateCourt, UpdateCourt, ViewCourt, SetCourtStatus,
		AssignSeed, GenerateSeeding, ViewSeeding,
		GenerateBracket, ViewBracket, ResetBracket,
		GenerateMatches, ViewMatch, ScheduleMatch, StartMatch, RecordMatchGame,
		FinalizeMatch, CancelMatch,
	}
}
