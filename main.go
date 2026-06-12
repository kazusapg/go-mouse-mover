package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/getlantern/systray"
	"github.com/go-vgo/robotgo"
	"github.com/kazusapg/go-mouse-mover/assets"
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

var (
	currentCancel  context.CancelFunc
	cancelMu       sync.Mutex
	hookMu         sync.Mutex
	hookInUse      bool
	trayStartItem  *systray.MenuItem
	trayStopItem   *systray.MenuItem
	trayStatusItem *systray.MenuItem
)

func main() {
	go runInteractive()
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(assets.DefaultIcon())
	systray.SetTitle("Mouse Mover")
	systray.SetTooltip("go-mouse-mover: Idle")

	trayStatusItem = systray.AddMenuItem("Status: Idle", "Current state")
	trayStatusItem.Disable()
	systray.AddSeparator()

	trayStartItem = systray.AddMenuItem("Start", "Start automatic movement")
	trayStopItem = systray.AddMenuItem("Stop", "Stop automatic movement")
	trayStopItem.Disable()

	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit the application")

	go func() {
		for {
			select {
			case <-trayStartItem.ClickedCh:
				startMovingFromTray()
			case <-trayStopItem.ClickedCh:
				stopMoving()
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func onExit() {
	stopMoving()
	// If recording is in progress (hook active but no movement),
	// terminate the hook to release native resources.
	hookMu.Lock()
	inUse := hookInUse
	hookMu.Unlock()
	if inUse {
		hook.End()
	}
}

func runInteractive() {
	defer systray.Quit()
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("[1]Input mouse positions [2]Move mouse [3]End")
		fmt.Printf(">")
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				fmt.Fprintln(os.Stderr, "Error reading input:", err)
			}
			return
		}
		mode, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
		if err != nil {
			fmt.Println("Invalid input. Please enter 1, 2, or 3.")
			continue
		}

		switch mode {
		case 1:
			positions := recordPos()
			if positions == nil {
				continue
			}
			if len(positions) == 0 {
				fmt.Println("No positions were recorded.")
				continue
			}
			intervalMilliSecond := inputIntervalMilliSecond(nil)
			mi := moveInfo{Positions: positions, IntervalMillisecond: intervalMilliSecond}
			file, err := json.MarshalIndent(mi, "", "  ")
			if err != nil {
				fmt.Println("Failed to encode JSON:", err)
				continue
			}
			err = os.WriteFile(jsonName, file, 0644)
			if err != nil {
				fmt.Println("Failed to save file:", err)
				continue
			}
			fmt.Println("Saved to", jsonName)
		case 2:
			jsonFile, err := os.ReadFile(jsonName)
			if err != nil {
				fmt.Println("Failed to read file:", err)
				continue
			}

			var mi moveInfo
			err = json.Unmarshal(jsonFile, &mi)
			if err != nil {
				fmt.Println("Failed to decode JSON:", err)
				continue
			}

			moveMouse(mi)
		case 3:
			fmt.Println("End")
			return
		default:
			fmt.Println("Invalid mode. Please enter 1, 2, or 3.")
		}
	}
}

// tryAcquireHook atomically checks if a hook-based operation (recording or moving)
// is already in progress and sets hookInUse if not.
func tryAcquireHook() bool {
	hookMu.Lock()
	defer hookMu.Unlock()
	if hookInUse {
		return false
	}
	hookInUse = true
	return true
}

// releaseHook marks the hook as no longer in use.
func releaseHook() {
	hookMu.Lock()
	hookInUse = false
	hookMu.Unlock()
}

func recordPos() []position {
	if !tryAcquireHook() {
		fmt.Println("Another operation is already in progress.")
		return nil
	}
	defer releaseHook()

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

// tryStartMoving atomically checks if movement is already in progress
// and sets currentCancel if not. Returns the context and cancel function,
// or (nil, nil, false) if already moving.
func tryStartMoving() (context.Context, context.CancelFunc, bool) {
	cancelMu.Lock()
	defer cancelMu.Unlock()
	if currentCancel != nil {
		return nil, nil, false
	}
	ctx, cancel := context.WithCancel(context.Background())
	currentCancel = cancel
	return ctx, cancel, true
}

func moveMouse(mi moveInfo) {
	if !tryAcquireHook() {
		fmt.Println("Another operation is already in progress.")
		return
	}
	defer releaseHook()

	ctx, cancel, ok := tryStartMoving()
	if !ok {
		fmt.Println("Already moving.")
		return
	}

	fmt.Println("The mouse starts moving. To stop moving, press Ctrl+Shift+q.")

	setTrayMoving(true)
	defer setTrayMoving(false)
	defer func() {
		cancelMu.Lock()
		currentCancel = nil
		cancelMu.Unlock()
	}()

	hook.Register(hook.KeyDown, []string{"q", "ctrl", "shift"}, func(e hook.Event) {
		cancel()
		hook.End()
	})

	go moveMousePosition(ctx, mi)

	s := hook.Start()
	<-hook.Process(s)
}

func stopMoving() {
	cancelMu.Lock()
	cancel := currentCancel
	cancelMu.Unlock()
	if cancel != nil {
		cancel()
		hook.End()
		hookMu.Lock()
		hookInUse = false
		hookMu.Unlock()
	}
}

func startMovingFromTray() {
	jsonFile, err := os.ReadFile(jsonName)
	if err != nil {
		log.Println(err)
		return
	}

	var mi moveInfo
	err = json.Unmarshal(jsonFile, &mi)
	if err != nil {
		log.Println(err)
		return
	}

	if len(mi.Positions) == 0 {
		fmt.Println("No positions recorded. Please record positions first.")
		return
	}

	go moveMouse(mi)
}

func setTrayMoving(moving bool) {
	if moving {
		systray.SetTooltip("go-mouse-mover: Moving")
		systray.SetIcon(assets.ActiveIcon())
		if trayStatusItem != nil {
			trayStatusItem.SetTitle("Status: Moving")
		}
		if trayStartItem != nil {
			trayStartItem.Disable()
		}
		if trayStopItem != nil {
			trayStopItem.Enable()
		}
	} else {
		systray.SetTooltip("go-mouse-mover: Idle")
		systray.SetIcon(assets.DefaultIcon())
		if trayStatusItem != nil {
			trayStatusItem.SetTitle("Status: Idle")
		}
		if trayStartItem != nil {
			trayStartItem.Enable()
		}
		if trayStopItem != nil {
			trayStopItem.Disable()
		}
	}
}

func moveMousePosition(ctx context.Context, mi moveInfo) {
	if len(mi.Positions) == 0 {
		fmt.Println("No positions to move to.")
		return
	}

	t := time.NewTicker(time.Duration(mi.IntervalMillisecond) * time.Millisecond)
	positionLength := len(mi.Positions)
	var currentNum int
	count := 1
	for {
		select {
		case <-ctx.Done():
			t.Stop()
			fmt.Println("\nStop.")
			return
		case <-t.C:
			if currentNum >= positionLength {
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
