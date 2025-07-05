# Cloud Gateway

[![Version](https://img.shields.io/badge/version-0.1.0-blue.svg)](https://github.com/yourusername/cloud-gateway/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/yourusername/cloud-gateway)](https://goreportcard.com/report/github.com/yourusername/cloud-gateway)
[![License](https://img.shields.io/badge/license-Apache%202.0-green.svg)](LICENSE)

A highly configurable API gateway/reverse proxy with declarative YAML/JSON configuration, supporting advanced routing, middleware, and domain-based routing.

## Features

- **Flexible Routing**: Proxy routes, redirect routes, and domain-based routing
- **Advanced Middleware**: Rate limiting (token bucket & fixed window), forward authentication
- **Multiple Configuration Formats**: Support for both YAML and JSON configuration
- **Docker Ready**: Containerized deployment with Docker Compose support
- **TLS Support**: Built-in HTTPS/TLS termination

## Quick Start
For quick start and configuration examples, visit the [complete documentation](https://cizzle.cloud/services/cloud-gateway).

## Core Components

### Routing Types

- **Proxy Routes**: Forward requests to backend services
- **Redirect Routes**: HTTP redirects with configurable status codes
- **Domain Routes**: Route based on incoming domain headers

### Middleware

- **Rate Limiters**: Token bucket and fixed window algorithms
- **Forward Auth**: External authentication service integration
- **Custom Headers**: Request/response header manipulation

## Use Cases

- **API Gateway**: Centralized entry point for microservices
- **Reverse Proxy**: Load balancing and request routing
- **Authentication Gateway**: Centralized auth with forward auth middleware
- **Rate Limiting**: Protect backend services from abuse
- **Domain Routing**: Multi-tenant applications with domain-based routing
- **Service Migration**: Gradual migration with redirect routes

## Documentation

For detailed configuration options and advanced usage, visit the [complete documentation](https://cizzle.cloud/services/cloud-gateway).

## Security

- **TLS Termination**: Built-in HTTPS support
- **Authentication Integration**: Forward auth middleware
- **Rate Limiting**: Protection against abuse and DDoS
- **Trusted Proxy Support**: Proper handling of forwarded headers
- **Request Validation**: Input sanitization and validation

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Support

- **Documentation**: [cizzle.cloud](https://cizzle.cloud/services/cloud-gateway)
- **Issues**: [GitHub Issues](https://github.com/cizzle-cloud/cloud-gateway/issues)
- **Discussions**: [GitHub Discussions](https://github.com/cizzle-cloud/cloud-gateway/discussions)

## Changelog
For detailed changelog information, see [CHANGELOG](CHANGELOG.md).

---
**Made with ❤️ by [Cizzle Cloud](https://github.com/cizzle-cloud)**