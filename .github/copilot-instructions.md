# GitHub Copilot Instructions for MrRSS

> **Documentation**: [AGENTS.md](../AGENTS.md) | [Architecture](../docs/ARCHITECTURE.md) | [Code Patterns](../docs/CODE_PATTERNS.md) | [Testing](../docs/TESTING.md) | [Build Requirements](../docs/BUILD_REQUIREMENTS.md)

## Project Context

MrRSS is a modern, privacy-focused, cross-platform desktop RSS reader built with Wails (Go + Vue.js).

**Core Principles**: Privacy-first, cross-platform, modern UI, high performance, accessible

## Tech Stack

- **Backend**: Go 1.27+ with Wails v3 (beta) framework, SQLite with `modernc.org/sqlite`
- **Frontend**: Vue 3.5+ Composition API, Pinia, Tailwind CSS 3.3+, Vite 5+
- **Tools**: Wails CLI v3, npm, Go modules
- **Icons**: Phosphor Icons | **I18n**: vue-i18n (English/Chinese)

## Quick Patterns Reference

### Backend (Go)

**Key Principles**:

- Always use `context.Context` for exported methods
- Error wrapping with `fmt.Errorf("operation failed: %w", err)`
- Prepared statements for all database queries
- Proper cleanup with `defer`
- Input validation before processing

