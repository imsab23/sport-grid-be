# Sport Grid Backend PRD Audit

**Audit date:** 2026-10-10  
**Audited repository:** `D:\Coding\sport-grid-be`  
**Audited commit:** `fc3f3b93492e80ae56b1db0b9daa014200ec4801`  
**Audited state:** working tree, including the uncommitted `internal/` rearchitecture, `main.go`, modified wiring/dependencies, deleted tracked `pkg/` tree, and all untracked files present at audit time  
**Primary authority:** `C:\Users\basam\Downloads\SPORT_GRID_PRD.md` (approved 2026-10-09)  
**Secondary authority:** `.agents/skills/sport-grid-business-flow/`  
**Engineering guidance:** `.cursor/skills/go-coding-standards/` and `.cursor/skills/go-folder-structure/`

This is an implementation audit. No backend source, route, migration, or test was changed. The only repository change made by this task is this report.

## 1. Executive Summary

The backend is **not ready for complete PRD-compliant CMS integration**. It provides a useful skeleton and several functioning workflows, but core authority, participant locking, competition formats, result authority, and reporting are incomplete or incorrect. Of 78 PRD requirements, 22 were fully implemented at baseline, 5 remediated findings are now verified, 24 remain partially implemented, 12 are not implemented, 5 are implemented incorrectly, 3 have a primary security gap, 4 have a primary data-model gap, 2 have a primary API-contract gap, and 1 is an out-of-scope implementation.

Strong areas are JWT login/refresh mechanics, basic global Player identity, tournament-owned resource scoping in several controllers, division and court CRUD, individual registration capacity locking, manual payment verification transactions, team membership operations, match scheduling persistence, match game recording, and single-elimination winner propagation.

The highest risks are:

1. **Cross-tenant and privilege escalation exposure.** Client and staff-user APIs lack adequate actor authority and tenant enforcement. A general CMS User can enumerate or create platform clients and users, and role assignment accepts arbitrary strings.
2. **Core runtime correctness failures.** Several SQL predicates compare integer status columns with strings. Bracket status uses Go strings against a `SMALLINT` database column. Player profile updates write an absent `updated_by` column.
3. **Missing competition foundations.** Check-in, participant-set locking, full team participation, double elimination, round robin, group stages, standings, loser progression, and authoritative result correction do not exist.
4. **Lifecycle gaps.** Tournament transitions do not enforce configuration, participant, or match preconditions; archive and competition-preparation states are absent.
5. **Incomplete payment and audit model.** Payment settings are not structured, review history is not append-only, and there is no discoverable review queue API.
6. **Weak scalability contracts.** List services do not pass requested pagination into the platform search helper, total counts remain unavailable, and high-volume filters have sparse indexing.

Obvious scope drift includes arbitrary Sport CRUD/status mutation even though the approved product treats Sports as predefined platform configuration. A Swiss format enum exists without PRD approval or an implementation. Proof-reference storage and a password-reset table are present without complete product flows.

Before CMS integration expands, fix platform authority and tenant isolation, normalize enum/database contracts, make pagination truthful, complete tournament lifecycle prerequisites, then add participant locking. Competition, scheduling, results, and reporting should follow those foundations.

### Audit method and evidence standard

Each requirement was traced through route, controller, service/domain behavior, SQL, and migration where present. Directory presence and route existence were not accepted as implementation proof. Invalid transitions, cross-tenant access, concurrent effects, database constraints, and response contracts were considered. Missing product policy was not invented.

The `.cursor` guidance was applied as an engineering-quality lens. The current uncommitted architecture is only partly aligned: it is domain-oriented and generally keeps transport separate, but several domain services and SQL live directly in `internal/<domain>/` rather than `internal/<domain>/service/`; transport and migrations remain under `internal/`; permissions are defined but not enforced; process-local mutable refresh state conflicts with production guidance; money uses `float64`; and the test suite does not cover required tenant, lifecycle, concurrency, or SQL behavior. These observations do not change PRD counts unless they directly prevent a requirement.

## 2. PRD Coverage Summary

### Overall classification

| Classification | Count | Percent |
|---|---:|---:|
| `FULLY_IMPLEMENTED` | 22 | 28.2% |
| `VERIFIED` | 5 | 6.4% |
| `PARTIALLY_IMPLEMENTED` | 24 | 30.8% |
| `NOT_IMPLEMENTED` | 12 | 15.4% |
| `IMPLEMENTED_INCORRECTLY` | 5 | 6.4% |
| `SECURITY_GAP` | 3 | 3.8% |
| `DATA_MODEL_GAP` | 4 | 5.1% |
| `API_CONTRACT_GAP` | 2 | 2.6% |
| `OUT_OF_SCOPE_IMPLEMENTATION` | 1 | 1.3% |
| **Total** | **78** | **100%** |

`OVER_IMPLEMENTED`, `DUPLICATE_IMPLEMENTATION`, `AMBIGUOUS_REQUIREMENT`, and `DOCUMENTATION_CONFLICT` are not primary classifications for any PRD row. Relevant non-PRD findings and unresolved subrules are inventoried in Sections 24–26 so coverage totals remain exact.

### Domain scorecard

| Domain | Req. | Full | Verified | Partial | Missing | Incorrect | Security | Data model | API contract | OOS | Confidence |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| Authentication | 6 | 2 | 0 | 2 | 1 | 0 | 0 | 0 | 1 | 0 | High |
| Clients / tenants | 5 | 2 | 1 | 2 | 0 | 0 | 0 | 0 | 0 | 0 | High |
| Users / roles | 6 | 1 | 4 | 0 | 1 | 0 | 0 | 0 | 0 | 0 | High |
| Players | 6 | 4 | 0 | 1 | 0 | 0 | 0 | 1 | 0 | 0 | High |
| Platform configuration | 4 | 0 | 0 | 2 | 0 | 0 | 1 | 0 | 0 | 1 | High |
| Tournaments | 6 | 3 | 0 | 3 | 0 | 0 | 0 | 0 | 0 | 0 | High |
| Divisions / eligibility | 4 | 1 | 0 | 2 | 0 | 1 | 0 | 0 | 0 | 0 | High |
| Registrations | 6 | 2 | 0 | 4 | 0 | 0 | 0 | 0 | 0 | 0 | High |
| Payments | 6 | 2 | 0 | 2 | 0 | 0 | 1 | 0 | 1 | 0 | High |
| Teams | 4 | 0 | 0 | 1 | 1 | 1 | 0 | 1 | 0 | 0 | High |
| Check-in / locking | 3 | 0 | 0 | 0 | 2 | 0 | 0 | 1 | 0 | 0 | High |
| Competition | 5 | 0 | 0 | 3 | 0 | 2 | 0 | 0 | 0 | 0 | High |
| Scheduling | 3 | 1 | 0 | 1 | 0 | 1 | 0 | 0 | 0 | 0 | High |
| Matches / scoring | 5 | 3 | 0 | 0 | 1 | 0 | 1 | 0 | 0 | 0 | High |
| Results / advancement | 5 | 0 | 0 | 1 | 3 | 0 | 0 | 1 | 0 | 0 | High |
| Reporting / auditability | 4 | 1 | 0 | 0 | 3 | 0 | 0 | 0 | 0 | 0 | High |

### Readiness by product flow

| Flow | Readiness | Blocking reason |
|---|---|---|
| Sign in and refresh | Partial | No logout/revoke endpoint, incomplete identity response, process-local refresh store |
| Client administration | Unsafe | Missing Super Admin enforcement and client-scoped administration |
| Staff administration | Unsafe | Arbitrary role assignment and cross-tenant user access |
| Player identity | Partial | Global identity is sound; update schema mismatch and no approved status policy |
| Tournament setup | Partial | Basic entity and divisions exist; lifecycle prerequisites and formats are incomplete |
| Registration | Partial | Window/capacity exist; eligibility, waitlist, and numeric SQL predicates are incomplete/incorrect |
| Manual payment | Partial | Submit/verify/reject exist; review discovery, settings, permissions, and history are incomplete |
| Team competition | Not ready | Team registration/generation path and membership constraints are incomplete |
| Competition generation | Not ready | Only individual single elimination materially works, with schema/status defects |
| Scheduling | Partial | Assignments persist; conflicts and officials are missing |
| Scoring/results | Partial | Games and declared winners exist; rules, correction, standings, loser progression are missing |
| Reporting/audit | Not ready | No report endpoints and no append-only sensitive-operation history |

## 3. Requirements Traceability Matrix

Priority appears only for gaps. Dependencies name prerequisites that must be remediated first.

