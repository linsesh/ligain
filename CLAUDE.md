# Project Guidelines

## Architecture Principles

### Route Handlers (Thin Controllers)
Route handlers should be a **thin translation/routing layer** with minimal logic:
- Parse and validate request input
- Call the appropriate service method
- Translate service response to HTTP response

**Routes should NOT:**
- Contain business logic
- Orchestrate multiple operations
- Know about storage details or other infrastructure

**Services should:**
- Contain all business logic
- Orchestrate operations (e.g., delete old file → upload new file → update DB)
- Handle transactions and rollbacks
- Know what to do next

### Example - Bad (logic in route):
```go
func (h *Handler) UploadAvatar(c *gin.Context) {
    // ... parse request ...

    // BAD: Route is orchestrating operations
    if player.AvatarObjectKey != nil {
        go h.storageService.DeleteAvatar(ctx, *player.AvatarObjectKey)
    }
    objectKey, _ := h.storageService.UploadAvatar(ctx, player.ID, image)
    signedURL, _ := h.storageService.GenerateSignedURL(ctx, objectKey, ttl)
    h.playerRepository.UpdateAvatar(ctx, player.ID, objectKey, signedURL, expiresAt)

    c.JSON(http.StatusOK, gin.H{"avatar_url": signedURL})
}
```

### Example - Good (logic in service):
```go
func (h *Handler) UploadAvatar(c *gin.Context) {
    // Parse request
    image, err := parseAvatarFromRequest(c)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Call service (service knows what to do)
    result, err := h.profileService.UpdateAvatar(ctx, player.ID, image)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update avatar"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"avatar_url": result.SignedURL})
}
```

### Repository as Source of Truth (Multi-Service Architecture)

When multiple services operate on the same data, they MUST coordinate through repositories, not shared in-memory state.

**Critical Constraint:** Services MUST NOT cache mutable state locally.

**Why:**
- Service A adds a player via repository
- Service B queries players - it MUST see the new player immediately
- If Service B cached player lists, it would return stale data

**Caching Rules:**
1. Service registries may cache service instances (immutable references)
2. Services fetching mutable data (players, bets, scores) MUST query repositories
3. All state changes go through repositories so other services see them immediately

**Example - Bad (local caching):**
```go
type GameService struct {
    players []Player  // BAD: cached locally
}

func (s *GameService) GetPlayers() []Player {
    return s.players  // Stale data!
}
```

**Example - Good (repo query):**
```go
type GameService struct {
    playerRepo PlayerRepository
    gameID     string
}

func (s *GameService) GetPlayers() ([]Player, error) {
    return s.playerRepo.GetPlayersInGame(s.gameID)  // Always fresh
}
```

### Transactions with Unit of Work Pattern

When multiple operations must succeed or fail together (atomic), use `UnitOfWork.WithinTx(ctx, fn)`.
Services remain database-agnostic. See `backend/repositories/postgres/unit_of_work.go` for implementation.

**Example:**
```go
err := s.uow.WithinTx(ctx, func(txCtx context.Context) error {
    if err := s.repoA.DoSomething(txCtx, ...); err != nil {
        return err  // Triggers rollback
    }
    if err := s.repoB.DoSomethingElse(txCtx, ...); err != nil {
        return err  // Triggers rollback
    }
    return nil  // Success = commit
})
```

### Frontend: Components vs Logic

**Components** should be pure display with no business logic:
- Receive data via props or hooks
- Render UI
- Delegate actions to hooks/services

**Hooks/Services** should contain all testable logic:
- State management
- API calls
- Business rules

**Testing:**
- Test hooks and services thoroughly (real behavior tests)
- Don't write tests for trivial display components (no value)

**Example - Good separation:**
```tsx
// Hook contains testable logic
const useUpdateRequired = () => {
  const [isUpdateRequired, setIsUpdateRequired] = useState(false);
  // ... logic here, tested separately
  return { isUpdateRequired, storeUrl };
};

// Component is pure display, no tests needed
const UpdateRequiredModal = () => {
  const { isUpdateRequired, storeUrl } = useUpdateRequired();
  if (!isUpdateRequired) return null;
  return <Modal>...</Modal>;
};
```

## Key Conventions

