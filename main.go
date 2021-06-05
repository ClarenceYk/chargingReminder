package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"time"
)

const (
	powerPercent80Display = "🔋 ⬆️ 80%"
	powerPercent20Display = "🔋 ⬇️ 20%"
	powerUpperLimit       = 59
	powerLowerLimit       = 58
)

func main() {
	var (
		bInfo               = getBatteryInfo()
		dischargingNotified = false
		acAttachingNotified = false
	)

	for {
		fmt.Println(bInfo)
		switch bInfo.state {
		case "AC attached", "charging":
			acAttachingNotified = false

			if bInfo.powerPercent >= powerUpperLimit && !dischargingNotified {
				fmt.Println("run charging notify")
				runCommands(getChargingTo80Notify())
				dischargingNotified = true
			}
		case "discharging":
			dischargingNotified = false

			if bInfo.powerPercent <= powerLowerLimit && !acAttachingNotified {
				fmt.Println("run drop notify")
				runCommands(getDropTo20Notify())
				acAttachingNotified = true
			}
		default:
			fmt.Println("unknown battery state: [" + bInfo.state + "]")
			dischargingNotified = false
			acAttachingNotified = false
		}

		time.Sleep(time.Second * 5)
		bInfo = getBatteryInfo()
	}
}

type batteryInfo struct {
	powerPercent int
	state        string
}

func getBatteryInfo() batteryInfo {
	var (
		powerSourceShow = exec.Command("pmset", "-g", "batt")
		bi              = batteryInfo{}
	)

	out, err := powerSourceShow.Output()
	if err != nil {
		fmt.Println("run:", powerSourceShow.String(), "error:", err)
		return bi
	}

	out = bytes.TrimSpace(out)
	lines := bytes.Split(out, []byte{'\n'})
	if len(lines) < 2 {
		fmt.Println("run:", powerSourceShow.String(), "error:", "number of output lines less than 2")
		return bi
	}

	fields := bytes.Split(lines[1], []byte{';'})
	if len(fields) < 2 {
		fmt.Println("run:", powerSourceShow.String(), "error:", "number of fields less than 2")
		return bi
	}

	if len(fields[0]) > 2 {
		fLen := len(fields[0])
		percent, err := strconv.ParseInt(string(fields[0][fLen-3:fLen-1]), 10, 0)
		if err == nil {
			bi.powerPercent = int(percent)
		}
	}
	bi.state = string(bytes.TrimSpace(fields[1]))

	return bi
}

func runCommands(cmds []*exec.Cmd) {
	for _, cmd := range cmds {
		_ = cmd.Run()
	}
}

func getChargingTo80Notify() []*exec.Cmd {
	return []*exec.Cmd{
		exec.Command("osascript", "-e", "display notification \""+powerPercent80Display+"\""),
		exec.Command("afplay", "/System/Library/Sounds/Ping.aiff", "-v", "6"),
	}
}

func getDropTo20Notify () []*exec.Cmd {
	return []*exec.Cmd{
		exec.Command("osascript", "-e", "display notification \""+powerPercent20Display+"\""),
		exec.Command("afplay", "/System/Library/Sounds/Ping.aiff", "-v", "6"),
	}
}
