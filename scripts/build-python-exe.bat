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

echo Installing or updating PyInstaller...
py -m pip install --upgrade pyinstaller
if errorlevel 1 goto :error

echo.
echo Building ProxyEnvSwitch.exe...
py -m PyInstaller ^
  --noconfirm ^
  --clean ^
  --onefile ^
  --windowed ^
  --name ProxyEnvSwitch ^
  --icon assets\proxyenv-switch.ico ^
  src\proxyenv_switch.pyw
if errorlevel 1 goto :error

echo.
echo Build completed successfully:
echo %CD%\dist\ProxyEnvSwitch.exe
pause
exit /b 0

:error
echo.
echo Build failed. Review the messages above.
pause
exit /b 1
