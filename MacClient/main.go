package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/getlantern/systray"
	"github.com/getlantern/systray/example/icon"
	"github.com/grandcat/zeroconf"
	_ "github.com/mattn/go-sqlite3"
	"github.com/skratchdot/open-golang/open"
	"howett.net/plist"
)

var (
	DB_PATH     = ""
	ProgramPath = ""
)

var Global = &Config{}

type Config struct {
	API_HOST    string `json:"api_host"`
	IncludeBody bool   `json:"include_body"`
	Debug       bool   `json:"debug"`
}

func Log(text string, args ...interface{}) {
	file, err := os.OpenFile(ProgramPath+"/ledofication.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(file)
	log.Printf(text, args...)
	if Global.Debug {
		fmt.Printf(text, args...)
	}
}

func (c *Config) Load(filePath string) {
	configuration, err := ioutil.ReadFile(filePath)
	if err != nil {
		Log("Couldn't read application configuration file\n")
	}

	err = json.Unmarshal(configuration, &c)
	if err != nil {
		Log("Couldn't unmarshal configuration\n")
	}
}

func DetectProgramPath() {
	Pathas, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}
	ProgramPath = Pathas
}

func DetectZeroConf() {
	Log("LED Matrix device is not defined in the configuration file, try to autodetect it...\n")
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		Log("Failed to initialize ZeroConf resolver: %s\n", err.Error())
	}

	entries := make(chan *zeroconf.ServiceEntry)
	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {
			Log("%s\n", entry)
			Log("Found device: %s:%d\n", entry.AddrIPv4, entry.Port)
			Global.API_HOST = fmt.Sprintf("http://%s:%d", entry.AddrIPv4, entry.Port)
		}
		Log("No more ZeroConf entries found. Finishing...\n")
	}(entries)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	err = resolver.Browse(ctx, "_leds._tcp", "local.", entries)
	if err != nil {
		Log("Failed to browse: %s\n", err.Error())
	}

	<-ctx.Done()
}

func main() {
	go func() {
		DetectProgramPath()
		Global.Load(ProgramPath + "/config.json")
		DetectDB()
		if Global.API_HOST == "" {
			DetectZeroConf()
		}
		if _, err := os.Stat(DB_PATH); os.IsNotExist(err) {
			Log("!!! CRITICAL ERROR !!! MacOS notification database file not found!\n")
			os.Exit(1)
		}
		if Global.Debug {
			Log("Database found: %s\n", DB_PATH)
		}
		DoMessage("Connected!")
		examine := flag.Bool("examine", false, "examine notification records")
		message := flag.String("text", "", "Send text to the LED from command line")
		flag.Parse()
		if *examine {
			ExamineDatabase()
			os.Exit(0)
		}
		if *message != "" {
			DoMessage(*message)
			os.Exit(0)
		}
		Listen()
	}()
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(icon.Data)
	systray.SetTooltip("Ledofication daemon")
	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")
	mSettings := systray.AddMenuItem("Settings", "Settings")
	mSend := systray.AddMenuItem("Send text", "Send text to led matrix")
	mOpenLog := systray.AddMenuItem("Open log", "Open program log file")

	// Sets the icon of a menu item. Only available on Mac and Windows.
	mQuit.SetIcon(icon.Data)
	for {
		select {
		case <-mSettings.ClickedCh:
			Log("Settings clicked\n")
			SettingsDlgShow()
		case <-mSend.ClickedCh:
			Log("Send text\n")
			DoMessage("Simple text")
		case <-mOpenLog.ClickedCh:
			OpenLog()
		case <-mQuit.ClickedCh:
			systray.Quit()
			return
		}

	}

}

func SettingsDlgShow() {
	Log("Not implemented yet\n")
}

func OpenLog() {
	open.Run(ProgramPath + "/ledofication.log")
}

func onExit() {
	// clean up here
	Log("Exit function called\n")
}

func DoMessage(text string) {
	Log("Texting: %s...\n", text)
	form := url.Values{}
	form.Add("text", text)
	request, _ := http.NewRequest(http.MethodPost, Global.API_HOST+"/api/send", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		Log("Error posting to remote api server: %s\n", err)
		return
	}
	var res map[string]interface{}
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		Log("Error reading data from remote api server: %s\n", err)
		return
	}

	err = json.Unmarshal(bytes, &res)
	if err != nil {
		Log("Error unmarshalling json data from remote api server: %s\n", err)
		return
	}

	errMsg, ok := res["error"]
	if ok {
		Log("Maybe something went wrong...: %s\n", errMsg)
		return
	}
	if res["sent"].(bool) {
		Log("Delivered successfully!\n")
	}
}