| Requirement ID | PRD Requirement | Expected Backend Behavior | Actual Implementation | Status | Evidence | Gap / Recommendation |
|---|---|---|---|---|---|---|
| AUTH-001 | Authenticate approved account types | Verify credentials and issue identity-bound tokens for staff and Players | User and Player login paths issue access/refresh tokens; response omits most identity context and session persistence is process-local | `PARTIALLY_IMPLEMENTED` | `internal/auth/auth.go:84-140`; `internal/auth/model.go:48-54` | **P0.** Return a stable identity projection and persist/coordinate refresh sessions; depends on approved auth contract |
| AUTH-002 | CMS can obtain current authenticated identity | Provide a protected current-user endpoint or equivalent complete login response | No `/me` endpoint; login response only includes first name and tokens | `API_CONTRACT_GAP` | `internal/transport/http/auth_controller.go:11-16`; `internal/auth/model.go:48-54` | **P0.** Add one authoritative current-identity contract after role/tenant model is fixed |
| AUTH-003 | Refresh tokens rotate securely | Validate, revoke, rotate, and reject replay | Refresh is serialized, old token is revoked, and a new pair is issued | `FULLY_IMPLEMENTED` | `internal/auth/auth.go:181-216`; `internal/auth/auth_test.go` | Preserve behavior; add distributed-store tests before multi-instance deployment |
| AUTH-004 | Sessions remain correct across instances/restarts | Persist revocation/rotation state or use an equivalent distributed design | Refresh session state is held in an in-memory map guarded by a mutex | `PARTIALLY_IMPLEMENTED` | `internal/auth/auth.go:40-68` | **P0.** Move session/revocation authority to Redis or durable storage; depends on token/session policy |
| AUTH-005 | Public Player registration creates a Player identity | Self-service registration must create the global Player domain, not staff User | Public registration creates a Player | `FULLY_IMPLEMENTED` | `internal/auth/auth.go:71-81`; `internal/auth/model.go:28-46` | Preserve Player/User separation |
| AUTH-006 | Authenticated sessions can be explicitly ended | Revoke the refresh session/token on logout | No logout or token-revocation route exists | `NOT_IMPLEMENTED` | `internal/transport/http/auth_controller.go:11-16` | **P0.** Add logout/revoke after durable session storage is established |
| CLT-001 | Super Admin manages Clients | Create, view, update, and control Client lifecycle with Super Admin authority | Create/search/get exist; activate/suspend service operations are not routed; update is absent | `PARTIALLY_IMPLEMENTED` | `internal/transport/http/client_controller.go:13-18`; `internal/client/client.go` | **P0.** Complete explicit lifecycle/update routes with Super Admin guard |
| CLT-002 | Client data is isolated by tenant | Client-scoped actors can access only their Client; platform operations require Super Admin | Client routes have no role or ownership middleware; general CMS users can create/list/get Clients | `SECURITY_GAP` | `internal/transport/http/client_controller.go:13-18`; `internal/transport/http/server.go:77-115` | **P0.** Enforce Super Admin for platform Client APIs and explicit client scope for allowed self-service operations |
| CLT-003 | Client status controls operational access | Suspended/inactive Client state must prevent prohibited operations | Status and service methods exist, but no route completes lifecycle and downstream requests do not consistently gate on Client status | `PARTIALLY_IMPLEMENTED` | `internal/client/model.go`; `internal/client/client.go` | **P1.** Centralize active-Client enforcement after lifecycle authority is fixed |
| CLT-004 | Client identity has stable unique slug | Persist normalized unique Client slug | Slug is generated and protected by a database uniqueness constraint | `FULLY_IMPLEMENTED` | `internal/client/model.go:51-68`; `internal/storage/postgres/migration/client.go:18-28` | Preserve constraint and normalization |
| CLT-005 | Client records expose ownership metadata | Persist status and timestamps needed for administration | Client schema/model include status and timestamps | `FULLY_IMPLEMENTED` | `internal/client/model.go`; `internal/storage/postgres/migration/client.go` | Add actor history separately; current entity metadata is adequate |
| USR-001 | Staff Users are distinct from Players | Maintain a staff/admin identity domain separate from global Player accounts | Separate User and Player models, stores, and login lookup exist | `FULLY_IMPLEMENTED` | `internal/user/model.go`; `internal/player/model.go`; `internal/auth/auth.go:84-140` | Preserve the boundary |
| USR-002 | Only authorized actors create staff Users | Super Admin/Client Admin authority must be checked server-side | User routes lack role middleware; intended authority block is commented out | `SECURITY_GAP` | `internal/transport/http/user_controller.go:16-42` | **P0.** Enforce actor authority in service/domain boundary and route middleware |
| USR-003 | Client Admin cannot assign equal/higher authority | Role assignment must prevent privilege escalation and remain in tenant | Create accepts any non-empty role string; no hierarchy validation | `SECURITY_GAP` | `internal/user/model.go:91-124`; `internal/transport/http/user_controller.go:24-42` | **P0.** Introduce predefined typed roles and an assignment policy before exposing staff creation |
| USR-004 | Staff lists/details are tenant-scoped | Client-scoped actors may only read their Client’s staff | Search trusts optional query scope and does not force actor Client; detail lacks ownership enforcement | `SECURITY_GAP` | `internal/transport/http/user_controller.go:16-21`; `internal/user/service/user_service.go` | **P0.** Derive scope from verified identity; never trust request `client_id` for Client Admin |
| USR-005 | Create/update returns integration-useful User data | Mutation response should return the created/updated resource or documented stable projection | User create returns a success message only | `API_CONTRACT_GAP` | `internal/transport/http/user_controller.go:49-55` | **P0.** Return canonical User projection after authority and sensitive-field policy are set |
| USR-006 | Staff User status can be managed explicitly | Provide validated activate/deactivate/suspend transitions | Command/model exists, service method is commented, and no route is mounted | `NOT_IMPLEMENTED` | `internal/user/user.go:9-18`; `internal/user/model.go:175-191`; `internal/transport/http/user_controller.go` | **P0.** Implement explicit status transition with tenant/role checks and audit history |
| PLY-001 | Player is global across Clients | One Player identity may participate across Client tournaments | Player has no Client ownership and registrations provide tournament/division scope | `FULLY_IMPLEMENTED` | `internal/player/model.go:20-40`; `internal/storage/postgres/migration/player.go`; `internal/storage/postgres/migration/registration.go` | Preserve global identity model |
| PLY-002 | Authorized CMS actors can find Players safely | Search must be paginated, permission-aware, and privacy-appropriate | Admin list/search exists, but any CMS User can list global Players and requested pagination is not propagated | `PARTIALLY_IMPLEMENTED` | `internal/transport/http/player_controller.go:30-80`; `internal/player/player.go` | **P1.** Define approved staff discovery scope, minimize PII, and fix pagination |
| PLY-003 | Player can maintain own profile | Verified Player identity may read/update its own profile | Self profile routes and identity checks exist | `FULLY_IMPLEMENTED` | `internal/transport/http/player_controller.go`; `internal/player/player.go` | Fix schema defect in PLY-005 before relying on update |
| PLY-004 | CMS cannot treat Player as staff User | Admin operations preserve domain distinction | Separate controllers/models and user types are maintained | `FULLY_IMPLEMENTED` | `internal/transport/http/server.go:77-115`; `internal/auth/model.go:21-26` | Preserve distinction in future roles and APIs |
| PLY-005 | Player updates persist against valid schema | All written fields must exist with correct ownership/audit semantics | Update writes `updated_by`, but Player migrations never add the column | `DATA_MODEL_GAP` | `internal/player/model.go:39,63`; `internal/storage/postgres/migration/player.go:10-128` | **P0.** Reconcile model/query/schema through a forward migration; decide whether actor audit belongs on entity or history |
| PLY-006 | Player credentials/profile data are validated | Validate required identity/profile inputs and protect credentials | Player commands validate required fields and password handling is encapsulated | `FULLY_IMPLEMENTED` | `internal/player/model.go`; `internal/auth/auth.go` | Expand validation only where PRD policy is explicit |
| CFG-001 | Roles are predefined platform configuration | Expose stable approved roles and permissions without arbitrary role CRUD | Three string role constants and unused permission constants exist; no canonical read contract | `PARTIALLY_IMPLEMENTED` | `internal/role/role.go:3-9`; `internal/role/permission/permission.go` | **P0.** Define typed approved roles, hierarchy, permissions, and read endpoint as one authority model |
| CFG-002 | Sports are predefined platform configuration | Sports should be stable platform data, not arbitrary tenant/product CRUD | Full Super Admin Sport create/update/status mutation API is exposed | `OUT_OF_SCOPE_IMPLEMENTATION` | `internal/transport/http/sport_controller.go:15-27`; `internal/sport/service/sport.go` | **P1.** Decide seed/config ownership, then remove or strictly constrain arbitrary mutations |
| CFG-003 | Platform configuration changes require correct authority | Only Super Admin may modify approved platform configuration | Mutation group is Super Admin guarded, but duplicate unguarded/guarded list registration and broader role model inconsistencies make authority ambiguous | `SECURITY_GAP` | `internal/transport/http/sport_controller.go:17,22-27`; `internal/transport/middleware.go:17-48` | **P0.** Remove duplicate route ownership and enforce one typed authority path |
| CFG-004 | Enum/API representation is consistent | Sport status/filter must use the same integer enum end-to-end | Entity/status are numeric, but search filters SQL on legacy `is_active` | `PARTIALLY_IMPLEMENTED` | `internal/sport/model.go:87-90`; `internal/storage/postgres/migration/sport.go` | **P0.** Replace legacy boolean filter with validated numeric status and contract tests |
| TRN-001 | Authorized organizers create Client-owned tournaments | Validate owner, active Client/Sport, required configuration, and tenant authority | Create validates active Sport/Client and defaults Draft; non-Super Admin Client ID is overwritten from identity | `PARTIALLY_IMPLEMENTED` | `internal/tournament/tournament.go:51-98`; `internal/transport/http/tournament_controller.go` | **P0.** Add permission policy and complete required field/config validation |
| TRN-002 | Tournament supports approved competition formats | Persist and validate only approved formats: single/double elimination, round robin, group stage, group+knockout | Enum has single, double, round robin, and unapproved Swiss; group formats are absent; engine supports only single elimination | `PARTIALLY_IMPLEMENTED` | `internal/tournament/model.go:40-55`; `internal/bracket/service/braket_service.go:107-135` | **P0.** Align enum with PRD only after participant locking and competition model are designed |
| TRN-003 | Tournament list/detail is tenant-scoped | Super Admin can access platform scope; Client actors only own Client | Controller resolves owned tournament and applies Client ownership checks | `FULLY_IMPLEMENTED` | `internal/transport/http/tournament_controller.go:81-143`; `internal/transport/middleware.go:33-48` | Preserve pattern and move invariant into reusable service policy where practical |
| TRN-004 | Lifecycle transitions enforce prerequisites | Publish/open/close/start/complete/cancel/archive must validate actor, current state, and business preconditions | Dedicated transitions check current state only; no configuration/participant/match preconditions; archive and preparation absent | `PARTIALLY_IMPLEMENTED` | `internal/tournament/tournament.go:178-230`; `internal/tournament/model.go:28-38` | **P0.** Implement formal state machine after required configuration and participant-lock rules exist |
| TRN-005 | General update cannot bypass lifecycle | Status is not accepted as an ordinary mutable field | Update model excludes status and explicit transition operations are used | `FULLY_IMPLEMENTED` | `internal/tournament/model.go`; `internal/tournament/tournament.go:141-230` | Preserve explicit transition boundary |
| TRN-006 | Tournament data includes core setup fields | Persist sport, owner, dates, venue, registration settings, fee, instructions, and format | Core model/schema fields are present | `FULLY_IMPLEMENTED` | `internal/tournament/model.go`; `internal/storage/postgres/migration/tournament.go` | Extend only for approved structured settings |
| DIV-001 | Tournament divisions/categories are configurable | Create/read/update/list divisions under an owned tournament | Nested CRUD with tournament ownership checks exists | `FULLY_IMPLEMENTED` | `internal/transport/http/division_controller.go:46-65`; `internal/division/division.go` | Preserve nested ownership pattern |
| DIV-002 | Division eligibility rules are configurable | Persist age/date-of-birth, gender, and skill rules using consistent types | Eligibility fields exist, but representation is incomplete and no registration enforcement consumes them | `PARTIALLY_IMPLEMENTED` | `internal/division/model.go`; `internal/storage/postgres/migration/tournament.go` | **P0.** Finalize rule model with timezone/date semantics; dependency for REG-001 |
| DIV-003 | Division lifecycle is controlled | Activate/deactivate/edit must respect tournament lifecycle and participant state | Edit guard blocks only completed/cancelled; activate/deactivate bypass parent edit guard | `PARTIALLY_IMPLEMENTED` | `internal/division/division.go:36-45,157-176` | **P0.** Enforce parent lifecycle and registration/lock preconditions in every mutation |
| DIV-004 | True enums use consistent integer contracts | Participant type and division status must align with integer enum standard | Both are string enums while database/API rules expect integer enums | `IMPLEMENTED_INCORRECTLY` | `internal/division/model.go:25-41`; division migration | **P0.** Migrate DB/model/request/response together with compatibility plan |
| REG-001 | Registration enforces eligibility | Validate Player against division age/gender/skill rules before acceptance | Registration verifies open tournament and active division, but does not evaluate eligibility fields | `PARTIALLY_IMPLEMENTED` | `internal/registration/service/registration.go:82-171`; `internal/division/model.go` | **P0.** Add authoritative eligibility evaluation after DIV-002 is settled |
| REG-002 | Registration respects window, capacity, and concurrency | Reject closed/out-of-window registration and atomically prevent over-capacity | Registration-open state and capacity row lock/count exist; explicit date-window enforcement is incomplete | `PARTIALLY_IMPLEMENTED` | `internal/registration/service/registration.go:82-171` | **P0.** Enforce registration dates and add concurrency integration tests |
| REG-003 | Duplicate active registration is prevented | A Player cannot hold duplicate active registration in one division | Duplicate check and database unique key exist | `FULLY_IMPLEMENTED` | `internal/registration/service/registration.go`; `internal/storage/postgres/migration/registration.go:33-38` | Reconcile cancellation/re-registration semantics under REG-004 |
| REG-004 | Registration lifecycle supports approved states | Pending/confirmed/waitlisted/cancelled/checked-in transitions must be explicit and valid | Only Pending, Confirmed, Cancelled exist; waitlist and check-in absent; unique key prevents re-registration after cancellation | `PARTIALLY_IMPLEMENTED` | `internal/registration/model.go:31-46`; registration migration | **P0.** Design full state machine and active-registration uniqueness before check-in |
| REG-005 | Registration ownership and nested scope are enforced | Staff operations must resolve tournament/division ownership; Player operations must use subject identity | Nested controllers resolve owned tournament and Player registration uses authenticated subject | `FULLY_IMPLEMENTED` | `internal/transport/http/registration_controller.go`; `internal/transport/http/middleware.go` | Replace broad role gate with permissions under REG-006 |
| REG-006 | Registration operations use approved staff permissions | Review/confirm/cancel actions require assigned capability | Routes allow broad `TournamentStaff` role; granular permission constants are unused | `PARTIALLY_IMPLEMENTED` | registration controller; `internal/role/permission/permission.go` | **P1.** Enforce capability-based actions after role/permission model is authoritative |
| PAY-001 | Player submits manual payment details | Persist amount, method, reference/instructions, and Pending review tied to registration | Submission exists and is transactional; structured organizer payment method/settings are absent | `PARTIALLY_IMPLEMENTED` | payment code in `internal/registration/service/registration.go`; registration migration | **P0.** Add approved payment-settings model before expanding submission contract |
| PAY-002 | Authorized reviewer verifies payment | Explicit idempotent review transition records verifier and confirms registration as approved | Verify operation is explicit and transactional | `FULLY_IMPLEMENTED` | `internal/registration/service/registration.go`; registration controller | Preserve atomicity; policy coupling is covered by PAY-006 |
| PAY-003 | Authorized reviewer rejects payment with reason | Explicit rejection transition stores reason and preserves consistent registration | Reject operation exists and is transactional | `FULLY_IMPLEMENTED` | `internal/registration/service/registration.go`; registration model | Add append-only history under PAY-006 |
| PAY-004 | Payment review is permission- and tenant-scoped | Only approved reviewer capability within owning Client may verify/reject | Tournament ownership is checked, but broad roles can review; no distinct reviewer assignment/permission enforcement | `SECURITY_GAP` | registration controller; `internal/role/permission/permission.go` | **P0.** Enforce payment-review permission and actor assignment; audit every attempt |
| PAY-005 | Reviewer can discover pending submissions | Provide tenant-scoped, paginated submission list/detail with filters and useful related data | Submit/verify/reject exist, but no payment-submission queue/list/detail API | `API_CONTRACT_GAP` | `internal/transport/http/registration_controller.go`; server route inventory | **P0.** Add review queue/detail contract after tenant permissions and pagination are fixed |
| PAY-006 | Payment settings and history are auditable | Client-managed structured methods/instructions and append-only verification history are persisted | Tournament has free-text instructions; latest verifier/rejection fields overwrite state; no history/settings entity | `PARTIALLY_IMPLEMENTED` | tournament model; registration/payment migration | **P0.** Add structured Client payment settings and append-only review events; depends on audit actor model |
| TEM-001 | Teams support approved organizer/Player creation flows | Create division-bound teams under defined authority and validate tournament state | Player creation and staff membership operations exist; organizer create-team workflow is not explicit | `PARTIALLY_IMPLEMENTED` | `internal/transport/http/team_controller.go`; `internal/team/team.go` | **P0.** Define approved creation actors and route ownership before competition integration |
| TEM-002 | Random/open team assignment follows approved product policy | Implement only if product rule requires it | No random/open assignment exists and PRD leaves exact policy unresolved | `NOT_IMPLEMENTED` | team route/service inventory | **P4.** Product decision required before implementation; do not invent assignment behavior |
| TEM-003 | Team participants can enter competition | Confirmed team registrations/rosters generate valid competition participants | Bracket generator explicitly rejects team divisions | `IMPLEMENTED_INCORRECTLY` | `internal/bracket/service/braket_service.go:112-123` | **P0.** Model team registration/participant snapshot before team bracket generation |
| TEM-004 | One Player has valid team membership per division | Enforce duplicate, roster, and lifecycle rules with database support | AddMember checks registration/duplicate, but create path can bypass prior membership and schema lacks division/player uniqueness | `DATA_MODEL_GAP` | `internal/team/team.go`; `internal/storage/postgres/migration/team.go` | **P0.** Add normalized membership scope and unique constraint; define roster limit/lock rules |
| CHK-001 | Authorized staff can check in registered participants | Explicit check-in action validates registration, actor, window, and state | No route, model, service, or schema exists | `NOT_IMPLEMENTED` | `internal/transport/http/server.go:38-115`; migration inventory | **P0.** Implement after REG-004 status model and permission authority |
| CHK-002 | Check-in state is queryable and auditable | Persist actor/time/state and expose scoped lists | No check-in persistence or API exists | `NOT_IMPLEMENTED` | migration and route inventory | **P0.** Add event/record with actor and timestamp; depends on CHK-001 |
| CHK-003 | Competition uses a locked participant set | Snapshot eligible checked-in participants and prevent silent mutation after generation | No participant-set/lock entity or invariant exists | `DATA_MODEL_GAP` | schema inventory; bracket generator input path | **P0.** Design participant snapshot/version before any additional format engine |
| CMP-001 | Generate approved individual competition formats | Materialize bracket/stages for configured format using locked participants | Individual single elimination is generated; all other approved formats are absent | `PARTIALLY_IMPLEMENTED` | `internal/bracket/service/braket_service.go:107-135` | **P0.** Depend on CHK-003 and aligned format enum before adding algorithms |
| CMP-002 | Seeding supports approved manual/random/order behavior | Validate eligible participant scope, uniqueness, and lifecycle before seeding | Manual/random/registration-order exist; assignment does not verify registration belongs to requested division | `PARTIALLY_IMPLEMENTED` | `internal/seeding/seeding.go:42-54,109-180`; `internal/seeding/model.go` | **P0.** Validate division membership and lock/version; block reseeding after downstream results |
| CMP-003 | Regeneration/reset is safe | Prevent destructive regeneration after matches/results or provide explicit audited reset policy | Reset marks bracket and immediately regenerates without result/dependency checks | `PARTIALLY_IMPLEMENTED` | `internal/bracket/service/braket_service.go:255-273` | **P0.** Add generated-competition version and downstream safety rules after participant locking |
| CMP-004 | Bracket status persists consistently | Domain, DB, request, and response use one enum representation | Bracket status is Go string (`ACTIVE`/`RESET`) while DB column is `SMALLINT` | `IMPLEMENTED_INCORRECTLY` | `internal/bracket/model.go`; `internal/storage/postgres/migration/bracket.go:59-60` | **P0.** Compatibility migration and typed integer enum before production generation |
| CMP-005 | Required formats have real algorithms | Double elimination, round robin, group stage, and group+knockout must generate correct structures | Enum claims double elimination/round robin, but generator is single-elimination only; groups/stages do not exist | `IMPLEMENTED_INCORRECTLY` | tournament format enum; bracket service generation code | **P0.** Implement formats only after common competition/stage model and locked participants |
| SCH-001 | Organizer schedules/reschedules matches | Persist match time and court/venue under owned tournament with lifecycle validation | Scheduling endpoint persists court/time and supports update | `PARTIALLY_IMPLEMENTED` | `internal/match/match.go:189-207`; match controller | **P0.** Validate tournament phase, court ownership, and reschedule consequences |
| SCH-002 | Scheduling prevents resource conflicts | Reject overlapping court/team/participant/official assignments | No court-availability or participant conflict query exists | `IMPLEMENTED_INCORRECTLY` | `internal/match/match.go:189-207`; court/match schema | **P0.** Add transactional conflict detection and supporting indexes; depends on participant model |
| SCH-003 | Courts/venues are manageable in tenant scope | CRUD/status operations remain inside owning tournament/Client | Court routes and ownership checks exist | `FULLY_IMPLEMENTED` | `internal/transport/http/court_controller.go`; `internal/court/court.go` | Preserve scope; add availability semantics under SCH-002 |
| MAT-001 | Matches are materialized and queryable | Generate and list match records linked to competition nodes | Single-elimination generation materializes match shells and detail/list APIs exist | `FULLY_IMPLEMENTED` | `internal/match/match.go`; match controller | Add uniqueness/concurrency constraints under Section 23 |
| MAT-002 | Live/final game scores can be recorded | Persist per-game scores through explicit match operations | Record-game operation and match-game persistence exist | `FULLY_IMPLEMENTED` | `internal/match/model.go:121-125`; `internal/match/match.go` | Validation quality is addressed by MAT-003 |
| MAT-003 | Score entry follows sport-specific rules | Validate score ranges, ties, sets/games, winner calculation, and match state | Only positive game number is validated; negative/tied scores are accepted and winner is caller-declared | `NOT_IMPLEMENTED` | `internal/match/model.go:121-125`; `internal/match/match.go:302-304` | **P0.** Activate scoring configuration and authoritative calculators before result finalization |
| MAT-004 | Forfeit/walkover/disqualification outcomes are representable | Persist approved non-played result types | ResultType includes normal win, walkover, forfeit, disqualification, and bye | `FULLY_IMPLEMENTED` | `internal/match/model.go` | Clarify explicit no-show mapping in Section 26 |
| MAT-005 | Only authorized assigned officials/scorers submit results | Enforce permission and assignment server-side | Any broad tournament staff role can score/finalize; no official assignment model exists | `SECURITY_GAP` | match controller; `internal/role/permission/permission.go`; route inventory | **P0.** Add Scorer/Referee authority and assignment model before production scoring |
| RES-001 | Finalization advances competition correctly | Determine winner and propagate advancement atomically | Declared winner advances through single-elimination winner path; loser/double-elimination paths absent; SQL status predicate is invalid | `PARTIALLY_IMPLEMENTED` | `internal/match/match.go:302-338`; bracket/match code | **P0.** Fix numeric SQL and make outcome calculation authoritative before expanding advancement |
| RES-002 | Standings and tie-breaks are authoritative | Calculate group/round-robin standings and approved tie-breakers | No standings model, service, route, or migration exists | `NOT_IMPLEMENTED` | route/schema inventory | **P0.** Depends on round/group model and sport scoring rules |
| RES-003 | Competition has authoritative champion/final result | Persist/query final placement, champion, and completion evidence | No champion/placement projection or final-result entity exists | `DATA_MODEL_GAP` | schema and route inventory | **P0.** Design competition result projection after advancement model |
| RES-004 | Result correction is explicit and audited | Authorized correction stores reason, before/after state, and recalculates downstream effects | No correction operation/history/recalculation exists | `NOT_IMPLEMENTED` | match/result route inventory | **P0.** Define downstream correction policy, then implement append-only correction events |
| RES-005 | Tournament completion validates competition closure | Complete only when required matches/results/standings are final | Tournament completion checks current status only | `NOT_IMPLEMENTED` | `internal/tournament/tournament.go:178-230`; result inventory | **P0.** Add completion preconditions after RES-002/003 and correction policy |
| RPT-001 | Authorized admins access operational reports | Provide scoped reports for tournaments, registrations, payments, and results | No reporting controller/service/routes exist | `NOT_IMPLEMENTED` | `internal/transport/http/server.go:38-115` | **P1.** Build read models only after source workflows and pagination contracts are correct |
| RPT-002 | Sensitive operations are auditable | Append actor, action, subject, before/after/reason/time for payments, roles, lifecycle, overrides, and corrections | Only latest mutable actor fields exist on some entities; no append-only audit history | `NOT_IMPLEMENTED` | migration inventory; payment/match/tournament services | **P0.** Establish audit event model before adding high-risk transitions and overrides |
| RPT-003 | Auditability does not imply a standalone CMS module | Backend may expose capability without inventing an admin product area | No unnecessary standalone Audit module is exposed | `FULLY_IMPLEMENTED` | route inventory | Preserve separation; expose history contextually unless product approves a module |
| RPT-004 | Large-data exports/report reads are bounded | Reporting must paginate/stream, scope tenants, and avoid N+1 reads | No report APIs exist; existing generic list pagination is miswired | `NOT_IMPLEMENTED` | service search implementations; platform query helper | **P2.** Fix shared list contract first, then design bounded exports/read models |

