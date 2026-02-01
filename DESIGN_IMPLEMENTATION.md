# Design Candidates - Implementation Examples

This document provides concrete code examples for each design candidate to illustrate the practical differences.

## Example Scenario: Handling Tab Selection

We'll show how a "select tab" command flows through each architecture.

---

## Candidate 1: Standard Go Project Layout (Current Implementation)

### File: `cmd/rofi-chrome-tab/main.go`
```go
func executeCommand(cmd command.Command, conn io.Writer, pid int) error {
    switch c := cmd.(type) {
    case command.SelectCommand:
        action.SendAction(os.Stdout, action.SelectAction{TabID: c.TabID})
        return nil
    // ...
    }
}
```

### File: `internal/command/command.go`
```go
type SelectCommand struct {
    TabID int
}

func ParseCommand(line string) (Command, error) {
    fields := strings.Fields(line)
    switch fields[0] {
    case "select":
        tabID, _ := strconv.Atoi(fields[1])
        return SelectCommand{TabID: tabID}, nil
    }
}
```

### File: `internal/action/action.go`
```go
type SelectAction struct {
    TabID int `json:"tabId"`
}

func SendAction(w io.Writer, a Action) error {
    // Encode and write to Chrome via stdout
    payload, _ := json.Marshal(a)
    // ... send via native messaging protocol
}
```

**Characteristics:**
- Direct, imperative flow
- Each package handles its own concern
- Main orchestrates the flow
- Simple to follow for small team

---

## Candidate 2: Domain-Driven Design Approach

### File: `cmd/rofi-chrome-tab/main.go`
```go
func main() {
    // Bootstrap dependencies
    chromeMessaging := messaging.NewChromeMessaging(os.Stdin, os.Stdout)
    socketCLI := cli.NewSocketCLI(pid)
    
    session := domain.NewSession(chromeMessaging)
    
    socketCLI.HandleCommands(func(cmd cli.Command) {
        session.ExecuteCommand(cmd)
    })
    
    chromeMessaging.HandleEvents(func(evt messaging.Event) {
        session.ProcessEvent(evt)
    })
}
```

### File: `internal/domain/session.go`
```go
// Session represents the core domain aggregate
type Session struct {
    tabs      []Tab
    messaging Messaging
}

func (s *Session) ExecuteCommand(cmd interface{}) error {
    switch c := cmd.(type) {
    case SelectTabCommand:
        return s.selectTab(c.TabID)
    }
}

func (s *Session) selectTab(tabID int) error {
    // Domain logic here
    if !s.hasTab(tabID) {
        return ErrTabNotFound
    }
    
    // Delegate to messaging
    return s.messaging.SelectTab(tabID)
}
```

### File: `internal/messaging/chrome_messaging.go`
```go
// ChromeMessaging is an adapter for Chrome native messaging
type ChromeMessaging struct {
    stdin  io.Reader
    stdout io.Writer
}

func (cm *ChromeMessaging) SelectTab(tabID int) error {
    action := SelectAction{TabID: tabID}
    return cm.sendAction(action)
}
```

### File: `internal/cli/socket_cli.go`
```go
// SocketCLI is an adapter for Unix socket commands
type SocketCLI struct {
    listener net.Listener
}

func (sc *SocketCLI) HandleCommands(handler func(Command)) {
    // Parse socket commands and invoke handler
}
```

**Characteristics:**
- Session is the core domain object
- Adapters (messaging, cli) provide interfaces
- Business logic in domain package
- Testable without I/O

---

## Candidate 3: Layered Architecture Approach

### File: `cmd/rofi-chrome-tab/main.go`
```go
func main() {
    // Layer 1: Model
    // (just data, no logic here)
    
    // Layer 2: Services
    tabService := service.NewTabService()
    chromeService := service.NewChromeService(os.Stdout, tabService)
    socketService := service.NewSocketService(pid, tabService, chromeService)
    
    // Layer 3: Transport
    eventReceiver := chrome.NewEventReceiver(os.Stdin, chromeService)
    commandReceiver := socket.NewCommandReceiver(pid, socketService)
    
    eventReceiver.Start()
    commandReceiver.Start()
    
    select {} // Run forever
}
```

### File: `internal/service/tab_service.go`
```go
// TabService manages tab state
type TabService struct {
    tabs []model.Tab
    mu   sync.RWMutex
}

func (ts *TabService) UpdateTabs(tabs []model.Tab) {
    ts.mu.Lock()
    defer ts.mu.Unlock()
    ts.tabs = tabs
}

func (ts *TabService) GetTab(tabID int) (*model.Tab, error) {
    ts.mu.RLock()
    defer ts.mu.RUnlock()
    
    for _, tab := range ts.tabs {
        if tab.ID == tabID {
            return &tab, nil
        }
    }
    return nil, ErrTabNotFound
}
```

### File: `internal/service/chrome_service.go`
```go
// ChromeService handles Chrome communication
type ChromeService struct {
    writer     io.Writer
    tabService *TabService
}

func (cs *ChromeService) SelectTab(tabID int) error {
    // Verify tab exists
    if _, err := cs.tabService.GetTab(tabID); err != nil {
        return err
    }
    
    // Send action to Chrome
    action := SelectAction{TabID: tabID}
    return cs.sendAction(action)
}

func (cs *ChromeService) HandleEvent(event Event) error {
    switch e := event.(type) {
    case UpdatedEvent:
        cs.tabService.UpdateTabs(e.Tabs)
    }
    return nil
}
```

