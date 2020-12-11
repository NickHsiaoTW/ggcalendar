@echo off
set "synccmd=ggcalendar.exe -func sync_desktopcal -c all -d C:/Users/%username%/AppData/Roaming/DesktopCal/Db/calendar.db"
REM or go.exe run .\ggcalendar.go .\quickstart.go -func sync_desktopcal -c all -d C:/Users/%username%/AppData/Roaming/DesktopCal/Db/calendar.db
FOR /F "tokens=*" %%g IN ('%synccmd%') do (SET VAR=%%g)
echo %VAR%
if "%VAR%"=="NEEDRESTART" (
    echo "To kill and start desktopcal.exe"
    taskkill /f /im desktopcal.exe
    start C:\Users\%username%\AppData\Roaming\DesktopCal\desktopcal.exe 
) else (
    echo "No need to restart"
)

REM -c : calendars to load ,please use 'ggcalendar.exe -func list_calendars' to list ids wanted to be shown which concated by ';',ex : 'aaa@bbb.com;ccc@ddd.com'
REM -d : Desktopcal sql file path