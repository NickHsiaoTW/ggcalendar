package main

import (
	"database/sql"
  "flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"strconv"
	"time"
	"unicode/utf16"

	"golang.org/x/text/width"
	"google.golang.org/api/calendar/v3"
	
	_ "github.com/mattn/go-sqlite3"
)

var CELL_WIDTH = 10

var CALENDAR_WIDTH = CELL_WIDTH*7+8 

var Today = time.Now()

const (
        white = iota
        yellow = iota
        red = iota
        cyan = iota
        magenta = iota
)

const helloday int = magenta
const allday int = cyan
const normal int = cyan
const now int = red

func write2UTF8(name string,data string) error {
  file, err := os.Create(name)
  if err != nil {
    return err
  }
  _,err = file.WriteString(data)
  return err
}

func write2UTF16(name string,data string) error {
  //ref : https://gist.github.com/Bablzz/acfec39aceb84ee3a8a9614a50c87eac
  var bytes [2]byte
  const BOM = '\ufffe' //LE. for BE '\ufeff' 


  file, err := os.Create(name)
  if err != nil {
    fmt.Errorf("Can't open file. %v", err)
    return err
  }
  defer file.Close()
	
  bytes[0] = BOM >> 8
  bytes[1] = BOM & 255

  file.Write(bytes[0:])
  runes := utf16.Encode([]rune(data))
  for _, r := range runes {
    bytes[1] = byte(r >> 8)
    bytes[0] = byte(r & 255)
    file.Write(bytes[0:])
  }
  return nil
}

func count_half(s string) int {
	size := 0
	for _, runeValue := range s {
		p := width.LookupRune(runeValue)
		if p.Kind() == width.EastAsianWide {
			size += 2
			continue
		}
		if p.Kind() == width.EastAsianNarrow {
			size += 1
			continue
		}

    if runeValue == 10 {
			continue
    }
		panic("cannot determine!")
	}
	return size
}


func list_calendars(srv *calendar.Service) {
	list, dataok := get_calendars(srv)
	if !dataok {
      fmt.Println("Unable to get calendars")
      return  
	}
	for _, i := range list {
		fmt.Println(i.Summary+",Id:"+i.Id)
	}  
}


type Event struct {
	Start	time.Time
	Start_s	string
	Title		string
	Type		int
}

func padding_string(orig string, count int) string{  
  if count == 0 {
    return ""
  }
  padding_s := ""
  for i := 1; i <= count; i++ {
    padding_s = padding_s + orig
  }
  return padding_s
}

func split_string(s string, c int) []string {

	w := 0
	var splittes_s []string
	single_line := ""
	for _, runeValue := range s {
	  char_w := 0
		p := width.LookupRune(runeValue)
		if p.Kind() == width.EastAsianWide {
		  char_w = 2
		} else if p.Kind() == width.EastAsianNarrow {
		  char_w = 1
		}
		if w + char_w <= c {
		  single_line = single_line + string(runeValue)
		  w = w + char_w
		} else {
		  single_line = single_line + padding_string(" ", c - w)
		  splittes_s = append(splittes_s, single_line)
		  //reset single_line
		  single_line = string(runeValue)
		  w = char_w
		}		
	}
	single_line = single_line + padding_string(" ", c - w)
  splittes_s = append(splittes_s, single_line)
	return splittes_s
}


func append_type(string_sl []string, event_type int) []string{
  for i := 0; i < len(string_sl); i++ {
    string_sl[i] = string_sl[i] + fmt.Sprintf(":%d",event_type)
  }
  return string_sl
}

func fill_cell(time_key string, events []Event) []string{

    sort.SliceStable(events, func(i, j int) bool {
        return events[i].Start_s < events[j].Start_s
    })

    var filled_s []string
    
    for i := 0; i < len(events); i++ {
      var new_added []string
      event := events[i]
      
      if (event.Start_s != "0:000" && event.Start_s != "0:001" && event.Start_s != "") {
        if count_half(event.Start_s+" "+event.Title) <= CELL_WIDTH {
          new_added = split_string(event.Start_s+" "+event.Title,CELL_WIDTH)
        } else {
          new_added = split_string(event.Start_s+padding_string(" ",CELL_WIDTH-count_half(event.Start_s))+event.Title,CELL_WIDTH)
        }
      } else {
        new_added = split_string(event.Title,CELL_WIDTH)
      }

      new_added = append_type(new_added,event.Type)
      filled_s = append(filled_s, new_added...) 
      if event.Type == now {
        continue
      }     
      if i < len(events)-1 {
        if events[i+1].Type == now {
          continue
        }
      }
      filled_s = append(filled_s, padding_string(" ",CELL_WIDTH)+fmt.Sprintf(":%d",normal))
    }
    return filled_s
}

