# macOS Build Safety Measures

This document outlines the multiple layers of protection implemented to prevent AppDelegate conflicts on macOS.

## The Problem

On macOS, both Wails and `fyne.io/systray` define an `AppDelegate` class in Objective-C, causing duplicate symbol errors during compilation:

```
duplicate symbol '_OBJC_METACLASS_$_AppDelegate'
duplicate symbol '_OBJC_CLASS_$_AppDelegate'
```

## The Solution: Multiple Layers of Protection

### 1. **Build Tags** (Primary Defense)

Three separate tray implementations with mutually exclusive build tags:

- **`manager_darwin.go`**: `//go:build darwin`
  - Uses DarwinKit for native NSStatusItem API
  - Always selected on macOS

- **`manager_systray.go`**: `//go:build !darwin && (windows || linux)`
  - Uses fyne.io/systray
  - Explicitly excluded on macOS

- **`manager_stub.go`**: `//go:build !darwin && !windows && !linux)`
  - No-op implementation for other platforms

### 2. **Build Command with `-tags nosystray`** (Critical)

The `-tags nosystray` flag prevents Go from compiling fyne.io/systray's Objective-C code:

```bash
# Development
wails dev -tags nosystray

# Production
wails build -skipbindings -tags nosystray
```

**Why This Is Necessary**: Even though build tags prevent importing fyne.io/systray in code, Go modules still download and attempt to compile the package's C/Objective-C files unless explicitly excluded with build tags.

### 3. **Automated Build Scripts** (Convenience)

#### Makefile (Recommended)
```bash
make dev    # Auto-detects macOS and uses -tags nosystray
make build  # Auto-detects macOS and uses -tags nosystray
```

The Makefile automatically applies correct flags based on OS:
```makefile
ifeq ($(DETECTED_OS),Darwin)
    EXTRA_BUILD_FLAGS := -tags nosystray
endif
```

#### Wrapper Scripts
- **`wails.dev.sh`** (Unix): Platform-aware development wrapper
- **`wails.dev.ps1`** (Windows): Development wrapper

### 4. **Compile-Time Safety Check** (Fail-Fast)

**File**: `internal/tray/build_check_darwin.go`

```go
//go:build darwin && !nosystray

// This file should NEVER compile on macOS.
// If it does, build tags are incorrect.
```

If someone tries to build on macOS without `-tags nosystray`, this file will compile and show a clear error message with instructions.

### 5. **Documentation Warnings** (Prevention)

Prominent warnings added to:

- **README.md**: Build instructions with macOS-specific warnings
- **README_zh.md**: Chinese version with warnings
- **docs/BUILD_REQUIREMENTS.md**: Detailed macOS requirements
- **AGENTS.md**: Development workflow guidelines
- **.github/copilot-instructions.md**: AI assistant instructions

### 6. **GitHub Actions Workflow** (CI/CD)

The release workflow automatically uses correct tags for macOS:

```yaml
- name: Build application (macOS)
  run: |
    wails build -clean -ldflags "-s -w" -skipbindings -tags nosystray \
      -platform darwin/universal
```

### 7. **Code Comments** (Developer Guidance)

Inline comments in tray files explain:
- Why each build tag is necessary
- The AppDelegate conflict issue
- Correct build commands

## Verification Checklist

When making changes, verify:

- [ ] `manager_darwin.go` has `//go:build darwin`
- [ ] `manager_systray.go` has `//go:build !darwin && (windows || linux)`
- [ ] `manager_stub.go` has `//go:build !darwin && !windows && !linux`
- [ ] Makefile uses `EXTRA_BUILD_FLAGS := -tags nosystray` for Darwin
- [ ] GitHub Actions uses `-tags nosystray` for macOS build
- [ ] `check.sh` uses `-tags nosystray` on macOS
- [ ] Documentation mentions macOS build requirements

## Testing

### On Windows/Linux
```bash
go build -v ./...  # Should succeed
```

### On macOS
```bash
# Should FAIL (missing nosystray tag):
go build -v ./...

# Should SUCCEED:
go build -v -tags nosystray ./...
make build
```

## Common Pitfalls

❌ **DON'T**:
- Use plain `wails build` on macOS
- Use plain `wails dev` on macOS
- Forget `-tags nosystray` when building manually

✅ **DO**:
- Use `make build` (recommended)
- Use `make dev` for development
- Use `./wails.dev.sh` wrapper script
- Always add `-tags nosystray` if calling wails directly

## Architecture Decision

We chose **DarwinKit over fyne.io/systray on macOS** because:

1. **Native API**: DarwinKit wraps native NSStatusItem directly
2. **No Conflicts**: Doesn't define AppDelegate
3. **Better Integration**: Follows macOS conventions
4. **Future-Proof**: Uses maintained Apple APIs

## Related Documentation

- [Build Requirements](BUILD_REQUIREMENTS.md)
- [Architecture](ARCHITECTURE.md)
- [Code Patterns](CODE_PATTERNS.md)