## 4. Authentication & Authorization

### Implemented surface

`internal/transport/http/server.go:73-115` mounts public authentication and protected API groups. `internal/auth/auth.go:84-140` resolves credentials against the staff User store and then the Player store, creating claims with subject, user type, role, and Client ID. Refresh token rotation/revocation is implemented at `internal/auth/auth.go:181-216` and has focused tests in `internal/auth/auth_test.go`.

### Authority matrix

| Actor | Intended authority | Current enforcement | Finding |
|---|---|---|---|
| Super Admin | Platform Clients, predefined configuration, platform visibility | `IsSuperAdmin` middleware exists | Applied to Sport mutations, but not Client/User platform APIs |
| Client Admin | Own Client and approved staff/tournament operations | `IsClientAdmin` and `RequireClientOwnership` exist | Ownership is strong for tournament-owned routes, absent for User/Client administration |
| Tournament Staff | Assigned operational permissions only | Broad role guard on many nested routes | Granular permission constants are unused; assignment is not modeled |
| Player | Own profile and participation actions | Subject identity is used on Player-facing routes | Generally aligned; global discovery/privacy policy remains incomplete |

### Security and lifecycle findings

- The refresh store at `internal/auth/auth.go:40-68` is package/process state. Restart loses sessions and multiple instances disagree about replay/revocation.
- There is no logout/revoke route and no protected current-identity route.
- `/auth/*` is mounted at the root and again under `/api/v1/auth/*`, creating two equivalent authentication surfaces (`internal/transport/http/server.go:73-80`).
- Permission constants in `internal/role/permission/permission.go` do not participate in controller or service decisions.
- Authorization and tenant checks are concentrated in HTTP controllers. Direct service callers do not consistently carry an actor/scope invariant. This materially increases the chance that a new transport or job bypasses policy.