func get_and_remove_first(slice *[]string) string{
  if len(*slice) == 0 {
    return ""
  }
  r := (*slice)[0]
  *slice = (*slice)[1:]
  return r
}

func draw_text(slice_s *[]string, color int, draw_t string, padding int) {
    for i := 0; i < len(*slice_s); i++ {
      if i == color {
        (*slice_s)[i] = (*slice_s)[i] + draw_t + padding_string(" ",padding - count_half(draw_t))
      } else {
        (*slice_s)[i] = (*slice_s)[i] + padding_string(" ",padding)
      }
    }
}

func draw_day(slice_s *[]string, day_t time.Time,day_s string,today_s string) {
     if day_s == today_s {
       draw_text(slice_s,red,fmt.Sprintf("%02d **",day_t.Day()),CELL_WIDTH)
     } else {
       draw_text(slice_s,yellow,fmt.Sprintf("%02d",day_t.Day()),CELL_WIDTH)
     }  
}

func padding_space(slice_s *[]string,items []int,count int) {
    spaces := padding_string(" ",count)
    for _,i := range items {
      (*slice_s)[i] = (*slice_s)[i] + spaces
    }
}

func padding_newline(slice_s *[]string) {
    for i := 0; i < len(*slice_s); i++ {
      (*slice_s)[i] = (*slice_s)[i] + "\n"
    }
}

func draw_1cell_1row(event_s []string,slice_s *[]string) []string{
         event_s_s := get_and_remove_first(&event_s)
         if (event_s_s == "" || count_half(event_s_s) < CELL_WIDTH+2){
           draw_text(slice_s,white," ",CELL_WIDTH)
         } else {
           color_code, err := strconv.Atoi(string(event_s_s[len(event_s_s)-1]))
           if err != nil {
             draw_text(slice_s,white," ",CELL_WIDTH)
           } else {
             event_s_s = event_s_s[:len(event_s_s)-2]
             draw_text(slice_s,color_code,event_s_s,CELL_WIDTH)
           }
         }
         return event_s
}         
         
func draw_horizontal_border_and_newline(slice_s *[]string) {
  //(*slice_s)[white] += "+----------+----------+----------+----------+----------+----------+----------+"
  text2draw := "+"
  for i:=0;i<7;i++ {
    text2draw = text2draw + padding_string("-",CELL_WIDTH) + "+"
  }
  draw_text(slice_s, white, text2draw, CALENDAR_WIDTH)
  //padding_space(slice_s,[]int{red,yellow,cyan,magenta},CALENDAR_WIDTH)
  padding_newline(slice_s)
}

func get_width(runeValue rune) int{
    p := width.LookupRune(runeValue)
		if p.Kind() == width.EastAsianWide {
			return 2
		}
		if p.Kind() == width.EastAsianNarrow {
			return 1
		}
		return 0
}

/*
func print_mono_color(drawable []string) {
 
  color_count := len(drawable)
  var drawable_rune [][]rune
  var rune_count []int
  
  for i:=0;i<color_count;i++{
    rune_single_color := []rune(drawable[i])
    drawable_rune = append(drawable_rune,rune_single_color)
    rune_count = append(rune_count,len(rune_single_color))
  }
  index := make([]int, color_count)
  mono_line := ""
  for true {
    var new_added rune
    var new_added_width int
    
    new_added = drawable_rune[white][index[white]]
    new_added_width = get_width(new_added)
    if (new_added != 32 && new_added_width > 0){
       index[white] += 
    }
  }
}
*/

