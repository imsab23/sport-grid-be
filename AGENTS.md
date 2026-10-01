# Repository Instructions

These rules apply to all work in this repository, including the Megamind and Forge skills.

## Project Principles

- Follow the existing repository architecture. Inspect it before making architectural or implementation decisions.
- Do not introduce a framework, library, pattern, or architectural style without a clear justification.
- Prefer simple, maintainable, idiomatic Go; favor explicit code over clever code.
- Keep business logic independent from infrastructure. Where the existing architecture uses DDD or Clean Architecture, preserve it and keep domain, application/use-case, infrastructure, transport/controller, and configuration concerns clearly separated.
- Keep dependencies pointing inward. Infrastructure must not contain business rules.
- Avoid unnecessary abstractions and premature optimization.

## Engineering Rules

- Never modify code blindly. Inspect actual files, validate assumptions against the repository, reuse existing utilities and patterns, and avoid duplicating functionality.
- Preserve backward compatibility unless the task explicitly requires a breaking change.
- Run appropriate tests after implementation. Run formatting and configured static checks when applicable.
- Never claim completion without verifying the result. Report checks that could not be run and why.

## Database Rules

- Respect the existing schema and migration strategy. Never silently change production-sensitive data behavior.
- Make transaction boundaries explicit where atomicity is required; consider concurrency and race conditions.
- Maintain appropriate indexes and constraints. Keep business logic out of SQL unless there is a clear reason.

## API Rules

- Follow existing API conventions and request/response structures.
- Validate input at the appropriate boundary and return appropriate errors without exposing internal implementation details.
- Preserve existing API contracts unless a change is explicitly required.

## Security Rules

Consider authentication, authorization, tenant isolation, input validation, IDOR/BOLA, SQL injection, sensitive data exposure, race conditions, replay attacks, duplicate requests, transaction consistency, privilege escalation, and sensitive information in logs as relevant to the change.

Never hardcode secrets, tokens, passwords, API keys, or credentials.

## Sport Grid Context

This repository is part of the Sport Grid backend. Where relevant, consider its multi-tenant tournament SaaS context: clients/organizations manage tournaments; player accounts are global and may join tournaments across clients; Super Admin controls global sports and tournament configuration; divisions and scoring rules may be configurable; payment verification and payment account/QR information are client-managed; authorization is RBAC/permission based; different sports may need different scoring models. The backend is Go with PostgreSQL, migrations, and Redis where appropriate; REST/HTTP and gRPC may apply. Strong tenant isolation matters.

Treat this as context, not proof that a feature belongs to any of these areas. Inspect code and requirements before applying it.

## Priority

Optimize for correctness, maintainability, security, clarity, performance, then implementation speed. Make performance work evidence-driven. Let architecture serve business requirements, code serve architecture, and tests verify behavior. Do not add complexity without reason.
