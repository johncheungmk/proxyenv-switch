# GitHub update instructions

For the existing repository, follow:

```text
docs/REPOSITORY_SETUP.md
```

Use Git or `robocopy` so the hidden `.github` and `.gitignore` paths are included. Do not rely on selecting visible files in Windows File Explorer and dragging them into GitHub, because hidden files may be omitted.

After pushing `main`, create the release:

```powershell
git tag v1.2.0
git push origin v1.2.0
```