func sync_desktopcal(srv *calendar.Service,calendar_ids string,dbpath string) {
  var calendars []string
  
  if calendar_ids == "all" {
  	list, dataok := get_calendars(srv)
	  if !dataok {
      fmt.Println("Unable to get calendars")
      return  
	  }
	  for _, i := range list {
		  calendars = append(calendars,i.Id)
	  }  
	} else if calendar_ids == "primary"{
		calendars = append(calendars,"primary")
	} else {
	  calendars = strings.Split(calendar_ids, ";")
	}


  event_map := make(map[string]string)

	var event_keys []string
	
	target := -1
	for index,value := range calendars {
	  if strings.Contains(value, "holiday@group.v.calendar.google.com") {
	    target = index
	    break
	  }
	}
	//put holiday calendar to the last so it could be written to top of event of day
	if (target >= 0 && target < len(calendars)) {
	  part3 := calendars[target]
	  part1 := calendars[:target]
	  part2 := calendars[target+1:]
	  
	  calendars = append(part1,part2...)
	  calendars = append(calendars,part3)
	}
	
	for _, c_id := range calendars {
	  //is_helloday := false
	  //if strings.Contains(c_id, "holiday@group.v.calendar.google.com") {
	    //is_helloday = true
	  //}
	  fmt.Printf("Retriving events from calendar:%s\n",c_id)
		items, dataok := get_events(srv, c_id, 1, 1) 
	  if !dataok {
	      fmt.Println("Unable to get events")
	      continue
	  }
	  fmt.Printf("Got %d items\n",len(items))
	  if len(items) > 0 {
	    for _, i := range items {
	        if i.Start.DateTime != "" {
	            //fmt.Println(i.Summary+","+i.Start.DateTime)
	            event_time, _ := time.Parse(time.RFC3339,i.Start.DateTime)
	            
	            it_unique_id := fmt.Sprintf("dkcal_mdays_%d%02d%02d", event_time.Year(), event_time.Month(), event_time.Day())
	            it_time := fmt.Sprintf("%02d:%02d", event_time.Hour(), event_time.Minute())
	            
	            if event_map[it_unique_id] == "" {
	              event_map[it_unique_id] = it_time + "\n" + i.Summary
	            } else {
	              event_map[it_unique_id] = event_map[it_unique_id] + "\n" + it_time + "\n" + i.Summary
	            }
	        } else {
	            //fmt.Println(i.Summary+","+i.Start.Date)
	            event_time, _ := time.Parse("2006-01-02",i.Start.Date)

	            it_unique_id := fmt.Sprintf("dkcal_mdays_%d%02d%02d", event_time.Year(), event_time.Month(), event_time.Day())
	            /*
	            if is_helloday {
	            } else {
	            }
	            */
	            if event_map[it_unique_id] == "" {
	              event_map[it_unique_id] = i.Summary
	            } else {
	              event_map[it_unique_id] = i.Summary + "\n" + event_map[it_unique_id] 
	            }
	        }
	    }
	  }
	} 
	
	for k := range event_map {
		event_keys = append(event_keys, k)
	}
	sort.Strings(event_keys)
	
	//load db
	var dbkeys []string
	dbmap := make(map[string]string)
	db, err := sql.Open("sqlite3", dbpath)
	if err != nil {
	  fmt.Println(err)
	  fmt.Println("can not open sql file:"+dbpath)
		fmt.Println("ERROR")
		return
	}

	rows, err := db.Query("SELECT it_unique_id,it_content FROM item_table ORDER BY it_unique_id ASC")
	if err != nil {
	  fmt.Println(err)
		fmt.Println("query item_table fail")
		fmt.Println("ERROR")		
		return
	}

	for rows.Next() {
		var it_content string
		var it_unique_id string
		err = rows.Scan(&it_unique_id, &it_content)
		dbkeys = append(dbkeys, it_unique_id)
		dbmap[it_unique_id] = it_content
	}
	sort.Strings(dbkeys)

	//if all empty....
	if len(event_keys) == 0 && len(dbkeys) == 0 {
		fmt.Println("NOCHANGE")
		return 
	}

	//if db map size = 0 ,just insert all ics map
	if len(dbkeys) == 0 && len(event_keys) > 0 {
		for k, v := range event_map {
			stmt, err := db.Prepare("INSERT INTO item_table(it_unique_id,it_content) values(?,?)")
			if err != nil {
				fmt.Println("db prepare inser fail")
				fmt.Println("ERROR")	
				return
			}
			_, err = stmt.Exec(k, v)
			if err != nil {
				fmt.Println("db do insert fail")
				fmt.Println("ERROR")	
				return
			}
		}
		fmt.Println("NEEDRESTART")	
		return
	}

	//if ics map size = 0 ,just delete all in item_table
	if len(event_keys) == 0 && len(dbkeys) > 0 {
		stmt, err := db.Prepare("delete from item_table")
		if err != nil {
			fmt.Println("db prepare delete fail")
			fmt.Println("ERROR")	
			return
		}
		_, err = stmt.Exec()
		if err != nil {
			fmt.Println("db do delete fail")
			fmt.Println("ERROR")	
			return
		}
		fmt.Println("NEEDRESTART")
		return
	}

	//start compare one by one
	iI := 0
	dI := 0
	need_reload := false
	fmt.Printf("ics len %d\n",len(event_keys))
	fmt.Printf("db len %d\n",len(dbkeys))
	for true {
    fmt.Printf("iI:%d,dI:%d\n",iI,dI)
		if (iI >= len(event_keys)) && (dI >= len(dbkeys)) {
			//all job done,get out
			break
		}

    icskey := "dkcal_mdays_99999999"
		if iI < len(event_keys) {
			icskey = event_keys[iI]
		}
			
    dbkey := "dkcal_mdays_99999999"
		if dI < len(dbkeys) {
			dbkey = dbkeys[dI]
		}
		
		//delete case
		if icskey > dbkey {
			//do delete
			stmt, err := db.Prepare("delete from item_table where it_unique_id=?")
			if err != nil {
				fmt.Println("db repare delete fail")
				fmt.Println("ERROR")	
				return
			}
			_, err = stmt.Exec(dbkey)
			if err != nil {
				fmt.Println("db do delete fail")
				fmt.Println("ERROR")	
				return
			}
      need_reload = true
			dI = dI + 1
			continue
		}

		//insert case
		if dbkey > icskey {
			//do insert
			stmt, err := db.Prepare("INSERT INTO item_table(it_unique_id,it_content) values(?,?)")
			if err != nil {
				fmt.Println("db repare insert fail")
				fmt.Println("ERROR")	
				return
			}
			_, err = stmt.Exec(icskey, event_map[icskey])
			if err != nil {
				fmt.Println("db do insert fail")
				fmt.Println("ERROR")	
				return
			}
			need_reload = true
			iI = iI + 1
			continue
		}

		//update case
		if dbkey == icskey {
			//do update
			if dbmap[dbkey] != event_map[icskey] {
			  fmt.Printf("Update %s to %s\n",icskey,event_map[icskey])
				stmt, err := db.Prepare("update item_table set it_content=? where it_unique_id=?")
				if err != nil {
					fmt.Println("db repare update fail")
					fmt.Println("ERROR")	
					return
				}
				_, err = stmt.Exec(event_map[icskey], icskey)
				if err != nil {
					fmt.Println("db do update fail")
					fmt.Println("ERROR")	
					return
				}
				need_reload = true
			}
			dI = dI + 1
			iI = iI + 1
		}
	}
	if need_reload {
	    fmt.Println("NEEDRESTART")	
	    return
	}
	fmt.Println("NOCHANGE")
}

