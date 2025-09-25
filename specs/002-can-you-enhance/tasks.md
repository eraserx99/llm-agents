# Tasks: mTLS Enhancement for MCP Servers

**Input**: Design documents from `/specs/002-can-you-enhance/`
**Prerequisites**: plan.md (required), research.md, data-model.md, contracts/

## Execution Flow (main)
```
1. Load plan.md from feature directory
   ✓ Tech stack: Go 1.25.1, Go standard library (crypto/tls, crypto/x509)
   ✓ Structure: Single project with internal/ structure
2. Load optional design documents:
   ✓ data-model.md: TLSConfig, Certificate, MCPServerConfig, MCPClientConfig
   ✓ contracts/: 3 contract files → 3 contract test tasks
   ✓ research.md: Certificate generation, TLS configuration decisions
3. Generate tasks by category:
   ✓ Setup: certificate generation utility, TLS infrastructure
   ✓ Tests: contract tests, integration tests, mTLS connection tests
   ✓ Core: TLS configuration, certificate management, server/client extensions
   ✓ Integration: server startup with TLS, client connection with mTLS
   ✓ Polish: performance tests, security validation, documentation
4. Apply task rules:
   ✓ Different files = mark [P] for parallel
   ✓ Same file = sequential (no [P])
   ✓ Tests before implementation (TDD)
```

## Format: `[ID] [P?] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- Include exact file paths in descriptions

## Path Conventions
- **Single project**: Uses existing `internal/`, `cmd/`, `test/` structure
- Certificate files stored in `certs/` directory at repository root
- TLS configuration integrated into existing MCP framework

## Phase 3.1: Setup & Infrastructure

- [x] T001 Create certificate directory structure at repository root (`certs/`)
- [x] T002 [P] Create certificate generation utility in `cmd/cert-gen/main.go`
- [x] T003 [P] Add TLS configuration struct in `internal/config/tls.go`
- [x] T004 [P] Add certificate management package in `internal/tls/certs.go`

## Phase 3.2: Tests First (TDD) ⚠️ MUST COMPLETE BEFORE 3.3

**CRITICAL: These tests MUST be written and MUST FAIL before ANY implementation**

### Contract Tests [P]
- [x] T005 [P] Contract test TLS configuration API in `test/contract/tls_config_test.go`
- [x] T006 [P] Contract test MCP server TLS API in `test/contract/server_tls_test.go`
- [x] T007 [P] Contract test MCP client TLS API in `test/contract/client_tls_test.go`

### Integration Tests [P]
- [x] T008 [P] Test certificate generation and validation in `test/integration/cert_generation_test.go`
- [x] T009 [P] Test mTLS connection establishment in `test/integration/mtls_connection_test.go`
- [x] T010 [P] Test weather MCP server with TLS in `test/integration/weather_tls_test.go`
- [x] T011 [P] Test datetime MCP server with TLS in `test/integration/datetime_tls_test.go`
- [x] T012 [P] Test echo MCP server with TLS in `test/integration/echo_tls_test.go`

### Security Tests [P]
- [x] T013 [P] Test client certificate validation in `test/security/client_cert_test.go`
- [x] T014 [P] Test demo mode vs strict validation in `test/security/validation_modes_test.go`
- [x] T015 [P] Test connection rejection without certificates in `test/security/unauthorized_test.go`

## Phase 3.3: Core Implementation (ONLY after tests are failing)

### Certificate Management
- [x] T016 [P] Implement certificate generation functions in `internal/tls/certs.go`
- [x] T017 [P] Implement TLS configuration validation in `internal/config/tls.go`
- [x] T018 [P] Add certificate loading and parsing in `internal/tls/loader.go`

### Server TLS Extensions
- [x] T019 Extend MCP server with TLS support in `internal/mcp/server/server.go`
- [x] T020 Add TLS listener configuration in `internal/mcp/server/tls.go`
- [x] T021 Implement mutual TLS handshake validation in `internal/mcp/server/mtls.go`

### Client TLS Extensions
- [x] T022 Extend MCP client with TLS support in `internal/mcp/client/client.go`
- [x] T023 Add client certificate configuration in `internal/mcp/client/tls.go`
- [x] T024 Implement TLS connection pool in `internal/mcp/client/pool.go`

### MCP Server Implementations
- [x] T025 [P] Update weather MCP server with TLS support in `cmd/weather-mcp/main.go`
- [x] T026 [P] Update datetime MCP server with TLS support in `cmd/datetime-mcp/main.go`
- [x] T027 [P] Update echo MCP server with TLS support in `cmd/echo-mcp/main.go`

### Coordinator Agent Updates
- [x] T028 Update coordinator agent to use TLS clients in `internal/agents/coordinator/coordinator.go`
- [x] T029 Add TLS configuration loading in `internal/agents/coordinator/tls_config.go`

## Phase 3.4: Integration & Configuration

- [ ] T030 Add environment variable configuration for TLS in `internal/config/config.go`
- [ ] T031 Add TLS status monitoring and health checks in `internal/utils/health.go`
- [ ] T032 Implement TLS connection logging in `internal/utils/logger.go`
- [ ] T033 Add certificate expiration monitoring in `internal/tls/monitor.go`

