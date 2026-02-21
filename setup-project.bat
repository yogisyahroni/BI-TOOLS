@echo off
echo Setting up project...

REM Buat folder .agent\skills
mkdir .agent\skills 2>nul

REM Copy skill dari global
xcopy /E /I /Y "%USERPROFILE%\.gemini\antigravity\skills\*" ".agent\skills\"

echo.
echo ✅ Done! Skills copied to .agent\skills
pause