func draw_gcalcli(srv *calendar.Service,calendar_ids string,path string, file_encoding string) {
  var calendars []string
  
  if calendar_ids == "all" {
  	list, dataok := get_calendars(srv)
	  if !dataok {
      fmt.Println("Unable to get calendars")
      return  
	  }
	  for _, i := range list {
		  calendars = append(calendars,i.Id)
	  }  
	} else if calendar_ids == "primary"{
		calendars = append(calendars,"primary")
	} else {
	  calendars = strings.Split(calendar_ids, ";")
	}


  event_map := make(map[string][]Event)
  event_map_print := make(map[string][]string)
	
	for _, c_id := range calendars {
	  is_helloday := false
	  if strings.Contains(c_id, "holiday@group.v.calendar.google.com") {
	    is_helloday = true
	  }
	  fmt.Printf("Retriving events from calendar:%s\n",c_id)
		items, dataok := get_events(srv, c_id, 1, 1) 
	  if !dataok {
	      fmt.Println("Unable to get events")
	      continue
	  }
	  fmt.Printf("Got %d items\n",len(items))
	  if len(items) > 0 {
	    for _, i := range items {
	        if i.Start.DateTime != "" {
	            //fmt.Println(i.Summary+","+i.Start.DateTime)
	            event_time, _ := time.Parse(time.RFC3339,i.Start.DateTime)
	            var event Event
	            event.Start = event_time
	            event.Title = i.Summary
	            event.Type = normal
	            event.Start_s = fmt.Sprintf("%d:%02d",event_time.Hour(),event_time.Minute())
	            event_key := fmt.Sprintf("%d-%02d-%02d",event_time.Year(),event_time.Month(),event_time.Day())
	            event_map[event_key] = append(event_map[event_key],event)
	        } else {
	            //fmt.Println(i.Summary+","+i.Start.Date)
	            event_time, _ := time.Parse("2006-01-02",i.Start.Date)
	            var event Event
	            event.Start = event_time
	            event.Title = i.Summary
	            if is_helloday {
	              event.Type = helloday
	              event.Start_s = "0:000"
	            } else {
	              event.Type = allday
	              event.Start_s = "0:001"
	            }
	            event_key := fmt.Sprintf("%d-%02d-%02d",event_time.Year(),event_time.Month(),event_time.Day())
	            event_map[event_key] = append(event_map[event_key],event)
	        }
	    }
	  }
	} 

  //final ,add red ------ as an event
  fmt.Println("Add now event")
  var event Event
  event.Start = Today
  event.Title = padding_string("-",CELL_WIDTH)
  event.Start_s = ""
  event.Type = now
  event_key := fmt.Sprintf("%d-%02d-%02d",Today.Year(),Today.Month(),Today.Day())
  event_map[event_key] = append(event_map[event_key],event)
     

     
  for k,v := range event_map {
      event_map_print[k] = fill_cell(k,v)
  }
  
  //draw calendar from here
  first_day_month := Today.AddDate(0,0,1-Today.Day())
  first_day_draw := first_day_month.AddDate(0, 0, -int(first_day_month.Weekday()))
  
  all_line := make([]string, 5)
  
  
  //draw firt row
  fmt.Println("Draw first row")
  
  text2draw := "+"+padding_string("-",CALENDAR_WIDTH-2)+"+"
  draw_text(&all_line,white,text2draw,CALENDAR_WIDTH)
       
  padding_newline(&all_line)
  
  //draw year month row
  fmt.Println("Draw year month row")
  year_month := fmt.Sprintf("%d %s",Today.Year(),Today.Month())
  year_month = " "+year_month+padding_string(" ",CALENDAR_WIDTH-1-count_half(year_month))
  all_line[yellow] = all_line[yellow] + year_month
  all_line[white] += "|"+padding_string(" ",CALENDAR_WIDTH-2)+"|"
  padding_space(&all_line,[]int{red,cyan,magenta},CALENDAR_WIDTH)
  padding_newline(&all_line)
  
  draw_horizontal_border_and_newline(&all_line)
  
  
  //draw week days
  fmt.Println("Draw week days")
  all_line[white] += "|          |          |          |          |          |          |          |"
  all_line[yellow] += " "+"Sunday"+padding_string(" ",CELL_WIDTH-len("Sunday"))
  all_line[yellow] += " "+"Monday"+padding_string(" ",CELL_WIDTH-len("Monday"))
  all_line[yellow] += " "+"Tuesday"+padding_string(" ",CELL_WIDTH-len("Tuesday"))  
  all_line[yellow] += " "+"Wednesday"+padding_string(" ",CELL_WIDTH-len("Wednesday")) 
  all_line[yellow] += " "+"Thursday"+padding_string(" ",CELL_WIDTH-len("Thursday")) 
  all_line[yellow] += " "+"Friday"+padding_string(" ",CELL_WIDTH-len("Friday")) 
  all_line[yellow] += " "+"Saturday"+padding_string(" ",CELL_WIDTH-len("Saturday")) + " "

  padding_space(&all_line,[]int{red,cyan,magenta},CALENDAR_WIDTH)
  padding_newline(&all_line)

  draw_horizontal_border_and_newline(&all_line)
  
  //start draw event 
  fmt.Println("Draw day events starts")
  today_s := fmt.Sprintf("%d-%02d-%02d",Today.Year(),Today.Month(),Today.Day())
  sun_t := first_day_draw
  for i := 0; i < 5; i++ {
     sun_s := fmt.Sprintf("%d-%02d-%02d",sun_t.Year(),sun_t.Month(),sun_t.Day())
     mon_t := sun_t.AddDate(0,0,1)
     mon_s := fmt.Sprintf("%d-%02d-%02d",mon_t.Year(),mon_t.Month(),mon_t.Day())
     tue_t := sun_t.AddDate(0,0,2)
     tue_s := fmt.Sprintf("%d-%02d-%02d",tue_t.Year(),tue_t.Month(),tue_t.Day())
     wed_t := sun_t.AddDate(0,0,3)
     wed_s := fmt.Sprintf("%d-%02d-%02d",wed_t.Year(),wed_t.Month(),wed_t.Day())
     thu_t := sun_t.AddDate(0,0,4)
     thu_s := fmt.Sprintf("%d-%02d-%02d",thu_t.Year(),thu_t.Month(),thu_t.Day())
     fri_t := sun_t.AddDate(0,0,5)
     fri_s := fmt.Sprintf("%d-%02d-%02d",fri_t.Year(),fri_t.Month(),fri_t.Day())
     sat_t := sun_t.AddDate(0,0,6)
     sat_s := fmt.Sprintf("%d-%02d-%02d",sat_t.Year(),sat_t.Month(),sat_t.Day())
     
     draw_text(&all_line,white,"|",1)
     draw_day(&all_line,sun_t, sun_s, today_s) 
     draw_text(&all_line,white,"|",1)
     draw_day(&all_line,mon_t, mon_s, today_s)   
     draw_text(&all_line,white,"|",1)
     draw_day(&all_line,tue_t, tue_s, today_s)  
     draw_text(&all_line,white,"|",1)
     draw_day(&all_line,wed_t, wed_s, today_s)  
     draw_text(&all_line,white,"|",1)
     draw_day(&all_line,thu_t, thu_s, today_s)      
     draw_text(&all_line,white,"|",1)
     draw_day(&all_line,fri_t, fri_s, today_s)  
     draw_text(&all_line,white,"|",1)
     draw_day(&all_line,sat_t, sat_s, today_s)
     draw_text(&all_line,white,"|",1)
     padding_newline(&all_line)
     
     finish := false
     for !finish {
         end_loop := true
         
         draw_text(&all_line,white,"|",1)
         event_map_print[sun_s] = draw_1cell_1row(event_map_print[sun_s],&all_line)
         if len(event_map_print[sun_s]) > 0 {
           end_loop = false
         }
         
         draw_text(&all_line,white,"|",1)
         event_map_print[mon_s] = draw_1cell_1row(event_map_print[mon_s],&all_line)
         if len(event_map_print[mon_s]) > 0 {
           end_loop = false
         }       
         
         draw_text(&all_line,white,"|",1)
         event_map_print[tue_s] = draw_1cell_1row(event_map_print[tue_s],&all_line)
         if len(event_map_print[tue_s]) > 0 {
           end_loop = false
         }  

         draw_text(&all_line,white,"|",1)
         event_map_print[wed_s] = draw_1cell_1row(event_map_print[wed_s],&all_line)
         if len(event_map_print[wed_s]) > 0 {
           end_loop = false
         }  

         draw_text(&all_line,white,"|",1)
         event_map_print[thu_s] = draw_1cell_1row(event_map_print[thu_s],&all_line)
         if len(event_map_print[thu_s]) > 0 {
           end_loop = false
         }  

         draw_text(&all_line,white,"|",1)
         event_map_print[fri_s] = draw_1cell_1row(event_map_print[fri_s],&all_line)
         if len(event_map_print[fri_s]) > 0 {
           end_loop = false
         } 

         draw_text(&all_line,white,"|",1)
         event_map_print[sat_s] = draw_1cell_1row(event_map_print[sat_s],&all_line)
         if len(event_map_print[sat_s]) > 0 {
           end_loop = false
         } 
         
         draw_text(&all_line,white,"|",1)
         padding_newline(&all_line)
         
         finish = end_loop                    
     }//end of for !finish
     
     
             
     //end of single loop
     sun_t = sun_t.AddDate(0,0,7)
     if i < 4 {
       draw_horizontal_border_and_newline(&all_line)  
     } else {
       text2draw := "+"+padding_string("-",CALENDAR_WIDTH-2)+"+"
       draw_text(&all_line,white,text2draw,CALENDAR_WIDTH)
       padding_newline(&all_line)
     }
  }//end of for i := 0; i < 5; i++
  
  //print_mono_color(all_line)
  
  if file_encoding == "utf-16" {
    fmt.Println("Writing to UTF-16 text file")
    write2UTF16(path+"white_line.txt",all_line[white])
    write2UTF16(path+"yellow_line.txt",all_line[yellow]) 
    write2UTF16(path+"cyan_line.txt",all_line[cyan])
    write2UTF16(path+"red_line.txt",all_line[red])
    write2UTF16(path+"magenta_line.txt",all_line[magenta])
  } else if file_encoding == "utf-8" {
    write2UTF8(path+"white_line.txt",all_line[white])
    write2UTF8(path+"yellow_line.txt",all_line[yellow]) 
    write2UTF8(path+"cyan_line.txt",all_line[cyan])
    write2UTF8(path+"red_line.txt",all_line[red])
    write2UTF8(path+"magenta_line.txt",all_line[magenta])
  } else {
    fmt.Println("Unknown encoding:"+file_encoding)
  }
  /*
	for k, v := range event_map_print {
	  fmt.Println(k)
	  fmt.Println("==========")
	  for _, line := range v {
	    fmt.Println(line)
	  }
	  fmt.Println("==========")
	}
  */
}

