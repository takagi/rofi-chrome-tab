# Design Candidates - Visual Comparison

## Candidate 1: Standard Go Project Layout (Current)

```
┌─────────────────────────────────────────────────────────────┐
│                    cmd/rofi-chrome-tab/                     │
│                      (main package)                         │
└────────────┬────────────────────────────────────────────────┘
             │ imports
             ▼
┌─────────────────────────────────────────────────────────────┐
│                      internal/                              │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐      │
│  │  types   │ │  logger  │ │  action  │ │  event   │      │
│  │   Tab    │ │ Logging  │ │  Select  │ │ Updated  │      │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘      │
│                                                              │
│  ┌──────────┐ ┌─────────────────────────────────┐          │
│  │ command  │ │         receiver                │          │
│  │List/Sel  │ │  EventRcv  |  CommandRcv        │          │
│  └──────────┘ └─────────────────────────────────┘          │
└─────────────────────────────────────────────────────────────┘

Flow: Chrome → stdin → EventReceiver → main → tabs
      rofi → socket → CommandReceiver → main → action → stdout → Chrome
```

## Candidate 2: Domain-Driven Design

```
┌─────────────────────────────────────────────────────────────┐
│                    cmd/rofi-chrome-tab/                     │
│                      (main package)                         │
└────────────┬────────────────────────────────────────────────┘
             │ imports
             ▼
┌─────────────────────────────────────────────────────────────┐
│                      internal/                              │
│                                                              │
│  ┌────────────────────────────────────────────────┐         │
│  │               domain/                          │         │
│  │  ┌──────────┐          ┌──────────┐           │         │
│  │  │   Tab    │          │ Session  │           │         │
│  │  │ (entity) │ ◄──────► │(aggregate)          │         │
│  │  └──────────┘          └──────────┘           │         │
│  └─────────────────┬──────────────────────────────┘         │
│                    │                                         │
│  ┌─────────────────┴─────────────┐  ┌──────────────────┐   │
│  │       messaging/              │  │      cli/        │   │
│  │  ┌────────────────┐           │  │  ┌────────────┐  │   │
│  │  │ Event Handler  │           │  │  │   Socket   │  │   │
│  │  │ Action Sender  │           │  │  │  Commands  │  │   │
│  │  │   Protocol     │           │  │  │  Handler   │  │   │
│  │  └────────────────┘           │  │  └────────────┘  │   │
│  └───────────────────────────────┘  └──────────────────┘   │
│                                                              │
│  ┌─────────────────────────────────────────────┐            │
│  │         infrastructure/                     │            │
│  │           Logger, Config                    │            │
│  └─────────────────────────────────────────────┘            │
└─────────────────────────────────────────────────────────────┘

Flow: Chrome → messaging → domain → Session
      rofi → cli → domain → Session → messaging → Chrome
```

## Candidate 3: Layered Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    cmd/rofi-chrome-tab/                     │
│                      (main package)                         │
└────────────┬────────────────────────────────────────────────┘
             │ imports
             ▼
┌─────────────────────────────────────────────────────────────┐
│                      internal/                              │
│                                                              │
│  Layer 3 (Transport - I/O)                                  │
│  ┌────────────────────────┐  ┌─────────────────────┐        │
│  │    transport/chrome/   │  │  transport/socket/  │        │
│  │  EventReceiver         │  │  CommandReceiver    │        │
│  │  ActionSender          │  │                     │        │
│  └───────────┬────────────┘  └──────────┬──────────┘        │
│              │ depends on              │                    │
│              ▼                          ▼                    │
│  Layer 2 (Service - Business Logic)                         │
│  ┌──────────────────────────────────────────────┐           │
│  │              service/                        │           │
│  │  TabService  ChromeService  SocketService    │           │
│  └───────────────────────┬──────────────────────┘           │
│                          │ depends on                       │
│                          ▼                                   │
│  Layer 1 (Model - Data)                                     │
│  ┌──────────────────────────────────────────────┐           │
│  │              model/                          │           │
│  │                Tab                           │           │
│  └──────────────────────────────────────────────┘           │
│                                                              │
│  ┌──────────────────────────────────────────────┐           │
│  │         util/ (Shared - Logger)              │           │
│  └──────────────────────────────────────────────┘           │
└─────────────────────────────────────────────────────────────┘

Flow: Chrome → transport → service → model
      rofi → transport → service → model → service → transport → Chrome

Dependency Direction: transport → service → model
```

## Key Differences

### Package Organization Philosophy

**Candidate 1**: Technical/functional grouping
- Packages grouped by what they do technically (event, command, action, receiver)
- Horizontal slicing by technology

**Candidate 2**: Business domain grouping
- Packages grouped by business subdomain (messaging, cli, domain)
- Vertical slicing by business capability

**Candidate 3**: Layer-based grouping
- Packages grouped by architectural layer (transport, service, model)
- Horizontal slicing by responsibility level

### Dependency Flow

**Candidate 1**: 
```
main → {types, logger, action, event, command, receiver}
receiver → {event, command}
event → types
```

**Candidate 2**:
```
main → {messaging, cli, domain, infrastructure}
messaging → domain
cli → domain
domain → (self-contained)
```

**Candidate 3**:
```
main → transport → service → model
(strict layering with downward dependencies only)
```

### Scalability Characteristics

**Candidate 1**: 
- Scales horizontally by adding new technical packages
- May develop circular dependencies over time
- Good for <5 packages

**Candidate 2**:
- Scales vertically by adding new subdomains
- Domain boundaries prevent coupling
- Good for >5 packages

**Candidate 3**:
- Scales by expanding layers
- Layer discipline prevents coupling
- Good for any size with stable requirements

## Test Organization

### Candidate 1
```
Tests alongside implementation:
internal/action/action_test.go
internal/command/command_test.go
```

### Candidate 2
```
Tests by subdomain:
internal/messaging/event_handler_test.go
internal/domain/session_test.go
```

### Candidate 3
```
Tests by layer:
internal/service/tab_service_test.go
internal/transport/chrome/event_receiver_test.go
```
