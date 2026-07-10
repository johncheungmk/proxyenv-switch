# Security Policy

## Supported version

Security fixes are applied to the latest release.

## Reporting a vulnerability

Please do not publish a suspected vulnerability in a public GitHub issue.
Use GitHub's private vulnerability reporting feature under the repository's
**Security** tab, or contact the repository owner privately.

Please include:

- The affected version
- Windows architecture and version
- Steps to reproduce
- Expected and actual behavior
- Any relevant screenshots or logs with sensitive information removed

## Application scope

ProxyEnv Switch modifies only these current-user registry values:

```text
HKEY_CURRENT_USER\Environment\HTTP_PROXY
HKEY_CURRENT_USER\Environment\HTTPS_PROXY
```

It does not require administrator rights, install a service, collect telemetry,
or make network requests.
