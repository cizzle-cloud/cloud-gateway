# Changelog

## [0.2.0] - 2026-07-11

### Added

- Added 'TrustHeaders' in config
- Added request context with IP related methods
- Added middleware interfaces and core types
- Added router interface

### Fixed

- Fixed rate limit algo shared across clients instead of per-client instances

### Changed

- migrate rate limiter to v0.1.2 GetOrCreate API
- Minor changes
- Changed rate limit to allow injecting request Context in middleware
- Changed forward auth to allow injecting request Context in middleware
- Changed registry to incorporate request Context
- General refactoring
- Refactored forward auth middleware to remove gin coupling
- Refactored forward auth tests to remove gin coupling
- Refactored pkgs due to middleware interfaces additions
- Updated dependencies
- Refactored due to router interface addition
- Updated test CI job
- Updated Dockerfile path in release CI job
- Updated folder structure

## [0.1.0] - 2025-07-05

### Added
- Proxy routes
- Redirect routes
- Domain routes
- Rate limiting middleware (token bucket and fixed window counter)
- Forward authentication middleware
- YAML/JSON configuration support
- TLS termination
- Test, Release, Deploy CI jobs
- README and LICENSE
