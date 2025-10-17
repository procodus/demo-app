# Contributing Guide

Thank you for your interest in contributing to the Demo App! This guide will help you get started.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Contribution Workflow](#contribution-workflow)
- [Development Guidelines](#development-guidelines)
- [Pull Request Process](#pull-request-process)
- [Issue Guidelines](#issue-guidelines)
- [Community](#community)

## Code of Conduct

### Our Pledge

We are committed to providing a welcoming and inclusive experience for everyone. We expect all contributors to:

- Use welcoming and inclusive language
- Be respectful of differing viewpoints and experiences
- Accept constructive criticism gracefully
- Focus on what is best for the community
- Show empathy towards other community members

### Unacceptable Behavior

- Harassment, trolling, or discriminatory comments
- Publishing others' private information
- Other conduct which could reasonably be considered inappropriate

## Getting Started

### Prerequisites

Before contributing, ensure you have:

1. **Go 1.25.3+** installed
2. **Git** configured with your name and email
3. **GitHub account** with SSH keys set up
4. Read the [Development Guide](./development.md)
5. Familiarized yourself with the [Architecture](./architecture.md)

### Fork and Clone

```bash
# Fork the repository on GitHub
# Then clone your fork
git clone git@github.com:YOUR_USERNAME/demo-app.git
cd demo-app

# Add upstream remote
git remote add upstream git@github.com:procodus/demo-app.git

# Verify remotes
git remote -v
```

### Set Up Development Environment

Follow the [Development Guide](./development.md) to set up your local environment.

## Contribution Workflow

### 1. Find or Create an Issue

Before starting work:

- Check existing issues: https://github.com/procodus/demo-app/issues
- If no issue exists, create one describing your proposed change
- Discuss approach with maintainers before significant work
- Wait for issue to be accepted before starting

### 2. Create a Feature Branch

```bash
# Update main branch
git checkout master
git pull upstream master

# Create feature branch
git checkout -b feature/your-feature-name

# Or for bug fixes
git checkout -b fix/issue-description
```

**Branch Naming Convention**:
- `feature/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation changes
- `refactor/` - Code refactoring
- `test/` - Test additions/improvements

### 3. Make Changes

- Write code following [Code Style Guidelines](#code-style-guidelines)
- Add tests for new functionality
- Update documentation as needed
- Commit frequently with clear messages

### 4. Test Your Changes

```bash
# Run unit tests
go test ./...

# Run E2E tests
go test ./test/e2e/...

# Run linter
golangci-lint run ./...

# Check formatting
go fmt ./...
```

### 5. Commit Changes

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```bash
git add .
git commit -m "feat: add new device type support"
```

**Commit Message Format**:
```
<type>: <subject>

<body>

<footer>
```

**Types**:
- `feat` - New feature
- `fix` - Bug fix
- `docs` - Documentation changes
- `style` - Code style changes (formatting, etc.)
- `refactor` - Code refactoring
- `test` - Adding or updating tests
- `chore` - Maintenance tasks

**Examples**:
```bash
feat: add pagination to sensor readings API

Add cursor-based pagination support to GetSensorReadingByDeviceID
method. This improves performance for devices with large reading
history.

Closes #123
```

```bash
fix: prevent consumer panic on malformed messages

Add validation for protobuf messages before processing. Invalid
messages are now logged and nacked instead of causing panic.

Fixes #456
```

### 6. Push Changes

```bash
git push origin feature/your-feature-name
```

### 7. Create Pull Request

1. Go to your fork on GitHub
2. Click "Compare & pull request"
3. Fill in the PR template (see below)
4. Link related issues
5. Request review from maintainers

## Development Guidelines

### Code Style Guidelines

#### Go Code Style

**Follow Standard Go Conventions**:
```go
// Good: Use camelCase for unexported names
func processMessage(data []byte) error {
    // ...
}

// Good: Use PascalCase for exported names
func ProcessMessage(data []byte) error {
    // ...
}

// Bad: Underscores in names (except for test files)
func process_message(data []byte) error {
    // ...
}
```

**Error Handling**:
```go
// Good: Check all errors
data, err := fetchData()
if err != nil {
    return fmt.Errorf("failed to fetch data: %w", err)
}

// Bad: Ignoring errors
data, _ := fetchData()
```

**Context Usage**:
```go
// Good: Pass context as first parameter
func ProcessMessage(ctx context.Context, msg []byte) error {
    // Use ctx for cancellation, deadlines
}

// Bad: Creating new background context
func ProcessMessage(msg []byte) error {
    ctx := context.Background()  // Don't do this
}
```

**Logging**:
```go
// Good: Use slog with structured fields
logger.Info("processing message",
    "device_id", deviceID,
    "timestamp", time.Now())

// Bad: Using fmt.Println
fmt.Println("processing message for", deviceID)  // Forbidden
```

#### Documentation

**Package Comments**:
```go
// Package generator provides utilities for generating synthetic IoT device data.
package generator
```

**Function Comments**:
```go
// ProcessMessage processes an incoming IoT device message and persists it to the database.
// Returns an error if the message is malformed or database operation fails.
func ProcessMessage(ctx context.Context, msg []byte) error {
    // ...
}
```

**Exported Types**:
```go
// Device represents an IoT device with metadata and configuration.
type Device struct {
    ID       string  // Unique device identifier
    Location string  // Physical location
    // ...
}
```

### Testing Guidelines

#### Unit Tests

**Use Ginkgo/Gomega**:
```go
var _ = Describe("DeviceProcessor", func() {
    var (
        processor *DeviceProcessor
        mockDB    *mock.Database
    )

    BeforeEach(func() {
        mockDB = mock.NewDatabase()
        processor = NewDeviceProcessor(mockDB)
    })

    Context("when processing valid device", func() {
        It("should save device to database", func() {
            device := &iot.IoTDevice{DeviceId: "device-001"}
            err := processor.Process(device)
            Expect(err).NotTo(HaveOccurred())
            Expect(mockDB.Devices).To(HaveLen(1))
        })
    })

    Context("when processing invalid device", func() {
        It("should return validation error", func() {
            device := &iot.IoTDevice{DeviceId: ""}  // Invalid
            err := processor.Process(device)
            Expect(err).To(HaveOccurred())
            Expect(err.Error()).To(ContainSubstring("device_id is required"))
        })
    })
})
```

**Test Coverage**:
- Aim for >80% coverage for new code
- Test happy paths and error cases
- Test edge cases and boundary conditions
- Use table-driven tests for multiple scenarios

#### E2E Tests

**Use Eventually for Async Operations**:
```go
It("should process message asynchronously", func() {
    // Publish message
    err := mqClient.Push(ctx, msgBytes)
    Expect(err).NotTo(HaveOccurred())

    // Poll until processed (don't use time.Sleep)
    Eventually(func() int {
        devices, _ := client.GetAllDevice(ctx, &iot.GetAllDeviceRequest{})
        return len(devices.GetDevice())
    }, 10*time.Second, 500*time.Millisecond).Should(BeNumerically(">=", 1))
})
```

### Security Guidelines

**Never Commit Secrets**:
```bash
# Bad
db_password="my-secret-password"

# Good
db_password="${DB_PASSWORD}"  # Read from environment
```

**Input Validation**:
```go
// Validate all external input
if device.GetDeviceId() == "" {
    return status.Error(codes.InvalidArgument, "device_id is required")
}
```

**SQL Injection Prevention**:
```go
// Good: Use GORM (parameterized queries)
db.Where("device_id = ?", deviceID).First(&device)

// Bad: String concatenation
db.Raw("SELECT * FROM devices WHERE device_id = '" + deviceID + "'")  // Never do this
```

## Pull Request Process

### PR Template

When creating a PR, include:

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Related Issues
Closes #123

## Changes Made
- Added pagination to sensor readings API
- Updated documentation
- Added unit tests

## Testing
- [ ] Unit tests pass
- [ ] E2E tests pass
- [ ] Linter passes
- [ ] Manual testing completed

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] Tests added/updated
- [ ] No breaking changes (or documented)

## Screenshots (if applicable)
```

### PR Review Process

1. **Automated Checks**: All CI checks must pass
   - Tests
   - Linting
   - Build verification

2. **Code Review**: At least one maintainer approval required
   - Review feedback should be addressed
   - Discussions should be resolved

3. **Documentation**: Ensure documentation is updated

4. **Testing**: Verify tests cover new functionality

5. **Merge**: Maintainer will merge after approval

### PR Best Practices

- **Keep PRs Small**: Aim for <500 lines of changes
- **One Feature Per PR**: Don't mix unrelated changes
- **Write Clear Descriptions**: Explain what and why
- **Respond Promptly**: Address review comments quickly
- **Be Patient**: Reviews may take a few days

## Issue Guidelines

### Reporting Bugs

Use the bug report template:

```markdown
## Bug Description
Clear description of the bug

## Steps to Reproduce
1. Start backend service
2. Send device message
3. Observe error

## Expected Behavior
Device should be saved to database

## Actual Behavior
Error: foreign key constraint violated

## Environment
- OS: macOS 14.0
- Go version: 1.25.3
- Deployment: Docker

## Logs
```
[logs here]
```

## Additional Context
Screenshots, error messages, etc.
```

### Requesting Features

Use the feature request template:

```markdown
## Feature Description
Brief description of proposed feature

## Use Case
Why is this feature needed?

## Proposed Solution
How should it work?

## Alternatives Considered
Other approaches considered

## Additional Context
Any other information
```

### Asking Questions

Use GitHub Discussions for:
- General questions
- Usage help
- Architecture discussions
- Best practices

## Community

### Communication Channels

- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: Questions and discussions
- **Pull Requests**: Code contributions

### Getting Help

If you need help:
1. Check the [documentation](./README.md)
2. Search existing [issues](https://github.com/procodus/demo-app/issues)
3. Ask in [discussions](https://github.com/procodus/demo-app/discussions)
4. Create a new issue if needed

### Recognition

Contributors will be:
- Listed in release notes
- Mentioned in the README
- Acknowledged in the CONTRIBUTORS file

## Review Process

### What Reviewers Look For

- **Correctness**: Does the code work as intended?
- **Tests**: Are there adequate tests?
- **Style**: Does it follow project conventions?
- **Documentation**: Is it properly documented?
- **Performance**: Are there performance concerns?
- **Security**: Are there security issues?

### Addressing Review Comments

```bash
# Make changes based on feedback
git add .
git commit -m "refactor: address review comments"
git push origin feature/your-feature-name
```

**DO**:
- Thank reviewers for their time
- Ask questions if feedback is unclear
- Explain your reasoning if you disagree

**DON'T**:
- Take feedback personally
- Argue without constructive discussion
- Ignore review comments

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

## Thank You!

Your contributions make this project better for everyone. We appreciate your time and effort!

## Next Steps

- [Development Guide](./development.md) - Set up your environment
- [Architecture](./architecture.md) - Understand the system
- [Testing Guide](./testing.md) - Write effective tests
- [API Reference](./api.md) - API documentation
