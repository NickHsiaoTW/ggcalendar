# ggcalendar for XWidget

# Installation:
 - go to <xwidget.com> download and install XWidget
 - put sync.cmd to the same folder of ggcalendar.go/quickstart.go or ggcalendar.exe
 - import ggcalendar.xwp by click the file
 - edit sync.cmd and change -p and -c flag to fit your need,view sync.cmd for more detail
 - edit ggcalendar in XWidget edit mode to adjust every text box font/size etc... to fit your need(default font is Minglu)
 - make sure the font is monospace or calendar would be deformed
 - edit ggcalendar in XWidget to change refresh time if you want to(default is 5min = 'sleep(300000)' in "Code" session of edit mode)
 - execute ggcalendar.exe in the first time to get permission from your google account

# usage
 - execute sync.cmd to fetch latest google calendar events
 - or use windows schedualer <https://www.windowscentral.com/how-create-automated-task-using-task-scheduler-windows-10> to execute sync.cmd to update google calendar data automatically