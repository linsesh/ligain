---
name: transactions
description: Unit of Work pattern for atomic database operations
triggers:
  - atomic operations
  - transaction
  - rollback
  - multiple database operations
---

# Unit of Work Pattern

When multiple operations must be atomic (all succeed or all fail), use the Unit of Work pattern.

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│  Service Layer (database-agnostic)                      │
│  - Depends on UnitOfWork interface                      │
│  - Depends on Repository interfaces                     │
│  - Uses uow.WithinTx(ctx, fn) for atomic operations     │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│  Infrastructure Layer (postgres package)                │
│  - Implements UnitOfWork (starts tx, puts in context)   │
│  - Repositories check context for *sql.Tx              │
│  - If tx in context: use it; otherwise use *sql.DB     │
└─────────────────────────────────────────────────────────┘
```

## Key Files

- `backend/repositories/unit_of_work.go` - Interface definition
- `backend/repositories/postgres/unit_of_work.go` - Postgres implementation

## UnitOfWork Interface

```go
// Located in backend/repositories/unit_of_work.go
type UnitOfWork interface {
    WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}
```

## Service Layer Usage

Services use `UnitOfWork.WithinTx(ctx, fn)` for atomic operations:

```go
func (s *Service) AtomicOperation(ctx context.Context, ...) error {
    return s.uow.WithinTx(ctx, func(txCtx context.Context) error {
        // All repo calls use txCtx - same transaction
        if err := s.repoA.DoSomething(txCtx, ...); err != nil {
            return err  // Triggers rollback
        }
        if err := s.repoB.DoSomethingElse(txCtx, ...); err != nil {
            return err  // Triggers rollback
        }
        return nil  // Success = commit
    })
}
```

## Repository Layer Pattern

Repositories use an `executor(ctx)` helper to check for transactions:

```go
func (r *PostgresRepo) executor(ctx context.Context) DBExecutor {
    if tx := TxFromContext(ctx); tx != nil {
        return tx  // Use transaction
    }
    return r.db  // Use regular connection
}

func (r *PostgresRepo) DoSomething(ctx context.Context, ...) error {
    _, err := r.executor(ctx).ExecContext(ctx, query, ...)
    return err
}
```

## Adding UnitOfWork to a Service

1. Add `uow repositories.UnitOfWork` field to the service struct
2. Create a constructor that accepts UnitOfWork (e.g., `NewServiceWithUoW`)
3. Use `s.uow.WithinTx(ctx, fn)` in methods that need atomicity
4. Pass `txCtx` to all repository calls within the transaction function

## Testing Transactions

Mock UnitOfWork to test rollback scenarios:

```go
// MockUnitOfWork executes fn and returns its error (simulating real behavior)
func (m *MockUnitOfWork) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
    args := m.Called(ctx, fn)
    if fn != nil {
        fnErr := fn(ctx)
        if args.Error(0) != nil {
            return args.Error(0)
        }
        return fnErr
    }
    return args.Error(0)
}

// Test setup
mockUoW.On("WithinTx", mock.Anything, mock.Anything).Return(nil)
```

## When to Use

Use UnitOfWork when:
- Database write + cache update must be atomic
- Multiple database writes must all succeed or all fail
- You need to rollback on business logic failures

Don't use UnitOfWork when:
- Single read-only operation
- Operations that are already idempotent
- Fire-and-forget operations where partial success is acceptable
