@echo off
setlocal
cd /d "%~dp0\.."

where py >nul 2>nul
if errorlevel 1 (
  echo Python Launcher ^(py.exe^) was not found.
  echo Install Python 3 for Windows and enable the Python Launcher.
  pause
  exit /b 1
)

echo Running unit tests...
py -m unittest discover -s tests -v
if errorlevel 1 goto :error

echo.
echo Installing the declared build dependency...
py -m pip install -r requirements-dev.txt
if errorlevel 1 goto :error

echo.
echo Building ProxyEnvSwitch.exe...
py -m PyInstaller ^
  --noconfirm ^
  --clean ^
  --onefile ^
  --windowed ^
  --paths src ^
  --name ProxyEnvSwitch ^
  --icon assets\proxyenv-switch.ico ^
  src\proxyenv_switch.pyw
if errorlevel 1 goto :error

echo.
for /f "tokens=*" %%H in ('powershell -NoProfile -Command "(Get-FileHash -Algorithm SHA256 'dist\ProxyEnvSwitch.exe').Hash.ToLower()"') do set HASH=%%H
> dist\ProxyEnvSwitch.sha256.txt echo %HASH%  ProxyEnvSwitch.exe

echo Build completed successfully:
echo %CD%\dist\ProxyEnvSwitch.exe
echo %CD%\dist\ProxyEnvSwitch.sha256.txt
pause
exit /b 0

:error
echo.
echo Build failed. Review the messages above.
pause
exit /b 1
