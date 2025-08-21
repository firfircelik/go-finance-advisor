# Contributing to Personal Finance Tracker

Thank you for your interest in contributing to the Personal Finance Tracker! This document provides guidelines and information for contributors.

## 🚀 Quick Start

1. **Fork the repository**
2. **Clone your fork**
   ```bash
   git clone https://github.com/YOUR_USERNAME/go-finance-advisor.git
   cd go-finance-advisor
   ```
3. **Install dependencies**
   ```bash
   go mod download
   ```
4. **Run tests**
   ```bash
   make test
   ```
5. **Start development server**
   ```bash
   make run
   ```

## 📋 Development Guidelines

### Code Style

- Follow standard Go conventions and formatting
- Use `gofmt` to format your code
- Run `golangci-lint` before submitting
- Write meaningful commit messages

### Architecture

This project follows **Clean Architecture** principles:

```
internal/
├── domain/          # Business entities
├── application/     # Use cases and business logic
└── infrastructure/  # External concerns (API, DB, etc.)
```

**Key Principles:**
- Domain layer has no external dependencies
- Application layer depends only on domain
- Infrastructure layer implements interfaces defined in inner layers

### Testing

- **Unit Tests**: Test individual components in isolation
- **Integration Tests**: Test component interactions
- **Benchmarks**: Performance testing for critical paths

**Test Coverage Requirements:**
- Minimum 80% coverage for new code
- All public functions must have tests
- Critical business logic requires 95%+ coverage

**Running Tests:**
```bash
# Unit tests
make test

# With coverage
make test-coverage

# Benchmarks
make benchmark

# Integration tests
make test-integration
```

### API Design

- Follow RESTful conventions
- Use proper HTTP status codes
- Implement comprehensive error handling
- Document all endpoints in OpenAPI spec
- Version APIs appropriately (`/api/v1/`)

### Database

- Use GORM for ORM operations
- Write migrations for schema changes
- Follow naming conventions (snake_case)
- Index frequently queried columns

## 🔧 Development Workflow

### 1. Issue Creation

- Check existing issues before creating new ones
- Use issue templates when available
- Provide clear reproduction steps for bugs
- Include acceptance criteria for features

### 2. Branch Naming

```
feature/add-expense-categories
bugfix/transaction-date-validation
hotfix/security-jwt-vulnerability
docs/api-documentation-update
```

### 3. Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add expense categorization
fix: resolve transaction date validation bug
docs: update API documentation
test: add unit tests for advisor service
refactor: improve database query performance
```

### 4. Pull Request Process

1. **Create feature branch** from `main`
2. **Implement changes** following guidelines
3. **Add/update tests** for your changes
4. **Update documentation** if needed
5. **Run full test suite**
6. **Submit pull request** with clear description

**PR Requirements:**
- [ ] Tests pass (`make test`)
- [ ] Code coverage maintained
- [ ] Linting passes (`make lint`)
- [ ] Documentation updated
- [ ] Breaking changes documented

### 5. Code Review

**For Reviewers:**
- Focus on code quality, security, and maintainability
- Provide constructive feedback
- Test the changes locally when possible
- Approve only when confident in the changes

**For Authors:**
- Respond to feedback promptly
- Make requested changes or discuss alternatives
- Keep PRs focused and reasonably sized

## 🏗️ Project Structure

```
go-finance-advisor/
├── cmd/api/                 # Application entry point
├── internal/
│   ├── domain/             # Business entities
│   ├── application/        # Use cases
│   └── infrastructure/     # External interfaces
├── tests/                  # Integration tests
├── benchmarks/             # Performance tests
├── docs/                   # Documentation
├── examples/               # Usage examples
├── .github/workflows/      # CI/CD pipelines
├── Dockerfile              # Container configuration
├── Makefile               # Build automation
└── README.md              # Project overview
```

## 🧪 Testing Strategy

### Unit Tests
- Test individual functions/methods
- Mock external dependencies
- Focus on business logic

### Integration Tests
- Test component interactions
- Use test database
- Verify API endpoints

### Benchmarks
- Performance-critical operations
- Database query optimization
- Memory usage profiling

## 📚 Documentation

### Code Documentation
- Use Go doc comments for public APIs
- Include examples in documentation
- Keep comments up-to-date with code changes

### API Documentation
- Maintain OpenAPI specification
- Include request/response examples
- Document error scenarios

### README Updates
- Keep installation instructions current
- Update feature lists
- Maintain example usage

## 🔒 Security Guidelines

- **Never commit secrets** (API keys, passwords)
- **Validate all inputs** at API boundaries
- **Use parameterized queries** to prevent SQL injection
- **Implement proper authentication** and authorization
- **Follow OWASP guidelines** for web security

### Security Review Checklist
- [ ] Input validation implemented
- [ ] SQL injection prevention
- [ ] XSS protection
- [ ] Authentication/authorization checks
- [ ] Sensitive data handling

## 🚀 Performance Guidelines

- **Database Optimization**
  - Use appropriate indexes
  - Optimize query patterns
  - Implement connection pooling

- **API Performance**
  - Implement caching where appropriate
  - Use pagination for large datasets
  - Monitor response times

- **Memory Management**
  - Avoid memory leaks
  - Use appropriate data structures
  - Profile memory usage

## 🐛 Bug Reports

When reporting bugs, please include:

1. **Environment details** (OS, Go version, etc.)
2. **Steps to reproduce** the issue
3. **Expected behavior**
4. **Actual behavior**
5. **Error messages** or logs
6. **Screenshots** if applicable

## 💡 Feature Requests

For new features, please provide:

1. **Use case description**
2. **Proposed solution**
3. **Alternative solutions** considered
4. **Additional context**

## 📞 Getting Help

- **GitHub Issues**: For bugs and feature requests
- **GitHub Discussions**: For questions and general discussion
- **Code Review**: For implementation guidance

## 🎉 Recognition

Contributors will be recognized in:
- README.md contributors section
- Release notes for significant contributions
- GitHub contributor graphs

## 📄 License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

**Thank you for contributing to Personal Finance Tracker!** 🙏

Your contributions help make financial management more accessible and effective for everyone.