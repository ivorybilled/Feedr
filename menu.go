package main

import (
	"fmt"
	"math"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

type mainMenu struct {
	text *text.Text
}

func (menu *mainMenu) NumOptionIndexes() int {
	additionalOptions := 1
	return (len(feederContexts) - 1) + additionalOptions
}

func (menu *mainMenu) ShowMainMenu(win *pixelgl.Window, canvas *pixelgl.Canvas, imd *imdraw.IMDraw, currentBirds *[]*bird) {
	imd.Clear()
	menuClosed := false
	selectedOptionNumber := 0

	// Set the selected context number to the index of the current context.
	for index, name := range feederContexts {
		if context.Name() == name {
			selectedOptionNumber = index
			break
		}
	}

	xValue := canvas.Bounds().Min.X + 100
	yValue := canvas.Bounds().Max.Y - 100
	atlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	menu.text = text.New(pixel.Vec{X: xValue, Y: yValue}, atlas)
	menu.PrintMenuText(selectedOptionNumber, context.Name())

	for !win.Closed() && !menuClosed {
		// Increment/decrement selected feeder context selector, then re-print with the selection highlighted.
		if win.JustPressed(pixelgl.KeyDown) {
			if selectedOptionNumber == menu.NumOptionIndexes() {
				selectedOptionNumber = 0
			} else {
				selectedOptionNumber++
			}
		} else if win.JustPressed(pixelgl.KeyUp) {
			if selectedOptionNumber == 0 {
				selectedOptionNumber = menu.NumOptionIndexes()
			} else {
				selectedOptionNumber--
			}
		}

		menu.PrintMenuText(selectedOptionNumber, context.Name())

		canvas.Clear(colornames.Rosybrown)
		menu.text.Draw(canvas, pixel.IM.Scaled(menu.text.Orig, 3))
		win.SetMatrix(pixel.IM.Scaled(pixel.ZV,
			math.Min(
				win.Bounds().W()/canvas.Bounds().W(),
				win.Bounds().H()/canvas.Bounds().H(),
			),
		).Moved(win.Bounds().Center()))
		canvas.Draw(win, pixel.IM.Moved(canvas.Bounds().Center()))
		win.Update()

		// Leave the menu when enter or escape is pressed.
		if win.JustPressed(pixelgl.KeyEnter) {
			// Only re-initialize the feeder context if a *different* one is selected.
			if len(feederContexts) > selectedOptionNumber && context.Name() != feederContexts[selectedOptionNumber] {
				context = feederContextMappings[feederContexts[selectedOptionNumber]]
				context.Initialize()
				*currentBirds = nil
			} else if selectedOptionNumber == menu.NumOptionIndexes() {
				if soundDisabled {
					enableSounds()
				} else {
					disableSounds()
				}

				// Don't close the menu if muting/unmuting.
				continue
			}

			menuClosed = true
		} else if win.JustPressed(pixelgl.KeyEscape) {
			menuClosed = true
		}
	}

	// Set the next frame to be a 'loading' screen.
	if !win.Closed() {
		menu.ShowLoadingScreen(win, canvas, imd)
	}
}

func (menu *mainMenu) ShowLoadingScreen(win *pixelgl.Window, canvas *pixelgl.Canvas, imd *imdraw.IMDraw) {
	imd.Clear()

	xValue := canvas.Bounds().Min.X + 100
	yValue := canvas.Bounds().Max.Y - 100
	atlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	menu.text = text.New(pixel.Vec{X: xValue, Y: yValue}, atlas)
	menu.PrintLoadingText()

	canvas.Clear(colornames.Rosybrown)
	menu.text.Draw(canvas, pixel.IM.Scaled(menu.text.Orig, 3))
	win.SetMatrix(pixel.IM.Scaled(pixel.ZV,
		math.Min(
			win.Bounds().W()/canvas.Bounds().W(),
			win.Bounds().H()/canvas.Bounds().H(),
		),
	).Moved(win.Bounds().Center()))
	canvas.Draw(win, pixel.IM.Moved(canvas.Bounds().Center()))
	win.Update()
}

func (menu *mainMenu) PrintMenuText(selectedContextNumber int, currentContextName string) {
	menu.text.Clear()

	menu.text.LineHeight = .9 * menu.text.Atlas().LineHeight()
	menu.text.Color = colornames.White
	fmt.Fprintln(menu.text, menuTitleText)
	fmt.Fprintln(menu.text, "------------------------------")

	menu.text.Color = colornames.Pink
	fmt.Fprintln(menu.text, menuSubtitleText)

	for index, name := range feederContexts {
		// Show the current feeder in red.
		if index == selectedContextNumber {
			menu.text.Color = colornames.Red
		} else {
			menu.text.Color = colornames.Blue
		}

		if currentContextName == name {
			name = name + " (current)"
		}

		fmt.Fprintln(menu.text, name)
	}

	fmt.Fprintln(menu.text) // New line.
	menu.text.Color = colornames.White
	fmt.Fprintln(menu.text, "------------------------------")

	// Print the mute/unmute button.
	if selectedContextNumber == len(feederContexts) {
		menu.text.Color = colornames.Red
	} else {
		menu.text.Color = colornames.Blue
	}

	soundOption := "Mute"

	if soundDisabled {
		soundOption = "Unmute"
	}

	fmt.Fprintln(menu.text, soundOption)

	menu.text.Color = colornames.White
	fmt.Fprintln(menu.text, "------------------------------")

	menu.text.Color = colornames.Pink
	fmt.Fprintln(menu.text, menuAdditionalText)

	if needsUpdate() {
		fmt.Fprintln(menu.text, updateRequiredText)
	}

	menu.text.Color = colornames.White
	fmt.Fprintln(menu.text, "------------------------------")

	menu.text.Color = colornames.Pink
	fmt.Fprintln(menu.text, "\nVersion: "+fmt.Sprint(versionNumber))
	fmt.Fprintln(menu.text, creditText)
}

func (menu *mainMenu) PrintLoadingText() {
	menu.text.Clear()

	menu.text.Color = colornames.Pink
	fmt.Fprintln(menu.text, "Loading...")
}

type pauseMenu struct {
	width                float64
	height               float64
	center               pixel.Vec
	text                 *text.Text
	selectedOptionNumber int
	open                 bool
	justOpenedMenu       bool
	upperYBound          float64
	lowerYBound          float64
}

func (menu *pauseMenu) NumOptionIndexes() int {
	additionalOptions := 1
	return (len(feederContexts) - 1) + additionalOptions
}

func (menu *pauseMenu) Show(canvas *pixelgl.Canvas) {
	menu.text.Draw(canvas, pixel.IM.Scaled(menu.text.Orig, 2))
}

func (menu *pauseMenu) ShowLoadingScreen(win *pixelgl.Window, imd *imdraw.IMDraw, canvas *pixelgl.Canvas) {
	menu.PrintLoadingText()

	// Draw to the canvas.
	imd.Draw(canvas)

	// Draw the pause menu when open.
	menu.Show(canvas)

	// Stretch the canvas to the window.
	win.Clear(colornames.White)
	win.SetMatrix(pixel.IM.Scaled(pixel.ZV,
		math.Min(
			win.Bounds().W()/canvas.Bounds().W(),
			win.Bounds().H()/canvas.Bounds().H(),
		),
	).Moved(win.Bounds().Center()))
	canvas.Draw(win, pixel.IM.Moved(canvas.Bounds().Center()))
	win.Update()
}

func (menu *pauseMenu) Render(win *pixelgl.Window, imd *imdraw.IMDraw, canvas *pixelgl.Canvas, currentBirds *[]*bird) {
	imd.Color = colornames.Rosybrown

	// Top left point.
	imd.Push(pixel.Vec{X: menu.center.X - menu.width/2, Y: menu.upperYBound})

	// Top right point.
	imd.Push(pixel.Vec{X: menu.center.X + menu.width/2, Y: menu.upperYBound})

	// Bottom left point.
	imd.Push(pixel.Vec{X: menu.center.X - menu.width/2, Y: menu.lowerYBound})

	// Bottom right point.
	imd.Push(pixel.Vec{X: menu.center.X + menu.width/2, Y: menu.lowerYBound})

	imd.Rectangle(0)

	if win.JustPressed(pixelgl.KeyDown) {
		if menu.selectedOptionNumber == menu.NumOptionIndexes() {
			menu.selectedOptionNumber = 0
		} else {
			menu.selectedOptionNumber++
		}
	} else if win.JustPressed(pixelgl.KeyUp) {
		if menu.selectedOptionNumber == 0 {
			menu.selectedOptionNumber = menu.NumOptionIndexes()
		} else {
			menu.selectedOptionNumber--
		}
	}

	// Print to the text.
	menu.PrintMenuText(menu.selectedOptionNumber, context.Name())

	if win.JustPressed(pixelgl.KeyEnter) {
		// Only re-initialize the feeder context if a *different* one is selected.
		if len(feederContexts) > menu.selectedOptionNumber && context.Name() != feederContexts[menu.selectedOptionNumber] {
			// Reinitialize everything.
			menu.ShowLoadingScreen(win, imd, canvas)
			context = feederContextMappings[feederContexts[menu.selectedOptionNumber]]
			context.Initialize()
			*currentBirds = nil
			initializeSounds()
		} else if menu.selectedOptionNumber == menu.NumOptionIndexes() {
			if soundDisabled {
				enableSounds()
			} else {
				disableSounds()
			}

			return
		}

		menu.open = false
	} else if win.JustPressed(pixelgl.KeyEscape) && !menu.justOpenedMenu {
		menu.open = false
	}
}

func (menu *pauseMenu) PreRender() {
	menu.open = false

	// Set the selected context number to the index of the current context.
	for index, name := range feederContexts {
		if context.Name() == name {
			menu.selectedOptionNumber = index
			break
		}
	}

	// Set the geometry of the menu.
	menu.center = pixel.V(0, 0)
	menu.width = pauseMenuWidth
	menu.height = pauseMenuHeight
	menu.upperYBound = menu.center.Y + menu.height/2
	menu.lowerYBound = menu.center.Y - menu.height/2

	// Set location and intialize text object.
	xValue := (menu.center.X - menu.width/2) + 50
	yValue := menu.upperYBound - 65
	atlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	menu.text = text.New(pixel.Vec{X: xValue, Y: yValue}, atlas)

	// Print text.
	menu.PrintMenuText(menu.selectedOptionNumber, context.Name())
}

func (menu *pauseMenu) PrintLoadingText() {
	menu.text.Clear()

	menu.text.Color = colornames.Pink
	fmt.Fprintln(menu.text, "Loading...")
}

func (menu *pauseMenu) PrintMenuText(selectedContextNumber int, currentContextName string) {
	menu.text.Clear()

	menu.text.Color = colornames.White
	fmt.Fprintln(menu.text, menuTitleText)
	fmt.Fprintln(menu.text, "------------------------------")

	menu.text.Color = colornames.Pink
	fmt.Fprintln(menu.text, menuSubtitleText)

	for index, name := range feederContexts {
		// Show the current feeder in red.
		if index == selectedContextNumber {
			menu.text.Color = colornames.Red
		} else {
			menu.text.Color = colornames.Blue
		}

		if currentContextName == name {
			name = name + " (current)"
		}

		fmt.Fprintln(menu.text, name)
	}

	fmt.Fprintln(menu.text) // New line.
	menu.text.Color = colornames.White
	fmt.Fprintln(menu.text, "------------------------------")

	// Print the mute/unmute button.
	if selectedContextNumber == len(feederContexts) {
		menu.text.Color = colornames.Red
	} else {
		menu.text.Color = colornames.Blue
	}

	soundOption := "Mute"

	if soundDisabled {
		soundOption = "Unmute"
	}

	fmt.Fprintln(menu.text, soundOption)

	menu.text.Color = colornames.White
	fmt.Fprintln(menu.text, "------------------------------")

	menu.text.Color = colornames.Pink
	fmt.Fprintln(menu.text, menuAdditionalText)

	menu.text.Color = colornames.White
	fmt.Fprintln(menu.text, "------------------------------")
}