### File: `internal/transport/socket/command_receiver.go`
```go
// CommandReceiver receives commands from socket
type CommandReceiver struct {
    socketService *service.SocketService
}

func (cr *CommandReceiver) handleConnection(conn net.Conn) {
    scanner := bufio.NewScanner(conn)
    scanner.Scan()
    line := scanner.Text()
    
    cmd, err := ParseCommand(line)
    if err != nil {
        return
    }
    
    // Delegate to service layer
    cr.socketService.ExecuteCommand(cmd, conn)
}
```

**Characteristics:**
- Services contain all business logic
- Transport layer is thin (just I/O)
- Clear separation of concerns
- Easy to mock services for testing

---

## Testing Comparison

### Candidate 1: Test Implementation Details

```go
// internal/action/action_test.go
func TestSendAction(t *testing.T) {
    var buf bytes.Buffer
    action := SelectAction{TabID: 42}
    
    err := SendAction(&buf, action)
    
    // Test implementation details (wire format)
    assert.NoError(t, err)
    // Verify binary protocol...
}
```

### Candidate 2: Test Business Logic

```go
// internal/domain/session_test.go
func TestSession_SelectTab(t *testing.T) {
    mockMessaging := &MockMessaging{}
    session := NewSession(mockMessaging)
    session.tabs = []Tab{{ID: 42}}
    
    err := session.ExecuteCommand(SelectTabCommand{TabID: 42})
    
    // Test business logic, not I/O
    assert.NoError(t, err)
    assert.True(t, mockMessaging.SelectTabCalled)
}
```

### Candidate 3: Test Service Layer

```go
// internal/service/chrome_service_test.go
func TestChromeService_SelectTab(t *testing.T) {
    tabService := NewTabService()
    tabService.UpdateTabs([]model.Tab{{ID: 42}})
    
    var buf bytes.Buffer
    chromeService := NewChromeService(&buf, tabService)
    
    err := chromeService.SelectTab(42)
    
    assert.NoError(t, err)
    // Can test either business logic or wire format
}
```

---

## Adding a New Feature: "Close Tab" Command

### Candidate 1: Changes Required

1. Add `CloseCommand` to `internal/command/command.go`
2. Add `CloseAction` to `internal/action/action.go`
3. Update `executeCommand()` in `cmd/rofi-chrome-tab/main.go`

**Files changed: 3**

### Candidate 2: Changes Required

1. Add `CloseTab()` method to `internal/domain/session.go`
2. Add `CloseTab()` to `internal/messaging/messaging.go` interface
3. Implement in `internal/messaging/chrome_messaging.go`
4. Add command parsing in `internal/cli/socket_cli.go`

**Files changed: 4**

### Candidate 3: Changes Required

1. Add `CloseTab()` to `internal/service/chrome_service.go`
2. Add `CloseTab()` to `internal/service/socket_service.go`
3. Add command parsing in `internal/transport/socket/command_receiver.go`

**Files changed: 3**

---

## Performance Characteristics

### Candidate 1
- **Latency**: Low (direct function calls)
- **Memory**: Low (minimal indirection)
- **CPU**: Low (no abstraction overhead)

### Candidate 2
- **Latency**: Medium (interface calls, potential indirection)
- **Memory**: Medium (more objects/interfaces)
- **CPU**: Low-Medium (minimal overhead from interfaces)

### Candidate 3
- **Latency**: Low-Medium (service layer adds one level)
- **Memory**: Medium (service objects)
- **CPU**: Low (minimal overhead)

**Note**: For this application, performance differences are negligible. All candidates are suitable.

---

## Error Handling Comparison

### Candidate 1: Localized Error Handling
```go
func executeCommand(cmd command.Command, conn io.Writer, pid int) error {
    switch c := cmd.(type) {
    case command.SelectCommand:
        if err := action.SendAction(os.Stdout, action.SelectAction{TabID: c.TabID}); err != nil {
            log.Printf("failed to send action: %v", err)
            return err
        }
    }
}
```

### Candidate 2: Domain Error Types
```go
// internal/domain/errors.go
var (
    ErrTabNotFound = errors.New("tab not found")
    ErrInvalidTabID = errors.New("invalid tab ID")
)

// Business logic returns domain errors
func (s *Session) selectTab(tabID int) error {
    if !s.hasTab(tabID) {
        return ErrTabNotFound  // Domain error
    }
}
```

### Candidate 3: Layered Error Handling
```go
// Service layer wraps transport errors
func (cs *ChromeService) SelectTab(tabID int) error {
    if _, err := cs.tabService.GetTab(tabID); err != nil {
        return fmt.Errorf("tab lookup failed: %w", err)
    }
    
    if err := cs.sendAction(action); err != nil {
        return fmt.Errorf("chrome communication failed: %w", err)
    }
}
```

---

## Conclusion

Each design candidate has different strengths:

- **Candidate 1**: Best for current project size, minimal overhead
- **Candidate 2**: Best for complex domain logic, long-term evolution
- **Candidate 3**: Best for multiple transports, clear separation

The choice depends on:
- Project complexity
- Team experience
- Growth expectations
- Maintenance priorities
