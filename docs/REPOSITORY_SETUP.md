# Update the existing GitHub repository

The current public repository was uploaded without the hidden `.github` directory, so GitHub does not see the build and release workflows. The safest update method is Git, not browser drag-and-drop.

## 1. Extract this package

Assume the extracted package is:

```text
C:\Temp\proxyenv-switch-v1.2.0
```

## 2. Clone the existing repository

```powershell
git clone https://github.com/johncheungmk/proxyenv-switch.git C:\Temp\proxyenv-switch-github
```

## 3. Copy the updated package, including hidden files

```powershell
robocopy C:\Temp\proxyenv-switch-v1.2.0 C:\Temp\proxyenv-switch-github /E /COPY:DAT /R:1 /W:1 /XD .git build __pycache__ /XF *.pyc
```

`robocopy` return codes from 0 through 7 normally mean success or successful copying with differences.

## 4. Remove old tracked build outputs

Executables should be published as Release assets rather than stored in the source tree:

```powershell
cd C:\Temp\proxyenv-switch-github
git rm --cached --ignore-unmatch dist\ProxyEnvSwitch_x64.exe dist\ProxyEnvSwitch_ARM64.exe dist\SHA256SUMS.txt dist\README.txt
```

## 5. Confirm the hidden workflow files are included

```powershell
git add -A
git status
```

Confirm that the staged files include:

```text
.github/workflows/build.yml
.github/workflows/release.yml
.github/dependabot.yml
.gitignore
```

## 6. Commit and push

```powershell
git commit -m "Release ProxyEnv Switch 1.2.0"
git push origin main
```

Open the GitHub **Actions** tab and confirm that **Build Windows executables** starts and completes.

## 7. Publish the release

```powershell
git tag v1.2.0
git push origin v1.2.0
```

The release workflow should publish three executables, SHA-256 checksums, and a portable Windows ZIP.

## 8. Complete the GitHub About section

On the repository page, select the gear icon beside **About** and set:

**Description**

```text
A lightweight Windows 11 utility to add, update, or remove user-level HTTP_PROXY and HTTPS_PROXY environment variables.
```

**Topics**

```text
windows windows-11 proxy environment-variables python go networking utility
```

Also enable **Releases** and **Issues**.
