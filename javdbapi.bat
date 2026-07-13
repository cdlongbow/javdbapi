@echo off
title javdbapi

:menu
cls
echo ========================================
echo        javdbapi - JAVDB Scraper
echo ========================================
echo.
echo [1] Video Detail
echo [2] Search
echo [3] Ranking
echo [4] Actor
echo [0] Exit
echo.
echo Output: JSON + NFO files saved in same folder
echo.
set /p c=Choice:

if "%c%"=="0" exit /b
if "%c%"=="1" call :video
if "%c%"=="2" call :search
if "%c%"=="3" call :ranking
if "%c%"=="4" call :actor
goto menu

:video
set /p v=VideoID (e.g. ZNdEbV):
javdbapi.exe video --id %v% --output both --stale-after 0s
echo.
echo Files saved in current folder
pause
goto :eof

:search
set /p k=Keyword:
javdbapi.exe search --keyword "%k%" --output both --stale-after 0s --max-pages 1
echo.
echo Files saved in current folder
pause
goto :eof

:ranking
set /p p=Period (daily/weekly/monthly, default weekly):
if "%p%"=="" set p=weekly
set /p t=Type (censored/uncensored/western, default censored):
if "%t%"=="" set t=censored
javdbapi.exe ranking --period %p% --type %t% --output both --stale-after 0s
echo.
echo Files saved in current folder
pause
goto :eof

:actor
set /p a=ActorID (e.g. neRNX):
set /p f=Filter (all/cnsub/download/playable, default all):
if "%f%"=="" set f=all
javdbapi.exe actor --id %a% --filter %f% --output both --stale-after 0s
echo.
echo Files saved in current folder
pause
goto :eof
