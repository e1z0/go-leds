package main

import (
	"fmt"
	//"strings"

	//"log"
	"flag"
	"os"
	"time"
	//. "github.com/cyoung/rpi"
)

const (
	//MatrixLedBlocks = 8
	//MatrixPosition = 3 //RotateClockwiseInvert
	DEBUG = 1
)

var (
	APP_VERSION     = "0.3"
	DISPLAY_INUSE   = false
	SETUP_AVAILABLE = true
)

func DisplayClock_v2(timenow string, lasttime string) {
	mtx := NewMatrix(Global.LedsLength, Global.LedsRotation)
	err := mtx.Open(0, 0, 1)
	if err != nil {
		fmt.Printf("Unable to handle led matrix")
		return
	}
	for i := 0; i < len(timenow); i++ {
		if []rune(lasttime)[i] != []rune(timenow)[i] {
			mtx.OutputChar(i, FontZXSpectrumRus, []rune(timenow)[i], true)
		}
	}

	mtx.Close()
}

func ExperimentalFont(text string) {
	mtx := NewMatrix(Global.LedsLength, Global.LedsRotation)
	err := mtx.Open(0, 0, 1)
	if err != nil {
		fmt.Printf("error: %s\n", err)
	}

	mtx.Clear()
	converted, lenas := UnicodeToAsciiChar(text)
	fmt.Printf("Got result after convert: %s\n", converted)
	fmt.Printf("Text length: %d\n", lenas)
	fmt.Printf("Sliding text...\n")
	mtx.SlideMessage(converted, FontLT, true, 50*time.Millisecond)
	fmt.Printf("Sleeping for some time\n")
	time.Sleep(2 * time.Second)
	for i := 0; i < lenas; i++ {
		fmt.Printf("Printing char: %d\n", []rune(converted)[i])
		mtx.OutputChar(i, FontLT, []rune(converted)[i], true)

	}
	mtx.Close()
}

func DisplayClock() {
	if Global.ClockAtIdle {
		mtx := NewMatrix(Global.LedsLength, Global.LedsRotation)
		err := mtx.Open(0, 0, 1)
		if err != nil {
			log.Fatal(err)
		}
		// 08:42
		currentTime := time.Now()
		TimeNow := fmt.Sprintf("%s", currentTime.Format("15:04:05"))

		for i := 0; i < len(TimeNow); i++ {
			mtx.OutputChar(i, FontZXSpectrumRus, []rune(TimeNow)[i], true)
		}

		mtx.Close()
	}
}

func ShowError(error_number int, show_seconds int) {
	DISPLAY_INUSE = true
	MSG := fmt.Sprintf("Error:%d", error_number)
	mtx := NewMatrix(Global.LedsLength, Global.LedsRotation)
	err := mtx.Open(0, 0, 1)
	if err != nil {
		fmt.Printf("error opening led matrix\n")
		return
	}

	for i := 0; i < len(MSG); i++ {
		mtx.OutputChar(i, FontZXSpectrumRus, []rune(MSG)[i], true)
	}
	time.Sleep(time.Duration(show_seconds) * time.Second)
	mtx.Close()
	DISPLAY_INUSE = false
}

func ShowMsg(text string, seconds int) {
	DISPLAY_INUSE = true
	MSG := fmt.Sprintf("%s", text)
	mtx := NewMatrix(Global.LedsLength, Global.LedsRotation)
	err := mtx.Open(0, 0, 1)
	if err != nil {
		fmt.Printf("error opening led matrix\n")
		return
	}

	//fmt.Printf("text length: %d\n",len(MSG))
	for i := 0; i < len(MSG); i++ {
		mtx.OutputChar(i, FontZXSpectrumRus, []rune(MSG)[i], true)
	}
	time.Sleep(time.Duration(seconds) * time.Second)
	mtx.Close()
	DISPLAY_INUSE = false

}

// new implementation
func WriteMessage(text string) {
        mtx := NewMatrix(Global.LedsLength, Global.LedsRotation)
        err := mtx.Open(0, 0, 1)
        if err != nil {
                fmt.Printf("error: %s\n", err)
        }

        mtx.Clear()
        converted, _ := UnicodeToAsciiChar(text)
        mtx.SlideMessage(converted, FontLT, true, 50*time.Millisecond)
        mtx.Close()

}

// old implementation
/*func WriteMessage(text string) {
	mtx := NewMatrix(Global.LedsLength, Global.LedsRotation)
	err := mtx.Open(0, 0, 1)
	if err != nil {
		log.Fatal(err)
	}
	mtx.Clear()
	mtx.SlideMessage(text, FontCP437, true, 50*time.Millisecond)
	mtx.Close()
}*/

func main() {
	Global.Load("config.json")
	experimental := flag.String("experimental", "", "experimental function")
	flag.Parse()
	if *experimental != "" {
		ExperimentalFont(*experimental)
		os.Exit(0)
	}

	ShowMsg("Loading.", 3)
	go RegisterZeroConf()
	go httpapi()
	go MainQueueLoop()
	ShowMsg(fmt.Sprintf("OK v%v", APP_VERSION), 5)
	////// CLOCK INITIALIZATION
	// Just a clock shit
	last_time := "99-99-99" // set the dumb text
	mtx := NewMatrix(Global.LedsLength, Global.LedsRotation)
	err := mtx.Open(0, 0, 1)
	if err != nil {
		fmt.Printf("Unable to handle led matrix")
		return
	}

	for {
		// here we should generate the clock if needed
		if DISPLAY_INUSE == false && Global.ClockAtIdle {
			currentTime := time.Now()
			TimeNow := fmt.Sprintf("%s", currentTime.Format("15:04:05"))
			for i := 0; i < len(TimeNow); i++ {
				if []rune(last_time)[i] != []rune(TimeNow)[i] {
					mtx.OutputChar(i, FontZXSpectrumRus, []rune(TimeNow)[i], true)
				}
			}
			last_time = TimeNow
		}
		time.Sleep(1 * time.Second)
	}
	mtx.Close()
}
