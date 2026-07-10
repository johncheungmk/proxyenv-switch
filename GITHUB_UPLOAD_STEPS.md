# Upload ProxyEnv Switch to GitHub

Suggested repository name:

```text
proxyenv-switch
```

Suggested description:

```text
A lightweight Windows 11 utility to add, update, or remove user-level HTTP_PROXY and HTTPS_PROXY environment variables.
```

## Option 1: Upload through the GitHub website

1. Sign in to GitHub.
2. Create a new repository named `proxyenv-switch`.
3. Do not add another README, `.gitignore`, or license because they are already included.
4. Extract the downloaded GitHub package.
5. On the empty repository page, choose **uploading an existing file**.
6. Upload all extracted files and folders.
7. Commit the files to the `main` branch.
8. Open the **Actions** tab and confirm that **Build Windows executables** completes successfully.

## Option 2: Upload with Git

Open PowerShell in the extracted folder and run:

```powershell
git init
git branch -M main
git add .
git commit -m "Initial release of ProxyEnv Switch"
git remote add origin https://github.com/johncheungmk/proxyenv-switch.git
git push -u origin main
```

Change the repository URL if a different GitHub account or repository name is used.

## Create the first GitHub release

After the initial upload and successful build:

```powershell
git tag v1.1.0
git push origin v1.1.0
```

The included release workflow will build and publish:

```text
ProxyEnvSwitch_x64.exe
ProxyEnvSwitch_ARM64.exe
ProxyEnvSwitch.exe
SHA256SUMS.txt
```

`ProxyEnvSwitch.exe` is the Python/Tkinter application packaged by PyInstaller on a Windows GitHub Actions runner. The architecture-specific files are dependency-free native builds.

## Recommended repository settings

Under **Settings → General → Releases**, keep GitHub Releases enabled.

Under **Settings → Actions → General**, allow GitHub Actions to run. The release workflow requires read and write permissions for repository contents; this permission is declared in the workflow.

Under **Settings → Security → Private vulnerability reporting**, enable private reporting when available.