func Listen() {
	SQL_MAIN := "SELECT (SELECT identifier from app where app.app_id=record.app_id) as app, rec_id, data, presented, delivered_date FROM record where app not like '%system%' and app not like '%transmission%' order by rec_id"
	SQL_MAXID := "SELECT (SELECT identifier from app where app.app_id=record.app_id) as app, max(rec_id) from record where app not like '%system%' and app not like '%transmission%'"
	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?cache=shared", DB_PATH))
	var Maxid int
	var App string
	if err != nil {
		Log("!!! CRITICAL ERROR !!! Unable to open notification database: %s\n", err)
		os.Exit(1)
	}
	defer db.Close()
	err = db.QueryRow(SQL_MAXID).Scan(&App, &Maxid)
	if err != nil {
		Log("Unable query notification database: %s (Maybe its empty?)\n", err)
		os.Exit(1)
	}
	Log("Got max id: %d, listening for the new events\n", Maxid)
	for {
		rows, err := db.Query(SQL_MAIN)
		if err != nil {
			Log("Error executing sql query: %s\n", err)
			//os.Exit(1)
		}
		defer rows.Close()
		for rows.Next() {
			var id int
			var app string
			var data []byte
			var presented sql.NullBool
			var delivered_date sql.NullFloat64
			err = rows.Scan(&app, &id, &data, &presented, &delivered_date)
			if err != nil {
				Log("Unable to scan data for sql query: %s\n", err)
			}
			if id > Maxid {
				Maxid = id
				Log("New event with id: %d\n", id)
				SendNotification(id, app, data, presented)
			}
		}
		err = rows.Err()
		if err != nil {
			Log("Cannot process further! err: %s\n", err)
		}

		time.Sleep(2 * time.Second)
	}

}

func SendNotification(id int, app string, data []byte, prestented sql.NullBool) {
	type Msg struct {
		Req struct {
			Title    string `plist:"titl"`
			SubTitle string `plist:"subt"`
			Message  string `plist:"body"`
		} `plist:"req"`
	}
	var Itemas Msg
	buf := bytes.NewReader(data)
	decoder := plist.NewDecoder(buf)
	err := decoder.Decode(&Itemas)
	if err != nil {
		Log("Some error: %s\n", err)
	}
	Log("Message from: %s SubTitle: %s Body: %s\n", Itemas.Req.Title, Itemas.Req.SubTitle, Itemas.Req.Message)
	msg := fmt.Sprintf("%s", Itemas.Req.Title)
	if Global.IncludeBody {
		msg = fmt.Sprintf("%s: %s", Itemas.Req.Title, Itemas.Req.Message)
	}
	DoMessage(fmt.Sprintf("Message from %s", msg))
}

func CheckRecord(data []byte) {
	type Msg struct {
		Req struct {
			Title    string `plist:"titl"`
			SubTitle string `plist:"subt"`
			Message  string `plist:"body"`
		} `plist:"req"`
	}

	var Itemas Msg
	buf := bytes.NewReader(data)
	decoder := plist.NewDecoder(buf)
	err := decoder.Decode(&Itemas)
	if err != nil {
		Log("Some error: %s\n", err)
	}
	Log("Item: %v\n", Itemas)
}

func ResolveTime(date float64) string {
	Taimas := date + 978307200
	sec, dec := math.Modf(Taimas)
	aha := time.Unix(int64(sec), int64(dec*(1e9)))
	return fmt.Sprintf("%s", aha)
}

func ExamineDatabase() {
	// removed uuid after rec_id
	SQL := "SELECT (SELECT identifier from app where app.app_id=record.app_id) as app, rec_id, data, presented, delivered_date FROM record where app not like '%system%' and app not like '%transmission%' order by rec_id"
	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?cache=shared", DB_PATH))
	if err != nil {
		Log("Unable to open notification database: %s\n", err)
	}
	defer db.Close()

	rows, err := db.Query(SQL)
	if err != nil {
		Log("Error executing sql query: %s\n", err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var app string
		var data []byte
		var presented bool
		var delivered_date float64
		err = rows.Scan(&app, &id, &data, &presented, &delivered_date)
		if err != nil {
			Log("Unable to scan data for sql query: %s\n", err)
		}
		Log("Notification ------>\n")

		Log("App: %s\nID: %d\nPresented: %v\nDelivered time: %s\n", app, id, presented, ResolveTime(delivered_date))
		CheckRecord(data)
		Log("<----- Notification end\n")
	}
	err = rows.Err()
	if err != nil {
		Log("Cannot process further! err: %s\n", err)
	}

}

func DetectDB() {
	out, err := exec.Command("/usr/bin/getconf", "DARWIN_USER_DIR").Output()
	if err != nil {
		Log("Unable to determine command output: %s\n", err)
		os.Exit(1)
	}
	DB_PATH = TruncateMaterial(string(out)) + "/com.apple.notificationcenter/db2/db"
}

func TruncateMaterial(str string) string {
	str = strings.Replace(str, "\t", "", -1)
	str = strings.Replace(str, "\n", "", -1)
	return str
}
