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

## Code Style

- Follow TDD: write tests first, then implementation
- Keep functions focused and small
- Use meaningful error messages with context