### Required remediation

First define typed roles and approved permissions. Then enforce platform operations, tenant scope, and assignment rules at both route and use-case boundaries. After that, introduce durable refresh-session authority, `/me`, and logout. Do not extend the existing string role model.

## 5. Client / Tenant Model

Clients have a sound base entity with status, unique slug, and timestamps. The service contains create/search/get and activate/suspend behavior. The HTTP layer only exposes create/search/get, and `internal/transport/http/client_controller.go:13-18` applies no Super Admin guard.

### Ownership coverage by operation

| Resource/operation | Scope evidence | Assessment |
|---|---|---|
| Client create/list/detail | No platform-role guard | **P0 cross-tenant/platform authority gap** |
| User create/list/detail | Request query/body can influence Client; actor scope not forced | **P0 cross-tenant/privilege gap** |
| Tournament create | Non-Super Admin Client ID overwritten from identity | Good |
| Tournament list/detail/update/status | Owned tournament helper plus `RequireClientOwnership` | Good |
| Divisions/courts/seeding/brackets/matches | Nested tournament ownership resolution | Good at HTTP boundary |
| Registrations/payments/teams | Nested tournament scope plus broad role gates | Tenant scope mostly good; permissions too broad |
| Player global identity | No tenant owner by design | Correct; access policy must protect global PII |

