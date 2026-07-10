# Contributing

Contributions and issue reports are welcome.

## Development

For the Python implementation:

```powershell
py src\proxyenv_switch.pyw
```

For a packaged Python executable:

```bat
scripts\build-python-exe.bat
```

For native Windows binaries:

```powershell
powershell -ExecutionPolicy Bypass -File scripts\build-native-exes.ps1
```

## Pull requests

Please keep changes focused and include:

- A clear description of the problem and solution
- Testing details for Windows 11
- Screenshots for visible interface changes
- Confirmation that no unrelated environment variables are modified
