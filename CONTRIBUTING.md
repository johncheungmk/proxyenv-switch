# Contributing

Contributions and issue reports are welcome.

## Development checks

Run the cross-platform Python tests:

```powershell
py -m unittest discover -s tests -v
```

Run all available Windows development checks:

```powershell
powershell -ExecutionPolicy Bypass -File scripts\test.ps1
```

Build the Python implementation:

```bat
scripts\build-python-exe.bat
```

Build the native Windows binaries:

```powershell
powershell -ExecutionPolicy Bypass -File scripts\build-native-exes.ps1
```

## Pull requests

Keep changes focused and include:

- A clear description of the problem and solution
- Testing details for Windows 11
- Screenshots for visible interface changes
- Confirmation that no unrelated environment variables or registry values are modified
- Updated tests and documentation when behavior changes

Do not include proxy credentials, access tokens, or institutional network details in issues, screenshots, or test data.
