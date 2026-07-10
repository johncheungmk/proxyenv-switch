# Security Policy

## Supported version

Security fixes are applied to the latest release.

## Reporting a vulnerability

Do not publish a suspected vulnerability in a public issue. Use GitHub private vulnerability reporting under the repository's **Security** tab.

Include:

- Affected version and executable name
- Windows version and architecture
- Steps to reproduce
- Expected and actual behavior
- Screenshots or logs with proxy credentials, tokens, and other sensitive information removed

## Application scope

ProxyEnv Switch modifies only:

```text
HKEY_CURRENT_USER\Environment\HTTP_PROXY
HKEY_CURRENT_USER\Environment\HTTPS_PROXY
```

It does not require administrator rights, install a service, collect telemetry, or make network requests. It does not modify the Windows system proxy.

## Release verification

Official GitHub Releases include `SHA256SUMS.txt`. Compare a downloaded executable with its published checksum before running it. The project binaries are not commercially code-signed, so Microsoft Defender SmartScreen may display an “unrecognized app” warning.
