# Domain Package

Read when writing or reviewing `model.go`, `<domain>.go`, or a service file, and when deciding whether to split a file.

## Contents
1. What each file holds
2. `model.go`
3. `<domain>.go`
4. The service file
5. Queries
6. Transactions
7. Cross-domain calls
8. Search and pagination
9. Splitting rules
10. Testing placement

## 1. What each file holds

```text
internal/client/
├── client.go                  errors and constants of the domain
├── model.go                   entities, commands, queries, results, enums, Validate()
└── service/
    └── client_service.go      package clientsvc: rules + SQL + transactions
```

## 2. `model.go`

The data contracts of the domain. It may contain:

- entity structs (with `json` and `db` tags in `snake_case`);
- commands (input that changes state) and queries (input that reads);
- update structs using pointer fields for partial updates;
- search result structs using the platform's pagination/meta type;
- enums and their `Validate()` methods;
- `Validate()` methods on commands and queries.

It does **not** contain SQL, I/O, calls to other domains, or business workflows. Keep related models in this one file until it becomes hard to maintain. Style details (enums, validation, partial updates) are in `go-code-standard/references/idioms-and-design.md`.

## 3. `<domain>.go`

Only the domain's named errors and constants, using the project's error system:

```go
package client

var (
	ErrClientNotFound = apperror.New("CLI0000", "Client not found.")
)
```

Create the file only when it has content; do not leave it empty. Do not put types or logic here.

## 4. The service file

Order inside `client_service.go`:

1. package clause and imports;
2. SQL constants;
3. the service struct;
4. the constructor `NewClientService`;
5. public methods, grouped by operation: reads (`GetByID`, `Search`), creates, updates, state changes (`Suspend`, `Activate`), deletes;
6. unexported helpers at the bottom.

The service struct holds its dependencies: the platform database, the platform logger if it needs one, and small interfaces for other domains or external systems.

```go
type ClientService struct {
	database db.DB
}

func NewClientService(database db.DB) *ClientService {
	return &ClientService{database: database}
}
```

Responsibilities:

- enforce business rules (state checks, ownership, uniqueness rules not covered by constraints);
- validate commands at the start of each public method;
- run the domain's SQL through the platform database type;
- translate storage errors into domain errors (`sql.ErrNoRows` becomes the not-found error);
- use transactions for operations that must be atomic;
- return errors unchanged unless meaning or context must change.

A service never imports `transport/`, never writes HTTP responses, and never reads environment variables.

## 5. Queries

- SQL is written in the service, as `const` strings at the top of the file, named for what they do: `selectClientByIDQuery`, `insertClientQuery`, `suspendClientQuery`.
- `?` placeholders; values passed as arguments; explicit column lists.
- Queries on owner-scoped data include the owner filter.
- Dynamic sort or filter parts come from an allow-list in code.
- When the constants grow large, move them to `client_query.go` in the same package.
- SQL never goes in `infrastructure/`, `migration/`, or a separate "sql" folder.

## 6. Transactions

Use the platform's transaction helper (`InTransaction`) for operations that must be atomic. Pass the callback's context to every call inside it. Keep transactions short and keep external calls (HTTP, queues, email) out of them; do those after commit.

## 7. Cross-domain calls

When service A needs behavior from domain B, A declares the small interface it needs and receives B's service through its constructor:

```go
package ordersvc

type clientReader interface {
	GetByID(ctx context.Context, clientID uuid.UUID) (*client.Client, error)
}

type OrderService struct {
	database     db.DB
	clientReader clientReader
}
```

`main.go` passes the concrete `*clientsvc.ClientService`. A may import domain B's **model** package (`internal/client`) for types and errors, never B's service package.

For simple reads of another domain's table inside a join, keep the SQL in the service that owns the use case; do not reach into another domain's service just to run SQL.

## 8. Search and pagination

Use the platform's pagination/meta types in result structs; do not define a new pagination type. A search method takes a query struct and a page request and returns a result struct. If search logic becomes substantial, move it to `client_search_service.go`.

## 9. Splitting rules

Start with one file. Split when complexity justifies it, never by function count.

| Signal | Action |
|---|---|
| The file is hard to scan | Split the largest independent concept out |
| A workflow has many steps and helpers | Give it its own file (`client_import.go`) |
| Search has its own logic and queries | `client_search_service.go` |
| SQL constants are numerous | `client_query.go` |
| `model.go` mixes clearly separate concepts | Split by concept, not per struct |

All files in `service/` share the package `clientsvc`; the service is still one type.

## 10. Testing placement

- Unit tests next to the code: `client_service_test.go`.
- SQL correctness, constraints, and transactions are covered by integration tests (`integration` build tag) against a real PostgreSQL test database, since there is no repository to fake.
- External systems behind small interfaces are faked in unit tests.
