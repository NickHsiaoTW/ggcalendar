#!/bin/bash

#./ggcalendar or ${path to ggcalendar executable file} or go run ggcalendar.go quickstart.go -func draw_conky -c "aaa@bbb.com;ccc@ddd.com"
#-c : calendars to load ,please use 'ggcalendar.exe -func list_calendars' to list ids to be shown concated by ';',ex : 'aaa@bbb.com;ccc@ddd.com'
#generated file ggcalendar4conky.txt default location is the same folder with ggcalendar executable or -p to designate path for ggcalendar4conky.txt

./ggcalendar -func draw_conky -c "aaa@bbb.com;ccc@ddd.com"


