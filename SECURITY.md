# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| 2.x     | Yes       |
| 1.x     | No        |

## Reporting a Vulnerability

If you discover a security vulnerability in MoniGo, please report it responsibly:

1. **Do not** open a public GitHub issue
2. Email `iyashjayesh@gmail.com` with:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

You should receive a response within 48 hours. We will work with you to understand the issue and coordinate disclosure.

## Security Considerations

MoniGo exposes a dashboard and API endpoints. When deploying in production:

- **Always use HTTPS** - MoniGo does not enforce TLS; deploy behind a TLS-terminating reverse proxy
- **Enable authentication** - Use `BasicAuthMiddleware`, `APIKeyMiddleware`, or a custom `AuthFunction`
- **Restrict network access** - Bind the dashboard to internal interfaces or use `IPWhitelistMiddleware`
- **Trusted proxy requirement** - `X-Forwarded-For` headers are trusted by default; only deploy behind a trusted reverse proxy when using IP-based access control
- **OTel transport** - The OTel exporter defaults to insecure gRPC; configure TLS for production collectors

## Known Limitations

- The `ViewFunctionMetrics` endpoint executes `go tool pprof` with user-provided function names. While `exec.Command` does not invoke a shell, function names should be validated against known traced functions.
- Rate limiting is per-process, in-memory only. It does not provide distributed rate limiting.