func list_events(srv *calendar.Service, calendar_ids string, backward int, forward int) {
  var calendars []string
  
  if calendar_ids == "all" {
  	list, dataok := get_calendars(srv)
	  if !dataok {
      fmt.Println("Unable to get calendars")
      return  
	  }
	  for _, i := range list {
		  calendars = append(calendars,i.Id)
	  }  
	} else if calendar_ids == "primary"{
		calendars = append(calendars,"primary")
	} else {
	  calendars = strings.Split(calendar_ids, ";")
	}
	
	//fmt.Println(calendars)
	
	for _, c_id := range calendars {
    items, dataok := get_events(srv, c_id, backward, forward) 
    if !dataok {
      fmt.Println("Unable to get events for calendar:"+c_id)
      continue
    }
    if len(items) > 0 {
      for _, i := range items {
        if i.Start.DateTime != "" {
            fmt.Println(i.Summary+","+i.Start.DateTime)
        } else {
            fmt.Println(i.Summary+","+i.Start.Date)
        }
      }
    }
  }  
}

func get_calendars(srv *calendar.Service) ([]*calendar.CalendarListEntry ,bool){
	list, err := srv.CalendarList.List().Do()
	if err != nil {
		fmt.Println("Error getting calendars")
		return nil,false
	}
	return list.Items,true  
}