Client suspension does not consistently block downstream operations. A suspended Client could retain valid staff tokens and reach routes that only compare ownership. Add active-Client checking in a shared policy after lifecycle routes and semantics are completed.

## 6. Users & Role Authority

The User domain correctly remains separate from Player. The platform authority model is not safe:

- `internal/transport/http/user_controller.go:16-21` mounts create/search/detail without role or tenant middleware.
- The intended authority check is commented out at `internal/transport/http/user_controller.go:24-42`; version control, not commented code, should retain old logic under the `.cursor` code-style rule.
- `internal/user/model.go:91-124` accepts any non-empty role string. A Client Admin can therefore request `super_admin` unless another unstated layer rejects it; no such service check was found.
- Search passes request-supplied Client scope rather than forcing the actor Client.
- User status commands/errors exist (`internal/user/user.go:9-18`, `internal/user/model.go:175-191`), but service/route behavior is absent.
- Create returns only a message (`internal/transport/http/user_controller.go:49-55`), unlike most other create operations.

Required authority order:

1. Replace arbitrary role strings with stable role IDs or typed integer enums matching approved predefined roles.
2. Encode who may assign each role; a Client Admin must never assign Super Admin or another disallowed authority.
3. Derive Client scope from the verified identity for non-platform actors.
4. Add explicit User status transitions and append-only role/status history.
5. Return a safe canonical User projection after successful creation.

## 7. Players / Members

The backend correctly models Product Player separately from admin/staff User. A Player is global and registrations are scoped through tournament/division relationships. This supports cross-Client participation without duplicating identity.

Gaps:

- Admin Player search is available to any authenticated CMS User; allowed visibility and PII projection are not defined or enforced.
- Requested pagination is not wired into the query helper, so global Player growth will break the advertised list contract.
- Player update writes `updated_by` (`internal/player/model.go:39,63`) but the migration never creates it (`internal/storage/postgres/migration/player.go:10-128`). This is a runtime schema defect.
- Player `date_of_birth` is stored as `TIMESTAMPTZ`; eligibility is date-based and should use a date-only semantic unless the PRD later defines a timezone-sensitive birth instant.
- The exact Player status set and authority are unresolved in the approved documents. No status workflow should be invented.

## 8. Sports

Sports are implemented as mutable platform records with create, update, status change, list, and detail endpoints. Mutation endpoints are Super Admin guarded, but this exceeds the approved model of predefined platform Sports.

Contract defects:

- `internal/sport/model.go:87-90` filters on legacy `is_active`, while entity/schema use numeric `status`.
- `internal/transport/http/sport_controller.go:17` and `:22` register `GET /sports/` twice in overlapping route groups.
- The enum itself is numeric and validates Active/Inactive, but API consumers can observe incorrect filtering because the repository query still targets the obsolete column.

Recommended resolution: decide whether Sports are code-seeded or migration-seeded predefined configuration. Keep a stable read API. Remove or constrain arbitrary CRUD only after confirming operational bootstrap requirements; do not delete it as part of the audit.

## 9. Tournaments

### Current state model

| Current state | Exposed action | Next state | Current precondition | Missing PRD enforcement |
|---|---|---|---|---|
| Draft | Publish | Published | Current state only | Required setup completeness, division/config readiness |
| Published | Open registration | Registration Open | Current state only | Registration window/settings readiness |
| Registration Open | Close registration | Registration Closed | Current state only | Date/capacity policy, pending review policy |
| Registration Closed | Start | Ongoing | Current state only | Check-in complete, participant set locked, competition generated/scheduled |
| Ongoing | Complete | Completed | Current state only | All required results final, disputes resolved, champion/standings materialized |
| Nonterminal | Cancel | Cancelled | Limited state check | Reason, audit event, downstream consistency |
| Any approved state | Archive | Archived | Not implemented | Retention/read-only policy |

`internal/tournament/model.go:28-38` lacks Competition Preparation and Archived states. `internal/tournament/tournament.go:178-230` implements explicit operations, which is preferable to a generic status patch, but only current-state checks are enforced. Cancel also omits an `UpdatedAt` value in its update payload.

Core ownership behavior is good: non-Super Admin creation derives Client ID from identity and tournament-owned reads/mutations use ownership checks. The format contract is incomplete: enum values include Single Elimination, Double Elimination, Round Robin, and Swiss, while the PRD requires Single Elimination, Double Elimination, Round Robin, Group Stage, and Group + Knockout. Only individual Single Elimination has a generator.

## 10. Divisions & Eligibility

Division CRUD is nested under an owned tournament and supports participant type, capacity, gender, age, and skill-related configuration. However:

- ParticipantType and Status are string enums (`internal/division/model.go:25-41`) rather than the approved integer enum contract.
- Registration never evaluates configured eligibility rules.
- Update is blocked only when the tournament is Completed or Cancelled. Registration/competition preparation is not protected.
- Activate/deactivate does not call the parent editability guard (`internal/division/division.go:157-176`).
- There is no archive/delete policy and no behavior for divisions that already have registrations.

The eligibility evaluator must be deterministic and date-aware. Define age calculation date, inclusive bounds, missing profile behavior, and override authority before implementation. Eligibility must execute server-side during registration and be rechecked or frozen when participants lock.

## 11. Registrations

The individual registration path has meaningful implementation: it requires Registration Open, loads an active division, checks duplicates, locks capacity-related rows, counts registrations, and inserts within a transaction (`internal/registration/service/registration.go:82-171`). Tenant ownership is enforced by nested controller resolution.

Critical defects:

- SQL compares numeric status columns with text literals: `status <> 'CANCELLED'` at `internal/registration/service/registration.go:39-47` and `:129`, and `status = 'CONFIRMED'` at `:64-72`. PostgreSQL will reject or mis-handle these predicates depending on implicit casts.
- Eligibility is not evaluated.
- The explicit registration start/end dates are not comprehensively enforced in the service.
- Statuses only include Pending, Confirmed, and Cancelled (`internal/registration/model.go:31-46`). Waitlisted and Checked In are absent.
- Unique `(division_id, player_id)` in `internal/storage/postgres/migration/registration.go:33-38` prevents a cancelled Player from registering again, even though service logic attempts to reason about active registration.
- No waitlist ordering, promotion, late override, or check-in exists.

### Registration transition assessment

| Current | Action | Expected next | Current behavior |
|---|---|---|---|
| None | Register | Pending/Confirmed/Waitlisted by policy | Pending path exists; eligibility/waitlist missing |
| Pending | Payment verified / organizer confirms | Confirmed | Payment verification confirms automatically |
| Pending/Confirmed | Cancel | Cancelled | Cancellation exists, but re-registration schema conflicts |
| Waitlisted | Promote | Pending/Confirmed | Not implemented |
| Confirmed | Check in | Checked In | Not implemented |
| Checked In | Lock participant | Competition participant snapshot | Not implemented |

## 12. Payments

The backend supports manual payment submission, verification, and rejection. Review operations are transactional and record a verifier/rejection reason. This is a good base for the PRD’s manual-review model.

Missing or unsafe areas:

- No Client-managed Payment Settings entity, structured method list, account/QR metadata, or policy version. Tournament free-text instructions are the only organizer configuration.
- No paginated pending-payment queue or payment detail endpoint exists; reviewers must already know nested identifiers.
- Broad roles can verify/reject. Permission constants exist but are not enforced.
- Review automatically confirms the registration without a separately documented policy switch.
- Mutable latest fields are not an append-only verification history.
- Money is represented as Go `float64` while PostgreSQL uses numeric. The `.cursor` production standard requires integer minor units plus currency to avoid rounding ambiguity.
- `proof_reference` exists in API/schema without a complete approved upload/storage flow. It must remain an optional text reference or be removed after product decision; do not imply secure file evidence handling.

### Payment transition assessment

| Current | Action | Next | Actor/precondition | Enforcement |
|---|---|---|---|---|
| None | Submit details | Pending | Registered Player; payment required | Partially enforced |
| Pending | Verify | Verified | Authorized reviewer in owning Client | Transaction exists; permission too broad |
| Pending | Reject | Rejected | Authorized reviewer; reason | Transaction exists; permission too broad |
| Rejected | Resubmit | Pending | Approved policy | Exact behavior not established |
| Verified | Reject/reverify | Invalid unless approved correction | No append-only correction policy |

## 13. Teams

Team create, membership add/remove, disband, and nested ownership routes exist. AddMember checks active registration and duplicate team participation and uses a division lock. The product flow is still incomplete:

- Team creation is primarily Player-facing; the approved organizer creation/assignment authority is not explicit.
- CreateTeam does not establish that the creator is not already a member of another team in the division.
- The database lacks a constraint that guarantees one active team membership per division/Player.
- Roster capacity, captain/manager authority, lock timing, and edits after competition generation are undefined or unenforced.
- Random/open assignment is absent and is product-ambiguous rather than automatically required.
- `internal/bracket/service/braket_service.go:112-123` explicitly rejects team divisions, so teams cannot complete the tournament flow.

Route ownership is centralized in the team controller, but the team prefix is mounted separately from several tournament-owned route groups. Future work should preserve one authoritative route owner and the existing controller tests should be expanded to tenant and participant invariants.

## 14. Check-In & Participant Locking

No check-in or participant-lock implementation exists in routes, services, models, or migrations. This is a foundational P0 gap because generation currently reads mutable registrations/seeding directly.

Minimum backend model needed after product clarification:

- explicit check-in event/state with registration, actor, timestamp, and optional approved override reason;
- tenant/tournament scope and permission checks;
- participant-set snapshot/version tied to division/competition;
- immutable participant identity and seed at generation time;
- explicit unlock/regenerate policy with audit and downstream-result safety;
- counts that reconcile checked-in, locked, seeded, and generated participants.

Do not build more competition formats before this foundation; otherwise a registration/team edit can silently invalidate generated brackets.

## 15. Competition Engine

### Format coverage

| Format | Enum/config | Participant support | Generation algorithm | Result progression | Assessment |
|---|---|---|---|---|---|
| Single elimination | Present | Individual only | Present | Winner path only | Partial |
| Double elimination | Enum present | None | Absent | Loser bracket absent | Incorrectly claimed |
| Round robin | Enum present | None | Absent | Standings absent | Incorrectly claimed |
| Group stage | Enum absent | None | Absent | Standings absent | Missing |
| Group + knockout | Enum absent | None | Absent | Absent | Missing |
| Swiss | Enum present | None | Absent | Absent | Out-of-scope/future enum only |

Seeding supports Manual, Random, and Registration Order. Manual assignment does not verify that the registration belongs to the requested division (`internal/seeding/seeding.go:42-54`). Generation replaces existing seeds without protecting generated matches/results (`:109-180`).

Bracket generation creates a single-elimination tree and byes, but:

- teams are rejected;
- participant set is not locked;
- bracket status is a Go string while database status is `SMALLINT`;
- no database uniqueness prevents multiple active brackets or concurrent duplicate generation;
- reset marks the prior bracket and immediately regenerates without checking results (`internal/bracket/service/braket_service.go:255-273`);
- no stages, groups, rounds, loser paths, standings, or format-specific configuration exists.

## 16. Scheduling

Courts can be created, listed, updated, and status-managed within an owned tournament. Matches can receive a court and scheduled time (`internal/match/match.go:189-207`). This persists a schedule but does not make it valid.

Missing enforcement:

- selected court belongs to the tournament and is active;
- no overlapping matches on a court;
- no Player/team overlap across simultaneous matches;
- venue operating constraints;
- match/tournament lifecycle eligibility;
- official/referee assignment and overlap;
- reschedule history and participant notification side effects.

Conflict prevention should use transaction-safe queries and supporting indexes. Application-only prechecks are insufficient under concurrent scheduling.

## 17. Matches & Scoring

Match records and per-game scores exist. The result model includes Normal Win, Walkover, Forfeit, Disqualification, and Bye. However, scoring is not authoritative:

- validation only ensures `game_number >= 1` (`internal/match/model.go:121-125`);
- negative scores and invalid ties are accepted;
- sport-specific set/game rules are absent;
- record-game is permitted in states beyond a clearly approved live-scoring window;
- finalization accepts a caller-declared winner rather than deriving it (`internal/match/match.go:302-304`);
- any broad tournament staff role can record/finalize; no Scorer/Referee assignment exists;
- the atomic finalize predicate compares numeric status with `'COMPLETED'` and `'CANCELLED'` strings (`internal/match/match.go:333-338`).

`SportScoringConfigMigration` exists but is not registered by `internal/storage/postgres/postgres.go:38-57`, and no active scoring service consumes it. Before enabling production result entry, define per-sport rule configuration, validate games, compute outcomes server-side, and enforce assigned scorer/official permissions.

## 18. Results & Advancement

Single-elimination finalization can place a declared winner into the next bracket node. It does not implement loser paths, group ranking, qualification, tie-breaks, final placement, or a champion read model.

### Result-authority gaps

| Capability | Current behavior | Required behavior |
|---|---|---|
| Winner calculation | Caller declares winner | Backend derives from validated sport rules or approved non-played result |
| Winner advancement | Single winner path | Format-specific winner/loser/qualification paths |
| Standings | None | Authoritative group/round-robin table and tie-breaks |
| Champion/final placement | None | Persisted/queryable competition result projection |
| Correction | None | Dedicated action with reason, before/after, authorization, downstream recomputation |
| Tournament completion | Current-state check only | Verify every required competition result is final and disputes/corrections resolved |

Result correction is particularly dependent on product policy. If downstream matches have started, the backend needs an approved rule for blocking, voiding, or recalculating them. The audit records this as an unresolved policy in Section 26 rather than inventing behavior.

## 19. Reporting & Auditability

There are no report, standings, audit-history, or export routes in `internal/transport/http/server.go:38-115`. Existing entity timestamps and latest reviewer/finalizer fields are operational metadata, not an audit log.

Sensitive actions requiring append-only history include:

- staff role assignment and status change;
- Client lifecycle changes;
- tournament lifecycle transitions and overrides;
- eligibility/check-in overrides;
- payment verify/reject/resubmit/correction;
- participant lock/unlock/regenerate;
- seeding changes;
- score finalization and result correction.

The PRD requires auditability, not necessarily a standalone CMS Audit module. A general audit event store or domain-specific histories may satisfy the backend need. Reporting should be built from stable, scoped read models after underlying workflows are correct, with pagination/streaming rather than fetch-all behavior.

## 20. Enum / Status Model

### Cross-layer enum matrix

| Domain enum | Go representation | DB representation | API representation | Assessment |
|---|---|---|---|---|
| Client status | Typed integer | `SMALLINT` | Number | Consistent base; lifecycle incomplete |
| User status | Typed integer | Numeric | Number | Base exists; no active transition API |
| Role | String constants | String/text association | String | Not an approved stable integer/predefined authority contract |
| Player status | Not defined | None | None | Product policy unresolved |
| Sport status | Typed integer | Numeric | Number | Search still uses legacy `is_active` |
| Tournament status | Typed integer | Numeric | Number | Missing preparation/archive states and preconditions |
| Tournament format | Typed integer | Numeric | Number | Missing group formats; unapproved Swiss; engine mismatch |
| Division status | String | Text/string | String | Violates integer enum rule |
| Division participant type | String | Text/string | String | Violates integer enum rule |
| Registration status | Typed integer | `SMALLINT` | Number | SQL uses string literals; states incomplete |
| Payment status | Typed integer | Numeric | Number | Base transitions exist; history/settings incomplete |
| Team status | Typed integer | Numeric | Number | Lifecycle and constraints incomplete |
| Seeding method/status | Typed integer | Numeric | Number | Scope/regeneration rules incomplete |
| Bracket status | String | `SMALLINT` | String | Direct model/schema mismatch |
| Match status | Typed integer | Numeric | Number | Finalize SQL uses string literals |
| Match result type | Typed integer | Numeric | Number | Good representation; no-show mapping ambiguous |

### Required enum migration discipline

Do not patch individual literals in isolation. Inventory stored values, define canonical typed constants and `Validate()` methods, introduce compatibility-safe migrations, update bind/response contracts, and add round-trip database/API tests. Correct Registration/Match/Bracket runtime mismatches first. Then align Division and role contracts. Status transitions must remain domain actions, not general update fields.

## 21. API Contract Gaps

### Route and mutation contract findings

