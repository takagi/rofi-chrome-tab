# Design Documentation Overview

This directory contains comprehensive design documentation for the Go codebase reorganization of rofi-chrome-tab.

## Documents

### 1. [DESIGN_CANDIDATES.md](./DESIGN_CANDIDATES.md)
**Main design document** presenting three architectural approaches for organizing the Go code.

- **Candidate 1**: Standard Go Project Layout (Current Implementation)
- **Candidate 2**: Domain-Driven Design Approach  
- **Candidate 3**: Layered Architecture Approach

Each candidate includes:
- Package structure
- Rationale and philosophy
- Pros and cons
- Comparison matrix
- Recommendation and migration paths

**Start here** to understand the design decisions.

### 2. [DESIGN_DIAGRAMS.md](./DESIGN_DIAGRAMS.md)
**Visual representations** of the three design candidates.

Contains:
- ASCII architecture diagrams
- Dependency flow charts
- Package organization visualizations
- Key differences between approaches
- Scalability characteristics

**Use this** for visual learners or quick reference.

### 3. [DESIGN_IMPLEMENTATION.md](./DESIGN_IMPLEMENTATION.md)
**Concrete code examples** showing how each design works in practice.

Includes:
- Implementation examples for tab selection flow
- Testing strategies for each approach
- Feature addition examples ("close tab")
- Performance characteristics
- Error handling patterns

**Read this** to understand practical implications of each design.

## Quick Summary

The project currently uses **Candidate 1: Standard Go Project Layout** because:

✅ Appropriate complexity for project size (~1500 LOC)  
✅ Follows Go community best practices  
✅ Easy to maintain for small teams  
✅ Clear upgrade paths if needs change  

## When to Reconsider

Consider migrating to a different design if:

- **→ Candidate 2 (DDD)**: Project grows to handle multiple browsers, profiles, or complex domain logic
- **→ Candidate 3 (Layered)**: Multiple transport mechanisms are needed (HTTP API, gRPC, etc.)

## Current Package Structure

```
cmd/rofi-chrome-tab/          # Application entry point
internal/
  ├── types/                  # Shared data types (Tab)
  ├── action/                 # Actions to send to Chrome
  ├── event/                  # Events from Chrome
  ├── command/                # Command definitions
  ├── receiver/               # Event and command receivers
  └── logger/                 # Logging utilities
```

## How to Use These Documents

1. **For new contributors**: Read DESIGN_CANDIDATES.md to understand why the code is organized this way
2. **For architecture discussions**: Use DESIGN_DIAGRAMS.md to visualize alternatives
3. **For implementation details**: Reference DESIGN_IMPLEMENTATION.md for concrete examples
4. **For future planning**: Consult the migration paths when considering changes

## Related Files

- `go.mod` - Module definition
- `.github/workflows/go-ci.yml` - CI configuration
- `cmd/rofi-chrome-tab/main.go` - Application entry point

## Questions?

If you have questions about the design choices:
1. Check the comparison matrix in DESIGN_CANDIDATES.md
2. Look at the concrete examples in DESIGN_IMPLEMENTATION.md
3. Review the pros/cons for each candidate
4. Open an issue for discussion

---

**Last Updated**: 2026-02-01  
**Version**: 1.0  
**Corresponds to**: Package reorganization PR
