# Tanki - AI Coding Agent Instructions

## Project Overview

**Tanki** is a terminal-based tank battle game written in Go. The player controls a tank (▲►▼◄) and must defeat enemy tanks (M) that spawn on the right side and chase the player. The game uses `termbox-go` for terminal rendering with a 40×20 character grid.

## Architecture & Key Components

### Game State (Global Variables)

- `player`: Player-controlled tank (Object)
- `enemies`: Slice of enemy tank pointers for dynamic list management
- `bullets`: Slice of active bullet pointers
- `walls`: Map[Point]bool for O(1) collision detection and destructible terrain
- `score`: Integer counter
- `isOver`: Boolean game state flag

**Why maps for walls?** Direct point lookup without iteration enables fast collision detection during bullet movement and enemy spawning.

### Core Game Loop (main function)

- Event channel: Handles keyboard input (arrow keys, Space, ESC) asynchronously
- Ticker: 70ms frame loop (≈14 FPS) drives game updates
- Select block: Non-blocking event handling with fallback to rendering on timeout

### Direction System

Direction is an `iota` enum (Up=0, Down=1, Left=2, Right=3). Used throughout for:

- Tank orientation (`drawTank` switches runes: ▲►▼◄)
- Bullet trajectory (`moveBullet` increments X or Y)
- Enemy AI (`updateEnemyAI` compares diffs to set direction)

## Critical Workflows & Patterns

### Movement & Collision Detection

```go
tryMove(obj *Object) bool  // Returns success/failure
```

**Pattern**: Always check in order: screen boundaries → walls → update position. This order prevents out-of-bounds writes.

### Bullet Lifecycle

1. `fireBullet(pos, dir)` → moves bullet once → appends to slice if still active
2. Each frame: `moveBullet(b)` → check bounds → **destroy wall** if hit (`delete(walls, b.Pos)`)
3. Dead bullets filtered out: `activeBullets := []*Bullet{}`

**Key insight**: Bullets that collide with walls during movement are removed from both the walls map and the bullets slice using separate filter loops in `updateState()`.

### Enemy AI (updateEnemyAI)

- **20% chance**: Random direction (prevents getting stuck)
- **80% chance**: Chase player by comparing `diffX` and `diffY` distances, moves along longest axis
- **10% chance to shoot** when aligned with player (same X or Y)
- **Movement throttle**: Enemy updates only happen with ~40% probability per frame (game loop line `if rand.Float32() < 0.4`)

### Collision & Game End

- **Enemy-Player collision**: Immediate game over
- **Bullet-Enemy collision**: Enemy removed, score += 100
- **Bullet-Player collision**: Game over
- Collision checks happen in `updateState()`, not during movement functions

## Development Guidelines

### Adding Features

1. **New tank types**: Add new constants (e.g., `BossChar`), create new Object slices (don't mix in `enemies` array)
2. **Weapon types**: Extend `Bullet` struct, modify `fireBullet()` and `moveBullet()` logic
3. **Map features**: Add to `walls` map in `generateWalls()`, check in `tryMove()` and `moveBullet()`
4. **UI elements**: Use `drawText()` and `drawNumber()` helpers; remember Y coordinate uses `Height` (20) for HUD, not `Height-1`

### Build & Run

```bash
go run main.go
```

Dependencies: `github.com/nsf/termbox-go` - must be installed via `go get`

### Testing Strategy

No test file exists. Manual testing is primary:

- **Player movement**: Arrow keys, boundary collision, wall collision
- **Enemy spawning**: 3 enemies initially, respawn when count < 3
- **Bullet mechanics**: Fire with Space, wall destruction, boundary exit
- **Score tracking**: +100 per enemy kill
- **Game over states**: Both collision types, restart on Space/any key

### Code Comments Convention

Comments are in Russian (Cyrillic). Maintain this for consistency with existing codebase. Use comments to explain non-obvious logic (e.g., why `moveBullet()` is called twice, why enemies have 40% update rate).

## Common Pitfalls

1. **Walls persistence**: Walls are destroyed when bullets hit. Remember to check if wall exists before rendering.
2. **Bullet movement ordering**: Bullet moves BEFORE being added to slice in `fireBullet()` - this is intentional to spawn bullets ahead of tank.
3. **Enemy pointer slices**: Use `[]*Object` not `[]Object` - dynamic deletion requires pointers for stability.
4. **Collision order in updateState()**: Must filter bullets AFTER checking collisions, or double-check `b.Active` flags.
5. **Screen boundaries**: X range is 0 to Width-1 (40), Y range is 0 to Height-1 (20). Use `<=` checks, not `<`.

## Integration Points

- **termbox-go library**: Low-level terminal control

  - `termbox.Init()` / `termbox.Close()`: Setup/cleanup
  - `termbox.SetCell(x, y, ch, fg, bg)`: Render single character
  - `termbox.PollEvent()`: Non-blocking input capture
  - `termbox.Flush()`: Apply all SetCell changes to screen
  - Color attributes: `ColorDefault`, `ColorRed`, `ColorGreen`, etc.

- **Standard library usage**:
  - `math.Abs()`: Distance calculation in enemy AI
  - `math/rand`: Enemy spawning, AI randomization
  - `time`: Frame tick, random seed initialization

## Files Reference

- `main.go`: Single-file monolithic game (no imports of local packages)

---

**Last updated**: December 2025 | Go 1.x | termbox-go required
