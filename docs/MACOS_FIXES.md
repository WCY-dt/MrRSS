# macOS-Specific Fixes Documentation

This document describes macOS-specific fixes implemented to address issues on macOS, particularly macOS 15 (Sequoia).

## Issues Fixed

### 1. Window Dragging on Dual-Screen Setup ✅

**Problem**: On macOS with dual-screen setup, window dragging was not working properly when focus was on the second screen.

**Root Cause**: The default Wails `TitleBarHiddenInset()` configuration sets `TitlebarAppearsTransparent: true`, which causes dragging issues as documented in [Wails Issue #1346](https://github.com/wailsapp/wails/issues/1346).

**Solution**: 
- Updated `main.go` TitleBar configuration to explicitly set `TitlebarAppearsTransparent: false`
- Configured proper title bar settings for draggable window behavior
- Frontend already had proper `-webkit-app-region: drag` CSS for drag regions

**Files Changed**:
- `main.go` (lines 319-329)

**Code**:
```go
TitleBar: &mac.TitleBar{
    TitlebarAppearsTransparent: false, // Set to false to fix dragging issues
    HideTitle:                  true,
    HideTitleBar:               false,
    FullSizeContent:            true,
    UseToolbar:                 false,
    HideToolbarSeparator:       true,
},
```

---

### 2. Hide White Bar in Full-Screen Mode ✅

**Problem**: In full-screen mode, a white bar (title bar padding) remained visible at the top of the window.

**Solution**:
- Added fullscreen state detection in `App.vue`
- Dynamically remove the 28px top padding when in fullscreen mode
- Applied `macos-fullscreen` CSS class to hide padding

**Files Changed**:
- `frontend/src/App.vue` (template and styles)

**Code**:
```typescript
// Track fullscreen state
const isFullscreen = ref(false);

// Detect fullscreen changes
const checkFullscreen = () => {
  isFullscreen.value =
    document.fullscreenElement !== null || 
    (document as any).webkitFullscreenElement !== null;
};
```

```css
/* Hide title bar padding in fullscreen */
.app-container.macos-fullscreen {
  padding-top: 0;
}
```

---

### 3. Black Screen on Full-Screen Close ✅

**Problem**: When closing the window while in full-screen mode, the screen would turn black. This was caused by the window state restoration and minimize-to-tray features interfering with fullscreen exit.

**Solution**:
- Detect fullscreen state in `OnBeforeClose` handler
- Exit fullscreen explicitly before closing on macOS
- Skip minimize-to-tray logic when in fullscreen mode
- Don't store window state when in fullscreen (prevents invalid state storage)

**Files Changed**:
- `main.go` (OnBeforeClose handler and storeWindowState function)

**Code**:
```go
// In storeWindowState function
isFullscreen := runtime.WindowIsFullscreen(ctx)
if isFullscreen {
    log.Println("Window is in fullscreen mode, skipping state storage")
    return
}

// In OnBeforeClose handler
isFullscreen := runtime.WindowIsFullscreen(ctx)
if utils.IsMacOS() && isFullscreen {
    log.Println("Exiting fullscreen before closing on macOS")
    runtime.WindowUnfullscreen(ctx)
}

// Skip tray logic when in fullscreen
if shouldCloseToTray() && !isFullscreen {
    // ... minimize to tray logic
}
```

---

### 4. macOS Sequoia (15.x) WKWebView Crash Fix ✅

**Problem**: On macOS 15 (Sequoia), the app would crash due to excessive UI refresh rates causing WKWebView instability. This is documented in [Wails Issue #4592](https://github.com/wailsapp/wails/issues/4592).

**Root Cause**: WKWebView on macOS 15 has become more restrictive with frequent UI updates, causing crashes when polling intervals are too aggressive.

**Solution** (TEMPORARY WORKAROUND):
- Created macOS version detection utility
- Added platform API endpoint to expose OS version to frontend
- Automatically detect macOS 15+ and apply 2x polling intervals
- All changes marked with "TEMPORARY WORKAROUND" comments for easy removal

**Files Created**:
- `internal/utils/macos_version.go` - macOS version detection (Darwin only)
- `internal/utils/macos_version_stub.go` - Stub for non-macOS platforms
- `internal/handlers/platform/platform_handlers.go` - Platform API endpoint
- `frontend/src/composables/core/useThrottledInterval.ts` - Throttling utility (unused currently)

**Files Modified**:
- `main.go` - Added platform API route
- `frontend/src/composables/core/usePlatform.ts` - Added Sequoia detection
- `frontend/src/stores/app.ts` - Throttled progress polling
- `frontend/src/composables/core/useWindowState.ts` - Throttled window state checks
- `frontend/src/composables/discovery/useFeedDiscovery.ts` - Throttled discovery polling
- `frontend/src/composables/discovery/useDiscoverAllFeeds.ts` - Throttled batch discovery
- `frontend/src/composables/core/useAppUpdates.ts` - Throttled update progress

**API Endpoint**:
```
GET /api/platform/info

Response:
{
  "os": "darwin",
  "arch": "amd64",
  "is_macos": true,
  "is_windows": false,
  "macos_version": "15.0.0",
  "is_macos_sequoia": true,
  "needs_ui_throttling": true  // true for macOS 15+
}
```

**Interval Changes** (macOS 15+ only):
- Progress polling: 500ms → 1000ms
- Window state checks: 2000ms → 4000ms
- Feed discovery polling: 500ms → 1000ms
- Batch discovery polling: 500ms → 1000ms
- Update progress: 500ms → 1000ms

**Minimum Intervals**: All throttled intervals have a minimum of 1 second (1000ms).

---

## How to Remove Workarounds

When Apple fixes the WKWebView issue in future macOS versions, follow these steps:

### 1. Search for "TEMPORARY WORKAROUND" Comments

All throttling code is marked with this comment. Use grep to find them:

```bash
grep -r "TEMPORARY WORKAROUND" --include="*.ts" --include="*.go" .
```

### 2. Remove Backend Files

```bash
rm internal/utils/macos_version.go
rm internal/utils/macos_version_stub.go
rm internal/handlers/platform/platform_handlers.go
```

### 3. Remove Platform Route

In `main.go`, remove:
```go
import platform "MrRSS/internal/handlers/platform"  // Remove this line
// ...
apiMux.HandleFunc("/api/platform/info", ...)  // Remove this line
```

### 4. Revert Frontend Changes

In each modified file, remove the platform detection code and restore original intervals:

**`frontend/src/stores/app.ts`**:
- Remove `getPlatformInfo()` and `getOptimizedInterval()` functions
- Remove initialization call
- Change `pollInterval` back to `500`

**`frontend/src/composables/core/useWindowState.ts`**:
- Remove `getPlatformInfo()` and `getOptimizedInterval()` functions
- Change `intervalMs` back to `2000`

**`frontend/src/composables/discovery/useFeedDiscovery.ts`**:
- Remove `getPlatformInfo()` and `getOptimizedInterval()` functions
- Change `pollIntervalMs` back to `500`

**`frontend/src/composables/discovery/useDiscoverAllFeeds.ts`**:
- Same as above

**`frontend/src/composables/core/useAppUpdates.ts`**:
- Same as above

**`frontend/src/composables/core/usePlatform.ts`**:
- Remove `isMacOSSequoia`, `needsUIThrottling`, `macOSVersion` refs
- Remove platform info fetching logic
- Keep basic platform detection

### 5. Optional: Remove Unused File

```bash
rm frontend/src/composables/core/useThrottledInterval.ts
```

### 6. Test on macOS

After removing workarounds, test on the latest macOS version to ensure:
- No crashes occur with normal polling intervals
- All discovery and progress features work correctly
- Performance is acceptable

---

## Testing Checklist

### Issue 1: Window Dragging
- [ ] Test on dual-screen setup with macOS
- [ ] Click on screen 1, verify window can be dragged
- [ ] Click on screen 2, verify window can be dragged
- [ ] Verify dragging works from the title area

### Issue 2: Fullscreen White Bar
- [ ] Enter fullscreen mode (Cmd+Ctrl+F or green button)
- [ ] Verify no white bar at the top
- [ ] Move mouse to top, verify menu bar appears
- [ ] Verify content shifts down when menu appears (native behavior)

### Issue 3: Fullscreen Close
- [ ] Enter fullscreen mode
- [ ] Close window (Cmd+Q or close button)
- [ ] Verify no black screen appears
- [ ] Verify app closes cleanly

### Issue 4: macOS Sequoia Performance
- [ ] Run app on macOS 15 (Sequoia)
- [ ] Trigger feed refresh (should poll progress)
- [ ] Trigger feed discovery (should poll discovery progress)
- [ ] Check for app updates (should poll download progress)
- [ ] Verify no crashes occur
- [ ] Check console logs for throttling messages

---

## Notes

1. **macOS Versions**: macOS 15 is referred to as "Sequoia" or "macOS 26" in some contexts. The internal version number is 15.x.

2. **Build Tags**: The macOS version detection uses Go build tags (`// +build darwin` and `// +build !darwin`) to compile different code on different platforms.

3. **Thread Safety**: All platform detection code uses caching to avoid repeated API calls, which could cause performance issues.

4. **Fallback Behavior**: If platform detection fails, the code safely defaults to non-throttled intervals (original behavior).

5. **Other Platforms**: All fixes are macOS-specific and don't affect Windows or Linux behavior.

---

## References

- [Wails Issue #1346 - Title bar dragging](https://github.com/wailsapp/wails/issues/1346)
- [Wails Issue #4592 - macOS 26 WKWebView crashes](https://github.com/wailsapp/wails/issues/4592)
- [Wails Runtime API Documentation](https://wails.io/docs/reference/runtime/)