func get_events(srv *calendar.Service, calendar_id string,backward int,forward int) ([]*calendar.Event, bool) {
  calendarEvents, err := srv.Events.List(calendar_id).TimeMin(time.Now().AddDate(0,-backward,0).Format(time.RFC3339)).TimeMax(time.Now().AddDate(0,backward,0).Format(time.RFC3339)).SingleEvents(true).OrderBy("startTime").Do()
  if err != nil {
    fmt.Println(err)
    return nil, false
  }

  return calendarEvents.Items, true
}

func main() {

  srv, err := getService()
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
		return
	}
	
	//finish authentication
	
	function := flag.String("func", "list_calendars", "main function list_calendars/list_events/draw_gcalcli")
	file_encoding := flag.String("e", "utf-8", "utf-8 or utf-16")
	write_path := flag.String("p", "./", "path to widget location,must be C:/xxx/xxx format")
	db_path := flag.String("d", "./calendar.db", "path to db ex:'C:/Users/xxxx/AppData/Roaming/DesktopCal/Db/calendar.db'") 
	forward_m := flag.Int("f", 1, "get event months after now")
	backward_m := flag.Int("b", 1, "get event months before now")
	calendars2load := flag.String("c", "all", "calendars to load seperate by ';' ex:'aaa@gmail.com;bbb@gmail.com' or just 'primary'")
	flag.Parse()
	fmt.Printf("Flags function:%s, forward_m:%d, backward_m:%d, calendars to load:%s\n",*function,*forward_m,*backward_m,*calendars2load)
	
	if(*file_encoding != "utf-8" && *file_encoding != "utf-16") {
	    *file_encoding = "utf-8"
	}
	fmt.Printf("Flags DB path:%s, Write path:%s, Encoding:%s\n\n",*db_path,*write_path,*file_encoding)	
	
	switch { 
    case *function == "list_calendars": 
        list_calendars(srv)
        return
    case *function == "list_events": 
        list_events(srv, *calendars2load, *backward_m, *forward_m)
        return
    case *function == "draw_gcalcli": 
        draw_gcalcli(srv, *calendars2load, *write_path, *file_encoding)     
        return
    case *function == "sync_desktopcal": 
        sync_desktopcal(srv, *calendars2load, *db_path)     
        return   
    default:
        fmt.Printf("Unknown function:%s,must be one of list_calendars/list_events/draw_gcalcli\n",*function) 
	} 
}

