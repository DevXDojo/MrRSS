# Contributing to MrRSS

Thank you for your interest in contributing to MrRSS! We welcome contributions from everyone.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [How to Contribute](#how-to-contribute)
- [Coding Guidelines](#coding-guidelines)
- [Testing](#testing)
- [Pull Request Process](#pull-request-process)
- [Reporting Bugs](#reporting-bugs)
- [Suggesting Enhancements](#suggesting-enhancements)

## Code of Conduct

This project and everyone participating in it is governed by our [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/MrRSS.git`
3. Add upstream remote: `git remote add upstream https://github.com/DevXDojo/MrRSS.git`
4. Create a new branch: `git checkout -b feature/your-feature-name`

## Development Setup

### Prerequisites

- Go 1.25 or higher
- Node.js (LTS version)
- Wails CLI: `go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0-alpha2.117`

**Linux only:** Install system dependencies:

```bash
# Ubuntu/Debian
sudo apt-get install libgtk-3-dev libwebkit2gtk-4.1-dev gcc pkg-config

# For older Ubuntu versions (before 24.04), use:
# sudo apt-get install libgtk-3-dev libwebkit2gtk-4.0-dev gcc pkg-config
```

### Setup Steps

1. Install frontend dependencies:

   ```bash
   cd frontend
   npm install
   cd ..
   ```

2. Run in development mode:

   ```bash
   wails3 dev
   ```

3. Build for production:

   ```bash
   wails3 build
   ```

## How to Contribute

### Types of Contributions

- 🐛 **Bug fixes**: Fix issues reported in GitHub Issues
- ✨ **Features**: Add new functionality
- 📝 **Documentation**: Improve or add documentation
- 🎨 **UI/UX**: Enhance the user interface
- 🌐 **Translations**: Add or improve translations
- ⚡ **Performance**: Optimize code performance
- ✅ **Tests**: Add or improve test coverage

### Workflow

1. **Check existing issues** - Look for related issues or create a new one
2. **Discuss first** - For major changes, discuss in an issue first
3. **Write code** - Follow our coding guidelines
4. **Write tests** - Add tests for new functionality
5. **Update documentation** - Update relevant documentation
6. **Submit PR** - Create a pull request with a clear description

## Coding Guidelines

### Go Backend

- Follow [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Use `gofmt` to format code
- Run `go vet` to check for common mistakes
- Add comments for exported functions and types
- Keep functions small and focused
- Handle errors explicitly

Example:

```go
// FetchFeed retrieves and parses an RSS feed from the given URL
func (f *Fetcher) FetchFeed(url string) (*Feed, error) {
    if url == "" {
        return nil, errors.New("url cannot be empty")
    }

    // Implementation...
}
```

### Vue.js Frontend

- Use Vue 3 Composition API
- Follow [Vue.js Style Guide](https://vuejs.org/style-guide/)
- Use TypeScript-style JSDoc comments for better IDE support
- Keep components small and reusable
- Use Tailwind CSS for styling
- Ensure responsive design

Example:

```javascript
<script setup>
import { ref, computed } from 'vue';

const props = defineProps({
    feed: { type: Object, required: true }
});

const isExpanded = ref(false);
</script>
```

### File Organization

- Backend: `internal/` for internal packages
- Frontend: `frontend/src/components/` for Vue components
- Tests: Co-locate tests with the code they test
- Assets: `frontend/assets/` for images, icons, etc.

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```plaintext
<type>(<scope>): <description>

[optional body]

[optional footer]
```

Types:

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

Examples:

```plaintext
feat(feed): add support for Atom feeds
fix(ui): resolve mobile layout issue
docs(readme): update installation instructions
```

## Testing

```bash
make test
```

### Cross-Platform Scripts

We provide automated scripts for running checks:

**Linux/macOS:**

```bash
# Run all quality checks (lint, test, build)
./scripts/check.sh

# Pre-release checks (includes version consistency)
./scripts/pre-release.sh
```

**Windows (PowerShell):**

```powershell
# Run all quality checks (lint, test, build)
.\scripts\check.ps1

# Pre-release checks (includes version consistency)
.\scripts\pre-release.ps1
```

### Manual Testing

Before submitting a PR, manually test:

1. Build the application
2. Test on your target platform
3. Verify all affected features work correctly
4. Check for console errors
5. Test edge cases

## Pull Request Process

1. **Update your branch** with the latest upstream changes:

   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Run tests** and ensure they all pass

3. **Update documentation** if you've changed APIs or added features

4. **Create a Pull Request** with:
   - Clear title describing the change
   - Description of what changed and why
   - Link to related issue(s)
   - Screenshots for UI changes
   - Notes about breaking changes (if any)

5. **Address review feedback** promptly

6. **Squash commits** if requested before merge

### PR Checklist

- [ ] Code follows project style guidelines
- [ ] Tests added/updated and passing
- [ ] Documentation updated
- [ ] Commit messages follow conventions
- [ ] No merge conflicts
- [ ] Screenshots included (for UI changes)

## Reporting Bugs

Use the [Bug Report template](.github/ISSUE_TEMPLATE/bug_report.md) and include:

- **Description**: Clear description of the bug
- **Steps to Reproduce**: Detailed steps
- **Expected Behavior**: What should happen
- **Actual Behavior**: What actually happens
- **Environment**: OS, version, etc.
- **Screenshots**: If applicable
- **Logs**: Relevant error messages

## Suggesting Enhancements

Use the [Feature Request template](.github/ISSUE_TEMPLATE/feature_request.md) and include:

- **Problem**: What problem does this solve?
- **Solution**: Your proposed solution
- **Alternatives**: Other solutions you considered
- **Additional Context**: Mockups, examples, etc.

## Questions?

- Check existing [Issues](https://github.com/DevXDojo/MrRSS/issues)
- Start a [Discussion](https://github.com/DevXDojo/MrRSS/discussions)
- Read the [README](README.md) and other documentation

## License

By contributing, you agree that your contributions will be licensed under the GPLv3 License.

---

Thank you for contributing to MrRSS! 🎉