## Phase 3.5: Command Line Tools

- [ ] T034 [P] Complete certificate generation CLI in `cmd/cert-gen/main.go`
- [ ] T035 [P] Add TLS connection test utility in `cmd/test-tls/main.go`
- [ ] T036 [P] Add certificate validation utility in `cmd/cert-check/main.go`
- [ ] T037 [P] Add performance comparison tool in `cmd/perf-test/main.go`

## Phase 3.6: Polish & Validation

### Unit Tests [P]
- [ ] T038 [P] Unit tests for certificate generation in `test/unit/cert_generation_test.go`
- [ ] T039 [P] Unit tests for TLS configuration in `test/unit/tls_config_test.go`
- [ ] T040 [P] Unit tests for certificate validation in `test/unit/cert_validation_test.go`

### Performance & Security
- [ ] T041 Performance test TLS vs HTTP latency in `test/performance/latency_test.go`
- [ ] T042 Memory usage test for TLS connections in `test/performance/memory_test.go`
- [ ] T043 Security audit of certificate handling in `test/security/audit_test.go`

### Documentation & Validation
- [ ] T044 [P] Update main project README with TLS setup instructions
- [ ] T045 [P] Create TLS troubleshooting guide in `docs/tls-troubleshooting.md`
- [ ] T046 Execute quickstart.md validation workflow
- [ ] T047 Manual security testing and validation

## Dependencies

### Phase Dependencies
- Setup (T001-T004) before all other phases
- Tests (T005-T015) before implementation (T016-T037)
- Core implementation (T016-T029) before integration (T030-T033)
- Integration before command line tools (T034-T037)
- Implementation before polish (T038-T047)

### Specific Dependencies
- T003 (TLS config struct) blocks T017 (TLS validation)
- T004 (certificate management) blocks T016 (certificate generation)
- T016 (certificate generation) blocks T018 (certificate loading)
- T018 (certificate loading) blocks T019-T024 (server/client TLS)
- T019-T024 (TLS extensions) block T025-T029 (server implementations)
- T025-T029 (server implementations) block T030-T033 (integration)

## Parallel Execution Examples

### Phase 3.1 Setup [Parallel]
```bash
# Launch T002-T004 together:
Task: "Create certificate generation utility in cmd/cert-gen/main.go"
Task: "Add TLS configuration struct in internal/config/tls.go"
Task: "Add certificate management package in internal/tls/certs.go"
```

### Phase 3.2 Contract Tests [Parallel]
```bash
# Launch T005-T007 together:
Task: "Contract test TLS configuration API in test/contract/tls_config_test.go"
Task: "Contract test MCP server TLS API in test/contract/server_tls_test.go"
Task: "Contract test MCP client TLS API in test/contract/client_tls_test.go"
```

### Phase 3.2 Integration Tests [Parallel]
```bash
# Launch T008-T012 together:
Task: "Test certificate generation and validation in test/integration/cert_generation_test.go"
Task: "Test mTLS connection establishment in test/integration/mtls_connection_test.go"
Task: "Test weather MCP server with TLS in test/integration/weather_tls_test.go"
Task: "Test datetime MCP server with TLS in test/integration/datetime_tls_test.go"
Task: "Test echo MCP server with TLS in test/integration/echo_tls_test.go"
```

### Phase 3.3 Core Implementation [Mixed Parallel]
```bash
# Launch T016-T018 together (certificate management):
Task: "Implement certificate generation functions in internal/tls/certs.go"
Task: "Implement TLS configuration validation in internal/config/tls.go"
Task: "Add certificate loading and parsing in internal/tls/loader.go"

# Then launch T025-T027 together (server updates):
Task: "Update weather MCP server with TLS support in cmd/weather-mcp/main.go"
Task: "Update datetime MCP server with TLS support in cmd/datetime-mcp/main.go"
Task: "Update echo MCP server with TLS support in cmd/echo-mcp/main.go"
```

## Notes
- [P] tasks = different files, no dependencies
- Verify all tests fail before implementing (TDD critical)
- Test mTLS connections thoroughly in demo and strict modes
- Ensure backward compatibility with existing HTTP MCP servers
- Validate certificate generation works on all target platforms
- Monitor performance impact and optimize if needed

## Validation Checklist

**Contract Coverage**:
- [x] TLS configuration API → T005 contract test
- [x] MCP server TLS API → T006 contract test
- [x] MCP client TLS API → T007 contract test

**Entity Coverage**:
- [x] TLSConfig → T003, T017 implementation
- [x] Certificate → T004, T016 implementation
- [x] MCPServerConfig → T019-T021 server extensions
- [x] MCPClientConfig → T022-T024 client extensions

**Test Coverage**:
- [x] All contract tests before implementation
- [x] Integration tests for all 3 MCP servers
- [x] Security tests for certificate validation
- [x] Performance tests for TLS overhead

**Parallel Safety**:
- [x] All [P] tasks operate on different files
- [x] No shared file conflicts in parallel groups
- [x] Dependencies properly sequenced

**File Path Specificity**:
- [x] Every task specifies exact file path
- [x] Follows existing project structure conventions
- [x] Clear separation between test and implementation files