| Area | Current contract | Gap |
|---|---|---|
| Authentication | Login/refresh/register mounted at root and `/api/v1` | Duplicate API ownership; no `/me` or logout |
| User create | Success message only | Caller must perform a follow-up read and may not know canonical projection |
| Client lifecycle | Service has activate/suspend; HTTP does not | Backend capability is unreachable and unauthorized surface remains incomplete |
| Payment review | Submit/verify/reject only | No queue/list/detail for reviewers |
| Bracket generation | Returns bracket header | Nodes/round summary require follow-up reads |
| Match generation | Returns match shells | Participants/details are incomplete for immediate UI integration |
| Status transitions | Mostly empty success; updates often return entities | Mutation responses are inconsistent |
| Reports/standings | No contract | Required read capabilities absent |

### List/search contract

Services populate sort fields but do not map requested `meta.Meta.Page`, `Limit`, or cursor into `searchHelper.Params.Pagination`. The pinned helper at `github.com/imsab23/platform-be/.../helper/query/query.go:204-249` only applies supplied pagination and otherwise defaults to 20. Because `WithCount` is not enabled, total counts remain zero. Consequences:

- request page/limit parameters are misleading;
- next-page/has-more semantics cannot be trusted consistently;
- the CMS cannot build accurate pagination;
- high-volume resources appear capped rather than truly paginated;
- `Total` metadata is not meaningful.

Fix one shared service pattern and add contract tests for page, limit cap, stable sort, cursor/page semantics, total/has-more, empty result, and tenant scope. Do not expose both offset and cursor semantics without a documented precedence rule.

### Search/filter/sort and display data

- Search wildcard escaping is disabled by default in the platform helper; `%` and `_` act as wildcards. This is parameterized and not SQL injection, but it violates literal-search expectations unless documented.
- Sport status filter references a removed/legacy boolean.
- Payment has no list filters at all.
- Many list responses primarily expose foreign keys. Add minimal display projections only where they prevent repeated Client/Sport/Tournament/Player lookups; do not embed full graphs.
- Dynamic sorting must remain allow-listed. Existing helpers provide the mechanism, but each service mapping must be reviewed.

### Large-data readiness

Players, Users, Tournaments, Registrations, Payments, Teams, Matches, and Results require bounded database reads. Current migrations contain useful primary/unique indexes but sparse explicit indexes for tenant foreign keys, statuses, dates, and common filter/sort combinations. No report/export path exists. N+1 risk is highest where the frontend must resolve Client, Tournament, Division, and Player labels separately.

## 22. Security / Tenant Gaps

| Priority | Gap | Exploit/impact | Evidence | Required control |
|---|---|---|---|---|
| P0 | Client APIs lack platform authority | Any CMS User may enumerate/create platform tenants | `client_controller.go:13-18` | Super Admin route and service policy |
| P0 | User APIs lack actor/tenant checks | Cross-tenant staff enumeration/creation | `user_controller.go:16-42` | Scope from verified identity; tenant predicates in queries |
| P0 | Arbitrary role string assignment | Client Admin may request higher authority | `user/model.go:91-124` | Predefined roles plus assignment matrix |
| P0 | Payment review uses broad role | Unauthorized staff can approve/reject money evidence | registration controller | Permission and assignment checks; audit event |
| P0 | Scoring uses broad role | Unauthorized staff can determine outcomes | match controller | Assigned official/scorer model and capability |
| P0 | Session revocation is process-local | Restart/multi-instance replay inconsistency | `auth/auth.go:40-68` | Durable/distributed session store |
| P0 | Service invariants rely on HTTP boundary | Alternate callers can bypass ownership policy | controller/service design | Actor/scope-bearing use-case boundary |
| P1 | Global Player list exposure | Excess PII visibility across Clients | `player_controller.go:30-80` | Approved discovery scope and minimized projection |
| P1 | Client suspension not globally enforced | Suspended tenant may continue operations | client/middleware inventory | Active-Client policy middleware/service guard |

Queries are parameterized, so no direct SQL-injection finding was observed in the audited paths. IDOR/BOLA risk remains material because missing tenant predicates and authority checks allow valid identifiers to cross scope. Frontend hiding cannot mitigate these gaps.

No credentials, secrets, or tokens were introduced by this audit. The repository’s `.env` was not copied into the report.

## 23. Data Model Gaps

### Blocking schema defects

1. Player update writes `updated_by` absent from Player migrations.
2. Registration and Match SQL use string literals against numeric status columns.
3. Bracket status is string in Go and `SMALLINT` in PostgreSQL.
4. Unique registration `(division_id, player_id)` cannot express “one active registration” if cancelled re-registration is allowed.
5. Team membership lacks a constraint for one Player per division/team scope.
6. Match generation lacks uniqueness on bracket node, permitting duplicate materialization under concurrency.
7. Match games lack a uniqueness constraint on match/game number.
8. No check-in, participant snapshot/lock, standings, champion/result projection, result correction, audit-history, payment-settings, promotion/discount, or official-assignment schema exists.

### Integrity and audit constraints

- Actor columns such as `approved_by`, `verified_by`, `seeded_by`, `generated_by`, and `finalized_by` are not consistently protected by foreign keys.
- Bracket-node participant references are weakly modeled for individual/team polymorphism.
- Enum columns lack database check constraints, allowing invalid numeric states.
- Score, amount, capacity, and date relationships lack comprehensive checks.
- Player date of birth uses timestamp semantics rather than date-only semantics.
- Money uses application `float64`; migrate to integer minor units plus currency or a rigorously defined decimal contract.

### Migration execution risk

`internal/storage/postgres/postgres.go:38-57` registers multiple migrations with the same `Version: 1`. The pinned platform runner sorts only by version using `sort.Slice` (`.../migration/runner.go:35-54`). Equal-version order is not stable, so foreign-key-dependent table creation can be nondeterministic on a fresh database. Migrations are keyed by version/name, so they may all execute, but order remains unsafe. Assign deterministic unique versions in a compatibility-safe migration plan.

`SportScoringConfigMigration` is defined but not registered. Password-reset and proof-reference schema exist without complete routes/use cases. These are inventory findings, not proof of a working capability.

### Index review

Add indexes only from observed query shapes. Immediate candidates are tenant/owner foreign keys plus status/date order for User, Tournament, Registration, Payment, Team, and Match lists; court/time overlap lookup; division/participant lookup; and audit event subject/time. Pair every proposed index with an `EXPLAIN` or representative query test during remediation.

## 24. Over-Implemented / Out-of-Scope Features

| Implementation | Classification | Evidence | Recommendation |
|---|---|---|---|
| Arbitrary Sport create/update/status API | `OUT_OF_SCOPE` | sport controller/service | Preserve read access; resolve seeding/config ownership before removal |
| Swiss tournament format enum | `FUTURE_READY_BUT_HARMLESS` but misleading | `tournament/model.go:40-55` | Remove from public accepted values unless roadmap approves and implements it |
| Payment proof-reference field | `FUTURE_READY_BUT_HARMLESS` | registration/payment model and migration | Keep as plain reference only or remove; do not imply file-upload support |
| Password-reset table without flow | `DEAD_CODE` / orphaned schema | Player migration inventory | Implement only if approved auth flow exists; otherwise retire in later migration |
| Broad permission constant catalog with no enforcement | `FUTURE_READY_BUT_MISLEADING` | `role/permission/permission.go` | Make it authoritative or reduce it after role design review |
| Sport scoring config migration not registered/used | `DEAD_CODE` / orphaned | storage migration and registration list | Wire only as part of approved scoring design, not as isolated schema |

No platform database/router/logging/transaction infrastructure was classified as over-implementation merely because the PRD does not mention it.

## 25. Duplicate / Obsolete Implementation

### Duplicate and orphaned behavior

| Finding | Classification | Evidence | Risk |
|---|---|---|---|
| Root `/auth/*` and `/api/v1/auth/*` | `DUPLICATE_IMPLEMENTATION` / `UNUSED_API` | `internal/transport/http/server.go:73-80` | Contract drift and inconsistent middleware/versioning |
| `GET /sports/` registered twice | `DUPLICATE_IMPLEMENTATION` | `sport_controller.go:17,22` | Ambiguous route ownership/middleware behavior |
| User status command/errors but no reachable service/route | `DEAD_CODE` / orphaned | `user/user.go:9-18`; `user/model.go:175-191` | Misleading capability |
| Commented-out User authority logic | `OBSOLETE_IMPLEMENTATION` | `user_controller.go:24-42` | Security intent exists only as comment |
| Deleted tracked `pkg/` tree plus untracked `internal/` tree | Audited rearchitecture state | working-tree snapshot | Risk of incomplete migration and difficult review; deleted tree is not active runtime code |

### Folder-structure and code-style alignment

The `.cursor` skills specify domain-first `internal/<domain>/model.go`, optional errors/constants file, `internal/<domain>/service/<domain>_service.go`, service-owned SQL, thin `transport/http/handler`, deterministic migrations, no repository layer, assign-then-check error style, descriptive names, explicit context/timeouts, integer money, and strong tests.

The current rearchitecture follows domain grouping and does not introduce a repository layer, but materially diverges in places:

- several services/SQL files remain at domain root (`internal/tournament/tournament.go`, `internal/match/match.go`, `internal/team/team.go`, `internal/division/division.go`, `internal/court/court.go`, `internal/seeding/seeding.go`) while other domains use `service/`, producing two internal conventions;
- transport is under `internal/transport/http` and names files/controllers rather than the documented root `transport/http/handler` layout;
- migrations are Go types under `internal/storage/postgres/migration` rather than deterministic numbered migration files;
- several controllers contain ownership/authority decisions that should also be explicit in use-case/service policy;
- package-level mutable refresh state conflicts with production rules;
- money uses `float64`;
- commented-out behavior and vague legacy naming (`braket_service.go`) conflict with the code standards;
- tests cover auth and team routing only, not database contracts, tenant isolation, lifecycle transitions, or concurrency.

These structural differences should be resolved deliberately after correctness/security remediation. Do not perform a mass folder rewrite alongside business fixes. First choose the `.cursor` structure as the target, then migrate one domain at a time with behavioral tests.

## 26. Ambiguous Requirements

No substantive PRD-versus-business-flow conflict was found. Therefore there is no `DOCUMENTATION_CONFLICT` row. The following subrules require product decisions and must not be guessed during remediation:

| Area | Ambiguity | Safe interim rule |
|---|---|---|
| Player status | Exact states and which actors can suspend/reactivate a global Player | Preserve current identity; do not add status mutation until approved |
| Late registration | Who may override close date/capacity/eligibility and how it is audited | Reject normal late registration; do not add override |
| Waitlist | Ordering, payment timing, expiry, and promotion rules | Do not implement promotion heuristics |
| Check-in | Window, late check-in, reversal, and override authority | Model explicit policy before route work |
| Participant lock | Reopen, replacement, withdrawal, and regeneration rules | Treat generated snapshot as immutable until policy exists |
| Team assignment | Player-created, organizer-created, random/open assignment, and roster limits | Preserve existing operations; do not add random assignment |
| Officials | Roles, assignment source, substitutions, and conflict rules | Do not infer from broad Tournament Staff |
| No-show | Whether it is a distinct result type or represented by walkover/forfeit | Preserve existing types pending clarification |
| Result correction | Behavior after downstream match starts or tournament completion | Do not permit correction without explicit policy |
| Reports | Exact report catalog, retention, export limits, and financial currency semantics | Start only from approved report specifications |
| Audit | Retention, visibility, redaction, and data-subject requirements | Capture minimum immutable operational facts once policy is approved |
| Payment gateways | Whether/when automated gateways, webhooks, refunds, and reconciliation enter scope | Keep manual verification only |
| Promotions/discounts | Approved discount types, stacking, eligibility, and lifecycle | No implementation until PRD amendment |

If implementation work reaches any of these policy boundaries, mark the affected finding `AMBIGUOUS_REQUIREMENT`, record the decision needed, and continue with unrelated confirmed gaps.

## 27. Recommended Remediation Order

### Forty-domain disposition

| # | Domain | Current disposition |
|---:|---|---|
| 1 | Authentication | Partial; session/current-user/logout gaps |
| 2 | Authorization / RBAC | P0 security gap |
| 3 | Clients | Partial; P0 authority/scope gap |
| 4 | Admin / Staff Users | P0 security and contract gaps |
| 5 | Roles | Partial predefined model; assignment unsafe |
| 6 | Players / Members | Global identity aligned; schema/list gaps |
| 7 | Sports | Predefined-read need; mutable CRUD is OOS |
| 8 | Tournament | Partial lifecycle/configuration |
| 9 | Divisions / Categories | CRUD present; enum/lifecycle gaps |
| 10 | Eligibility | Config partly present; enforcement absent |
| 11 | Registration | Partial; status SQL, window, waitlist gaps |
| 12 | Payments | Manual review partial |
| 13 | Payment Settings | Missing structured model |
| 14 | Promotions / Discounts | Not approved in current PRD; ambiguous/out of scope |
| 15 | Teams | Partial operations; competition unsupported |
| 16 | Team Membership | Partial; constraint/lifecycle gaps |
| 17 | Check-In | Missing |
| 18 | Participant Locking | Missing data model |
| 19 | Competition Configuration | Partial format/seeding config |
| 20 | Competition Generation | Individual single elimination only |
| 21 | Seeding | Partial; scope/reset gaps |
| 22 | Randomization | Random seeding present; deterministic/audit policy incomplete |
| 23 | Stages | Missing |
| 24 | Groups | Missing |
| 25 | Rounds | Implicit bracket rounds only; no general model |
| 26 | Brackets | Partial with status mismatch |
| 27 | Scheduling | Partial; conflict prevention absent |
| 28 | Courts / Venues | Basic CRUD/scoping implemented |
| 29 | Officials | Missing |
| 30 | Matches | Basic materialization/query implemented |
| 31 | Scoring | Storage present; rule authority missing |
| 32 | Result Confirmation | Partial caller-declared finalization |
| 33 | Result Correction | Missing/ambiguous downstream policy |
| 34 | Forfeit / No-show | Forfeit/walkover present; no-show mapping ambiguous |
| 35 | Standings | Missing |
| 36 | Advancement | Winner-only single-elimination partial |
| 37 | Tournament Completion | State change exists; closure validation missing |
| 38 | Reporting | Missing |
| 39 | Audit / History | Missing append-only history |
| 40 | Platform Configuration | Roles/Sports partial and inconsistent |

### Dependency-aware phases

**Phase 0A — runtime and migration safety**

1. Repair integer/string status SQL for Registration and Match.
2. Align Bracket status across Go/DB/API.
3. Repair Player `updated_by` schema mismatch.
4. Make migration ordering deterministic.
5. Add database integration tests for these paths.

**Phase 0B — authority and tenant foundation**

1. Establish typed predefined roles, hierarchy, and permission evaluation.
2. Protect Client APIs with Super Admin authority.
3. Scope User create/list/detail/status to actor Client and prevent privilege escalation.
4. Enforce active-Client state consistently.
5. Add append-only audit foundation for role/status and sensitive transitions.

**Phase 0C — authentication and API foundation**

1. Move refresh sessions/revocation to durable shared storage.
2. Add current-identity and logout contracts.
3. Normalize created-resource responses and machine-readable domain errors.
4. Fix shared pagination/filter/sort metadata and tests.

**Phase 0D — tournament participation foundation**

1. Align tournament/division enums and approved formats.
2. Complete lifecycle prerequisites and explicit archive/preparation semantics.
3. Finalize eligibility and registration status/window/waitlist rules.
4. Add structured payment settings, review queue, reviewer permission, and history.
5. Enforce team membership constraints and team-registration semantics.
6. Implement check-in and locked participant snapshots.

**Phase 0E — competition and results**

1. Make seeding division-safe and immutable after lock/generation.
2. Add team participant generation.
3. Implement approved format models/algorithms in order: single-elimination hardening, round robin, groups, group+knockout, double elimination.
4. Add schedule conflict detection and officials.
5. Add sport-specific scoring and server-derived outcomes.
6. Add standings, loser/winner advancement, champion projection, result correction, and completion validation.

**Phase 1/2 — reports and operational scale**

1. Add scoped operational read models and bounded exports.
2. Add indexes proven by actual query plans.
3. Complete audit retrieval/redaction/retention policy.
4. Remove reviewed duplicate, obsolete, and out-of-scope surfaces only after dependency analysis.

### P0/P1 dependency constraints

- Do not implement more competition formats before participant locking.
- Do not implement standings before round/group and scoring rules.
- Do not permit result correction before downstream-impact policy and audit history.
- Do not add payment review UI/API breadth before reviewer permission and tenant scope.
- Do not remove Sport CRUD until bootstrap/integration dependencies are known.
- Do not restructure folders during security/data fixes; bracket behavior with tests, then migrate structure separately.

## 28. Baseline Build/Test Status

Commands were rerun against the final audited working tree on 2026-10-10.

| Command | Result | Evidence/notes |
|---|---|---|
| `go test ./...` | **PASS** | Auth and HTTP controller tests passed; most packages report `[no test files]` |
| `go vet ./...` | **PASS** | No output, exit code 0 |
| `go build ./...` | **ENVIRONMENT/REPOSITORY-METADATA FAILURE** | Go VCS stamping failed because `git status` returns `fatal: this operation must be run in a work tree` despite `rev-parse` identifying the work tree |
| `go build -buildvcs=false ./...` | **PASS** | Confirms source compiles when VCS stamping is disabled |

The passing suite does not validate PostgreSQL runtime behavior for the identified enum/column mismatches, tenant isolation, lifecycle invalid transitions, capacity concurrency, payment review authority, competition generation safety, or result advancement. Only auth and team-controller tests are materially present. The repository should add integration coverage before treating compile success as readiness.

### Final audit totals

- Fully implemented: **22**
- Partially implemented: **24**
- Not implemented: **12**
- Implemented incorrectly: **5**
- Security gaps: **7**
- Data-model gaps: **4**
- API-contract gaps: **3**
- Out-of-scope implementation: **1**
- Ambiguous primary requirement rows: **0**
- Documentation conflicts: **0**
- Total PRD requirements: **78**

