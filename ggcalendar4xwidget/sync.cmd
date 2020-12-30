@echo off
ggcalendar.exe -func draw_gcalcli -c all -e utf-16 -p C:/Users/%username%/Documents/XWidget/Widgets/ggcalendar/
REM or go.exe run .\ggcalendar.go .\quickstart.go -func draw_gcalcli -c all -e utf-16 -p C:/Users/%username%/Documents/XWidget/Widgets/ggcalendar/

REM -c : calendars to load ,please use 'ggcalendar.exe -func list_calendars' to list ids to be shown concated by ';',ex : 'aaa@bbb.com;ccc@ddd.com'
REM -e : encoding to utf-8 or utf-16(XWidget doesnt decode utf-8 well)