📚 **Full Patterns**: See [CODE_PATTERNS.md](../docs/CODE_PATTERNS.md#backend-patterns-go)

### Frontend (Vue 3)

When writing Vue components, follow this pattern:

```vue
<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useAppStore } from '@/stores/app';
import { useI18n } from 'vue-i18n';

// Props with proper typing
interface Props {
  item: Article;
  isActive?: boolean;
}
const props = withDefaults(defineProps<Props>(), {
  isActive: false
});

// Store and i18n
const store = useAppStore();
const { t } = useI18n();

// Reactive state
const isLoading = ref(false);

// Async operations with error handling
async function loadData() {
  isLoading.value = true;
  try {
    const data = await fetch('/api/articles').then(r => r.json());
    // Process data...
  } catch (error) {
    console.error('Failed to load:', error);
    window.showToast(t('error'), 'error');
  } finally {
    isLoading.value = false;
  }
}

onMounted(() => loadData());
</script>

<template>
  <div class="component-container">
    <!-- Content with proper i18n -->
  </div>
</template>
```

📚 **Full Patterns**: See [CODE_PATTERNS.md](../docs/CODE_PATTERNS.md#frontend-patterns-vue)

## Internationalization

Always use i18n for user-facing strings:

```vue
<!-- Template -->
<h1>{{ t('welcome') }}</h1>
<button :title="t('clickToOpen')">{{ t('open') }}</button>

<!-- Script -->
window.showToast(t('successMessage'), 'success');
```

## UI Components

### Common Patterns

**Card Container**:

```html
<div class="bg-bg-primary border border-border rounded-lg p-4">
  <h3 class="text-text-primary font-semibold">{{ t('title') }}</h3>
  <p class="text-text-secondary text-sm">{{ t('description') }}</p>
</div>
```

**Modal/Dialog**:

```html
<div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
  <div class="bg-bg-primary w-full max-w-2xl rounded-2xl shadow-2xl border border-border">
    <div class="modal-header">
      <h2 class="text-xl font-bold">{{ t('modalTitle') }}</h2>
      <button @click="close" class="btn-icon">
        <i class="ph ph-x"></i>
      </button>
    </div>
  </div>
</div>
```

### Toast Notifications

```javascript
// Success message
window.showToast(message, 'success');

// Error message
window.showToast(t('operationFailed'), 'error');

// Info message with custom duration
window.showToast(t('updateAvailable'), 'info', 5000);
```

### Confirm Dialogs

```javascript
const confirmed = await window.showConfirm(
  t('confirmDelete'),
  t('deleteWarning'),
  true  // isDanger - shows red confirmation button
);

if (confirmed) {
  // Proceed with dangerous operation
}
```

### Context Menu Pattern

```vue
<script setup>
import { useContextMenu } from '@/composables/ui/useContextMenu';

const { contextMenu, openContextMenu, closeContextMenu } = useContextMenu();

// Define menu items
const menuItems = [
  { label: t('edit'), action: 'edit', icon: 'ph-pencil' },
  { label: t('delete'), action: 'delete', icon: 'ph-trash', danger: true },
  { type: 'divider' },
  { label: t('markAsRead'), action: 'mark-read' }
];

// Handle right-click
function handleRightClick(event: MouseEvent, item: Article) {
  event.preventDefault();
  openContextMenu(event, menuItems, item);
}

// Handle menu action
function handleMenuAction(action: string, item: Article) {
  switch (action) {
    case 'edit':
      // Handle edit
      break;
    case 'delete':
      // Handle delete
      break;
    case 'mark-read':
      // Handle mark as read
      break;
  }
}
</script>

<template>
  <div @contextmenu="handleRightClick($event, item)">
    <!-- Item content -->
  </div>
</template>
```

## Settings Management (OPTIMIZED)

✅ **The settings system has been optimized with schema-driven code generation!**

### Quick Method (3 Steps)

**Step 1**: Edit `internal/config/settings_schema.json`

```json
"new_setting_key": {
  "type": "bool",
  "default": false,
  "category": "general",
  "encrypted": false,
  "frontend_key": "new_setting_key"
}
```

**Step 2**: Generate all code

```bash
go run tools/settings-generator/main.go
```

**Step 3**: Add UI (optional)

```vue
<SettingItem :title="t('newSettingKey')">
  <Toggle
    :model-value="settings.new_setting_key"
    @update:model-value="updateSetting('new_setting_key', $event)"
  />
</SettingItem>
```

### What Gets Generated Automatically

- ✅ Backend types and handlers
- ✅ Frontend types and composables
- ✅ Database initialization keys
- ✅ Default values

### Old Method (Deprecated)

The manual 8-file checklist is **no longer needed**. All new settings should use the schema-driven approach.

📚 **Complete Guide**: See [docs/SETTINGS.md](../docs/SETTINGS.md)

## Security Best Practices

### Input Validation

Always validate user inputs, especially URLs and file paths:

```go
// Validate URL format and scheme
func validateFeedURL(urlStr string) error {
    u, err := url.Parse(urlStr)
    if err != nil {
        return fmt.Errorf("invalid URL: %w", err)
    }

    if u.Scheme != "http" && u.Scheme != "https" {
        return errors.New("URL must use HTTP or HTTPS")
    }

    return nil
}

// Validate file path to prevent directory traversal
func validateFilePath(baseDir, filePath string) error {
    cleanPath := filepath.Clean(filePath)
    if !strings.HasPrefix(cleanPath, filepath.Clean(baseDir)) {
        return errors.New("invalid file path: path traversal detected")
    }
    return nil
}
```

### Safe Command Execution

**NEVER** use shell command concatenation:

```go
// ❌ BAD: Command injection vulnerability
cmd := exec.Command("sh", "-c", "rm " + filePath)

// ✅ GOOD: Use Go standard library
if err := os.Remove(filePath); err != nil {
    return fmt.Errorf("remove file: %w", err)
}

// ✅ GOOD: If external command is necessary, use separate args
cmd := exec.Command("installer.exe", "/S") // No concatenation
```

### File Operations

Always clean up temporary files and use proper error handling:

```go
// Schedule cleanup with timeout
scheduleCleanup := func(filePath string, delay time.Duration) {
    go func() {
        time.Sleep(delay)
        if err := os.Remove(filePath); err != nil {
            log.Printf("Failed to cleanup %s: %v", filePath, err)
        } else {
            log.Printf("Cleaned up temporary file: %s", filePath)
        }
    }()
}
```

## Don'ts

❌ **Don't**:

- Use `var` declarations in Vue (use `ref` or `reactive`)
- Hardcode user-facing strings (always use i18n `t()`)
- Use inline styles (use Tailwind classes or scoped styles)
- Forget error handling in async operations
- Use `any` type without strong justification
- Commit API keys, secrets, or sensitive data
- Use `v-html` for user content (XSS risk)
- Make breaking changes without migration path
- Use shell command concatenation (security risk)
- Create multiple deep watchers when one suffices
- Forget to clean up timers/intervals on component unmount
- Delete favorited articles during cleanup operations
- Use synchronous operations in UI thread for long tasks

## Do's

✅ **Do**:

- Use TypeScript with proper type annotations
- Follow existing code patterns and conventions
- Write comprehensive tests for new features
- Keep functions small and focused (single responsibility)
- Use meaningful variable and function names
- Handle all edge cases and error conditions
- Validate inputs thoroughly (URLs, file paths, user data)
- Log errors with appropriate context for debugging
- Use semantic HTML with proper ARIA attributes
- Debounce frequent operations (auto-save, search, etc.)
- Use `os.Remove()` instead of shell commands for file operations
- Clean up resources (timers, goroutines, event listeners) properly
- Preserve favorited articles during any cleanup operation
- Use prepared statements for all database queries
- Implement proper loading states and progress indicators
- Follow semantic versioning (MAJOR.MINOR.PATCH)
- Document exported functions and complex logic
- Use goroutines for concurrent operations
- Implement graceful degradation for network failures

## Quick Reference

**Build Commands**:

- Development: `wails3 dev`
- Production Build: `wails3 build`
- Important: MrRSS uses HTTP API, not Wails bindings

**Store Access**:

- `const store = useAppStore()`
- `const { t } = useI18n()`
- Theme: `store.theme` returns `'light'` or `'dark'`
- Language: `store.i18n.locale.value` returns `'en'` or `'zh'`

**UI Helpers**:

- Toast: `window.showToast(message, type)`
- Confirm: `await window.showConfirm(title, message, isDanger)`

**API Endpoints**:

- Settings: `GET/POST /api/settings`
- Articles: `GET /api/articles` with query params
- Progress: `GET /api/progress` for async operations

---

When generating code, prioritize:

1. **Correctness**: Code that works and handles errors properly
2. **Consistency**: Follow existing patterns in the codebase
3. **Clarity**: Easy to understand and maintain
4. **Performance**: Efficient queries, minimal re-renders, proper cleanup
5. **Security**: Input validation, safe file operations, no injection vulnerabilities
6. **User Experience**: Loading states, progress indicators, error messages
