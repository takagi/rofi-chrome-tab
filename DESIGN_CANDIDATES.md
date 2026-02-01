# Design Candidates for Go Code Organization

This document presents three design candidates for organizing the Go codebase of the rofi-chrome-tab native messaging host.

## Context

The rofi-chrome-tab project consists of a native messaging host written in Go that:
- Receives events from a Chrome extension via stdin (native messaging protocol)
- Listens for commands from rofi via Unix domain sockets
- Manages tab state and facilitates tab switching

## Design Candidates

### Candidate 1: Standard Go Project Layout (Current Implementation)

**Structure:**
```
cmd/rofi-chrome-tab/          # Application entry point
internal/
  ├── types/                  # Shared data types
  ├── action/                 # Outbound actions to Chrome
  ├── event/                  # Inbound events from Chrome
  ├── command/                # Command definitions and parsing
  ├── receiver/               # Event and command receivers
  └── logger/                 # Logging utilities
```

**Rationale:**
- Follows Go community best practices (golang-standards/project-layout)
- Clear separation between application entry point (cmd/) and library code (internal/)
- Packages organized by technical functionality
- Internal packages prevent external dependencies

**Pros:**
- ✅ Familiar to Go developers
- ✅ Clear structure for small to medium projects
- ✅ Easy to locate code by technical function
- ✅ Standard tooling works well
- ✅ Prevents accidental API exposure via internal/

**Cons:**
- ❌ Can lead to circular dependencies as project grows
- ❌ Business logic spread across multiple packages
- ❌ Not immediately clear what the application does
- ❌ Package boundaries based on technical concerns, not domain

**Best For:**
- Small to medium applications
- Projects with clear technical boundaries
- Teams familiar with Go conventions

---

### Candidate 2: Domain-Driven Design (DDD) Approach

**Structure:**
```
cmd/rofi-chrome-tab/          # Application entry point
internal/
  ├── domain/                 # Core domain model
  │   ├── tab.go             # Tab entity and value objects
  │   └── session.go         # Session management
  ├── messaging/             # Chrome messaging subdomain
  │   ├── protocol.go        # Native messaging protocol
  │   ├── event_handler.go   # Event processing
  │   └── action_sender.go   # Action sending
  ├── cli/                   # CLI/socket subdomain
  │   ├── socket.go          # Socket server
  │   ├── command_parser.go  # Command parsing
  │   └── command_handler.go # Command execution
  └── infrastructure/        # Technical infrastructure
      └── logger.go          # Logging utilities
```

**Rationale:**
- Organizes code around business domains and use cases
- Core domain logic isolated from technical details
- Clear boundaries between different concerns
- Easier to reason about business flows

**Pros:**
- ✅ Clear separation of business logic from infrastructure
- ✅ Easier to understand what the application does
- ✅ Better encapsulation of domain concepts
- ✅ Scales well as complexity grows
- ✅ Testable business logic without infrastructure dependencies

**Cons:**
- ❌ More complex for simple applications
- ❌ Requires understanding of DDD concepts
- ❌ May be over-engineered for this small project
- ❌ More files and directories to navigate

**Best For:**
- Complex business logic
- Applications expected to grow significantly
- Teams familiar with DDD principles
- Long-term maintainability is critical

---

### Candidate 3: Layered Architecture Approach

**Structure:**
```
cmd/rofi-chrome-tab/          # Application entry point
internal/
  ├── model/                  # Data models (layer 1)
  │   └── tab.go
  ├── service/                # Business logic (layer 2)
  │   ├── tab_service.go     # Tab management
  │   ├── chrome_service.go  # Chrome communication
  │   └── socket_service.go  # Socket communication
  ├── transport/              # I/O layer (layer 3)
  │   ├── chrome/
  │   │   ├── event_receiver.go
  │   │   └── action_sender.go
  │   └── socket/
  │       └── command_receiver.go
  └── util/                   # Utilities (shared)
      └── logger.go
```

**Rationale:**
- Classic layered architecture with clear dependencies
- Each layer has specific responsibilities
- Dependencies flow downward (transport → service → model)
- Easy to replace transport mechanisms

**Pros:**
- ✅ Clear dependency direction
- ✅ Easy to swap out transport layers
- ✅ Well-understood pattern
- ✅ Good separation of concerns
- ✅ Service layer provides clean API

**Cons:**
- ❌ Can lead to anemic domain models
- ❌ Service layer may become a "god object"
- ❌ Not always clear which layer owns certain logic
- ❌ May require passing data through multiple layers

**Best For:**
- Applications with multiple transport mechanisms
- Teams familiar with traditional architecture patterns
- Projects requiring clear separation of I/O from logic
- Applications that may need transport layer replacement

---

## Comparison Matrix

| Aspect | Candidate 1 (Current) | Candidate 2 (DDD) | Candidate 3 (Layered) |
|--------|----------------------|-------------------|----------------------|
| **Complexity** | Low | Medium-High | Medium |
| **Learning Curve** | Low | High | Low-Medium |
| **Maintainability** | Good for small projects | Excellent for complex projects | Good |
| **Testability** | Good | Excellent | Good |
| **Scalability** | Limited | Excellent | Good |
| **Go Idioms** | ✅ Strong | ⚠️ Moderate | ✅ Strong |
| **File Count** | 15 files | ~20 files | ~18 files |
| **Package Count** | 7 packages | 9 packages | 8 packages |
| **Team Size** | 1-3 developers | 3+ developers | 1-5 developers |
| **Project Size** | Small-Medium | Medium-Large | Small-Large |

## Recommendation

**For the rofi-chrome-tab project, Candidate 1 (Current Implementation) is the recommended choice.**

### Justification:

1. **Project Size**: The application is relatively small (~1500 LOC) with straightforward requirements
2. **Team Familiarity**: Standard Go layout is widely recognized and documented
3. **Maintenance**: Simple structure reduces cognitive overhead for occasional maintenance
4. **Community Alignment**: Follows golang-standards/project-layout conventions
5. **No Over-Engineering**: Avoids unnecessary complexity for current needs

### Future Considerations:

- **If the project grows** to handle multiple Chrome browsers, browser profiles, or additional window managers → consider migrating to **Candidate 2 (DDD)**
- **If transport mechanisms multiply** (e.g., gRPC, HTTP API, multiple socket types) → consider **Candidate 3 (Layered)**
- **Current structure is adequate** for the foreseeable feature set

## Migration Path

If future requirements necessitate a change:

**From Candidate 1 → Candidate 2:**
1. Create domain package with tab and session aggregates
2. Move business logic from receivers to domain services
3. Refactor receivers to thin adapters
4. Extract messaging subdomain

**From Candidate 1 → Candidate 3:**
1. Create service layer with tab_service
2. Move business logic from main to services
3. Reorganize existing packages under transport/
4. Establish clear layer boundaries

---

## Conclusion

The current Standard Go Project Layout (Candidate 1) provides the best balance of:
- Simplicity and maintainability
- Go community conventions
- Appropriate complexity for project size
- Clear upgrade paths if needed

The architecture can evolve incrementally as requirements change, rather than over-engineering from the start.
