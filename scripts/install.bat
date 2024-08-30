@echo off
setlocal enabledelayedexpansion

:: Set variables
set "TOOL_NAME=Next"
set "INSTALL_DIR=%ProgramFiles%\%TOOL_NAME%"

:: Check for admin rights
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo Please run this script with administrator privileges.
    pause
    exit /b 1
)

:: Create installation directory
if not exist "%INSTALL_DIR%" mkdir "%INSTALL_DIR%"

:: Copy files and directories (except install.bat)
echo Copying files to installation directory...
for %%F in (*) do (
    if not "%%~nxF"=="install.bat" (
        if "%%~xF"=="" (
            xcopy "%%F" "%INSTALL_DIR%\%%F\" /E /I /Y
        ) else (
            copy "%%F" "%INSTALL_DIR%\"
        )
    )
)

:: Add Next/bin to PATH
echo Adding %INSTALL_DIR%\bin to system PATH...
for /f "tokens=2*" %%A in ('reg query "HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Environment" /v Path') do set "CURRENT_PATH=%%B"
setx PATH "%CURRENT_PATH%;%INSTALL_DIR%\bin" /M
if errorlevel 1 (
    echo Failed to add to PATH.
    goto :error
)

echo Installation complete.
echo %TOOL_NAME% has been installed to %INSTALL_DIR%
echo %INSTALL_DIR%\bin has been added to the system PATH.
echo Please restart your command prompt or PowerShell for PATH changes to take effect.
goto :eof

:error
echo An error occurred during installation.
pause
exit /b 1

:eof
echo Press any key to exit...
pause >nul