**Testing — TDD is mandatory, not optional:**
- Mark tests with `@pytest.mark.unit` or `@pytest.mark.integration`
- Use `@pytest.mark.parametrize`when it is an adapted solution to simplify and consolidate tests
- Integration tests usually use testcontainers and docker
- **STRICT TDD PROTOCOL — follow these steps in order, never skip ahead:**
  1. **Write the test first.** Do NOT write any code before the test exists.
  2. **Ensure the test compiles.** Create any missing interfaces, stubs, or type definitions so the code has no import/type errors. RED means a failing assertion, not a compilation error.
  3. **Run the test and confirm it is RED.** Show the test output. If it passes already, the test is wrong — fix it before proceeding.
  4. **Only now write the minimal code** to make the test pass.
  5. **Run the test again and confirm it is GREEN.** Show the output.
  6. **Refactor if needed**, keeping tests green.
- **If you catch yourself writing code before a test exists, STOP immediately**, delete the code, and go back to step 1.
- **Test services with mocks, don't test mocks directly**: Mock implementations exist as test doubles to enable testing business logic. Test the service layer using mocks, not the mock implementation itself. Integration tests validate that real implementations behave correctly.

**Architecture:**
- Follow SOLID principles and clean architecture
- Use dependency inversion with interfaces - business code shouldn't know about infrastructure
- Keep classes focused on one use case

**Problem-Solving Approach:**
- **Strongly bias towards simple solutions**: Choose the simplest approach that solves the problem. Avoid over-engineering, premature optimization, or complex abstractions when straightforward solutions will work
- **Ask questions liberally in plan mode**: Instead of making assumptions about requirements, user preferences, or implementation details, ask clarifying questions to understand the exact needs and constraints
- Prefer direct, obvious implementations over clever or sophisticated ones
- When multiple approaches exist, default to the one with fewer moving parts and less complexity

**Communication Style:**
- **Challenge ideas when you disagree**: Don't be overly agreeable. If you see issues with an approach, say so directly and explain why. Be factual and honest rather than accommodating.

## Where to retrieve documentation

Always use Context7 MCP when I need library/API documentation, code generation, setup or configuration steps without me having to explicitly ask.


## Frontend: Full UI Rewrite in Progress

The frontend is being **completely rewritten** screen by screen. The old dark-theme UI with `StyleSheet.create()` is legacy — every screen will be rebuilt with NativeWind (Tailwind for RN) + react-native-reusables (RNR) components.

**Rewrite approach:**
- Convert screens **one at a time** — delete `StyleSheet.create` block when redesigning that screen, replace with NativeWind `className` props
- It's OK to break visual polish on screens not yet converted — they will all be redone
- Old `StyleSheet.create` code is throwaway; don't polish it, just leave it until the screen is rebuilt
- Light theme with grid background is the new visual foundation

**Color system:** `frontend/ligain/src/constants/colors.ts` is the source of truth. Tailwind semantic colors in `tailwind.config.js` mirror these values.

**RNR components:** installed individually as needed (e.g. `@rnr/button`, `@rnr/text`), placed in `src/components/ui/`.

**Global layout (`app/_layout.tsx`):**
- `SafeAreaView` wraps all content — no screen needs its own safe area handling
- `GridBackground` renders edge-to-edge behind everything — screens must use `backgroundColor: 'transparent'`
- `PortalHost` at the bottom for RNR modals/sheets

**NativeWind config files:**
- `frontend/ligain/global.css` — Tailwind imports (must be imported in `app/_layout.tsx`)
- `frontend/ligain/tailwind.config.js` — content paths, semantic colors
- `frontend/ligain/metro.config.js` — wrapped with `withNativeWind()`
- `frontend/ligain/nativewind-env.d.ts` — TypeScript reference

## Code Search Strategy

**Prefer grepai MCP tools over raw Grep/Glob for code exploration:**
- Use `grepai_search` as the **first choice** for semantic code questions ("how does X work?", "where is Y computed?", "find code related to Z"). It understands meaning, not just text patterns.
- Use `grepai_trace_callers` / `grepai_trace_callees` / `grepai_trace_graph` to follow call chains and understand data flow between functions.
- Fall back to Grep/Glob only for **exact literal matches** (specific variable names, error strings, import paths) where pattern matching is more precise than semantic search.
- When using the Explore subagent, instruct it to use grepai MCP tools as well.