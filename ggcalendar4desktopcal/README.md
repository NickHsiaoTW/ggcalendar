# ggcalendar for Desktopcal

# Installation:
 - go to <https://www.desktopcal.com/> download and install Desktopcal
 - put sync.cmd to the same folder of ggcalendar.go/quickstart.go or ggcalendar.exe
 - edit sync.cmd and change -d and -c flag to fit your need,view sync.cmd for more detail
 - make sure C:\Users\%username%\AppData\Roaming\DesktopCal\desktopcal.exe exists if not,you need to edit sync.cmd line:7 or Desktopcal would not refresh
 - execute ggcalendar.exe in the first time to get permission from your google account
 
# usage
 - execute sync.cmd with "Run as administrator" to fetch latest google calendar data to update calendar database and possible need to kill and start Desktopcal to refresh UI,that's why it needs admin privilege
 - or use windows schedualer <https://www.windowscentral.com/how-create-automated-task-using-task-scheduler-windows-10> to execute sync.cmd to update google calendar data automatically(with admin privilege,IMPORTANT!)