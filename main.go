package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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
			intervalMilliSecond := inputIntervalMilliSecond(nil)
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
	fmt.Println("Please click the left mouse button to record mouse positions.")
	fmt.Println("To quit recording mouse positions, press Ctrl+Shift+q.")

	hook.Register(hook.KeyDown, []string{"q", "ctrl", "shift"}, func(e hook.Event) {
		hook.End()
	})

	hook.Register(hook.MouseDown, []string{}, func(e hook.Event) {
		x, y := robotgo.Location()
		c := position{X: x, Y: y}
		positions = append(positions, c)
		fmt.Println(x, y)
	})

	s := hook.Start()
	<-hook.Process(s)

	return positions
}

func inputIntervalMilliSecond(in io.Reader) int {
	if in == nil {
		in = os.Stdin
	}

	var waitMilliSecond int
	for {
		var inputMilliSecond string
		fmt.Println("Please enter the interval in milliseconds at which the mouse moves to the next coordinate.")
		fmt.Println("If you want to wait 1 second, enter 1000.")
		fmt.Print(">")
		fmt.Fscan(in, &inputMilliSecond)
		i, err := strconv.Atoi(inputMilliSecond)
		if err != nil {
			fmt.Println("The interval milliseconds must be number.")
			continue
		}
		if i < 1 {
			fmt.Println("The interval milliseconds must be greater than or equal to 1.")
			continue
		}
		waitMilliSecond = i
		break
	}
	return waitMilliSecond
}

func moveMouse(mi moveInfo) {
	fmt.Println("The mouse starts moving. To stop moving, press Ctrl+Shift+q.")

	ctx, cancel := context.WithCancel(context.Background())
	hook.Register(hook.KeyDown, []string{"q", "ctrl", "shift"}, func(e hook.Event) {
		cancel()
		hook.End()
	})

	go moveMousePosition(ctx, mi)

	s := hook.Start()
	<-hook.Process(s)
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
