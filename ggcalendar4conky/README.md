# ggcalendar for Conky

# Installation:
 - Install Conky depend on your desktop dist
 - put sync.sh to the same folder of ggcalendar.go/quickstart.go or ggcalendar executable
 - edit sync.sh to desinate ggcalendar executable path
 - edit sync.sh and change -p or -c flag to fit your need,view sync.sh for more detail
 - edit your .conkyrc(refer to my example in the same folder,default installed path should be $HOME) to display ggcalendar4conky.txt generated in previous step
 - variable xftfont in .conkyrc must be mono space
 - execute ggcalendar in the first time to get permission from your google account
 
# usage
 - start conky as daemon
```sh
$ conky &
```
 - execute sync.sh to fetch calendar event and update to conky.txt
 - or use linux schedualer tool like crontab to execute sync.sh to update google calendar data automatically
