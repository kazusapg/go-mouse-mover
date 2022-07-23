package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/go-vgo/robotgo"
	hook "github.com/robotn/gohook"
)

type position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type moveInfo struct {
	Positions           []position `json:"positions"`
	IntervalMillisecond int        `json:"interval_millisecond"`
}

const jsonName = "moveinfo.json"

func main() {

Loop:
	for {
		var mode int
		fmt.Println("[1]Input mouse positions [2]Move mouse [3]End")
		fmt.Printf(">")
		fmt.Scan(&mode)

		switch mode {
		case 1:
			positions := recordPos()
			intervalMilliSecond := inputIntervalMilliSecond()
			mi := moveInfo{Positions: positions, IntervalMillisecond: intervalMilliSecond}
			file, err := json.MarshalIndent(mi, "", "  ")
			if err != nil {
				log.Println(err)
				break Loop
			}
			err = os.WriteFile(jsonName, file, os.ModePerm)
			if err != nil {
				log.Println(err)
				break Loop
			}
		case 2:
			jsonFile, err := os.ReadFile(jsonName)
			if err != nil {
				log.Println(err)
				break Loop
			}

			var mi moveInfo
			err = json.Unmarshal(jsonFile, &mi)
			if err != nil {
				log.Println(err)
				break Loop
			}

			moveMouse(mi)
		case 3:
			fmt.Println("End")
			return
		}
	}
}

func recordPos() []position {
	var positions = []position{}
	fmt.Println("Please mouse left click to record mouse positions.")
	fmt.Println("To quit recording mouse positions, please enter Ctrl+Shift+q.")

	robotgo.EventHook(hook.KeyDown, []string{"q", "ctrl", "shift"}, func(e hook.Event) {
		robotgo.EventEnd()
	})

	robotgo.EventHook(hook.MouseDown, []string{}, func(e hook.Event) {
		x, y := robotgo.GetMousePos()
		c := position{X: x, Y: y}
		positions = append(positions, c)
		fmt.Println(x, y)
	})

	s := robotgo.EventStart()
	<-robotgo.EventProcess(s)

	return positions
}

func inputIntervalMilliSecond() int {
	var waitMilliSecond int
	for {
		var inputMilliSecond string
		fmt.Println("Please enter the interval in milliseconds at which the mouse moves to the next coordinate.")
		fmt.Println("If you want to wait 1 second, enter 1000.")
		fmt.Print(">")
		fmt.Scan(&inputMilliSecond)
		i, err := strconv.Atoi(inputMilliSecond)
		if err != nil {
			fmt.Println("The interval milliseconds must be number.")
			continue
		}
		if i < 1 {
			fmt.Println("The interval milliseconds must be greater than or equal to 1")
			continue
		}
		waitMilliSecond = i
		break
	}
	return waitMilliSecond
}

func moveMouse(mi moveInfo) {
	fmt.Println("Mouse moving start. To stop enter Ctrl+Shift+q.")

	ctx, cancel := context.WithCancel(context.Background())
	robotgo.EventHook(hook.KeyDown, []string{"q", "ctrl", "shift"}, func(e hook.Event) {
		cancel()
		robotgo.EventEnd()
	})

	go moveMousePosition(ctx, mi)

	s := robotgo.EventStart()
	<-robotgo.EventProcess(s)
}

func moveMousePosition(ctx context.Context, mi moveInfo) {
	t := time.NewTicker(time.Duration(mi.IntervalMillisecond) * time.Millisecond)
	postionLength := len(mi.Positions)
	var currentNum int
	count := 1
	for {
		select {
		case <-ctx.Done():
			t.Stop()
			fmt.Println("\nStop.")
			return
		case <-t.C:
			if currentNum >= postionLength {
				currentNum = 0
			}
			position := mi.Positions[currentNum]
			fmt.Printf("\rCount:%d. Position %+v", count, position)
			robotgo.Move(position.X, position.Y)
			count++
			currentNum++
		}
	}
}
