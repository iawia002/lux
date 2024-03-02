package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"

	"github.com/iawia002/lux/app"
	"github.com/iawia002/lux/utils"
)

func main() {
	errFFMPEG := utils.IsFfmpegInstalled()
	if errFFMPEG != nil {
		fmt.Fprintf(
			color.Output,
			"Dependency not found: %s\n",
			color.RedString("%s", errFFMPEG),
		)
		os.Exit(1)
	}
	if err := app.New().Run(os.Args); err != nil {
		fmt.Fprintf(
			color.Output,
			"Run %s failed: %s\n",
			color.CyanString("%s", app.Name), color.RedString("%v", err),
		)
		os.Exit(2)
	}
}
