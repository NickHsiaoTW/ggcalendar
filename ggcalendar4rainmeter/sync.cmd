@echo off
ggcalendar.exe -func draw_gcalcli -c all -e utf-8 -p C:/Users/%username%/Documents/Rainmeter/Skins/ggcalendar/
REM or go.exe run .\ggcalendar.go .\quickstart.go -func draw_gcalcli -c all -e utf-8 -p C:/Users/%username%/Documents/Rainmeter/Skins/ggcalendar/

REM -c : calendars to load ,please use 'ggcalendar.exe -func list_calendars' to list ids wanted to be shown which concated by ';',ex : 'aaa@bbb.com;ccc@ddd.com'
REM -e : encoding to utf-8 or utf-16(Rainmeter doesnt decode utf-16 well)