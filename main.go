package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	_ "image/png"

	wr "github.com/mroth/weightedrand"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

var nextBirdSpawnTime time.Time
var context feederContext
var soundBuffers = make(map[string]*beep.Buffer)

// Set of channels that if flipped to 'true' terminate the background sound loop for their thread.
var backgroundSoundControllers []chan bool
var backgroundSoundControllerIndex int

// Whether or not sound is globally disabled.
var soundDisabled bool

var backgroundImd *imdraw.IMDraw
var seedsImd *imdraw.IMDraw

var updateAvailableEndpoint = "https://n4jexxccj8.execute-api.us-east-2.amazonaws.com/default/UpdateAvailable"
var versionNumber = float64(1.0)

func run() {
	// Seed the randomness.
	rand.Seed(time.Now().UnixNano())

	// Establish the game window.
	cfg := pixelgl.WindowConfig{
		Title:  "Feeder",
		Bounds: pixel.R(0, 0, winWidth, winHeight),
		VSync:  true,
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	// Establish new bird list.
	birds := []*bird{}

	// Establish the canvas.
	canvas := pixelgl.NewCanvas(pixel.R(-winWidth/2, -winHeight/2, winWidth/2, winHeight/2))
	globalImd := imdraw.New(nil)

	// Establish a camera position.
	camPos := pixel.ZV

	// Establish the new feeder context. Default to standard house feeder, and show feeder context selection menu.
	initializeFeederContext(win, canvas, globalImd)

	// Initialize and buffer all game sounds, play background sounds.
	initializeSounds()

	// Initialize the pause menu.
	pauseMenu := pauseMenu{}
	pauseMenu.PreRender()

	// Establish when the first bird will be spawned.
	setNextBirdSpawnTime()

	// Document the last time before the game loop begins.
	last := time.Now()

	for !win.Closed() {
		// Track the time elapsed since the last frame.
		elapsed := time.Since(last).Seconds()
		last = time.Now()

		// Escape shows the pause menu and allows the user to change feeder contexts (which will clear all birds).
		pauseMenu.justOpenedMenu = false
		if win.JustPressed(pixelgl.KeyEscape) && !pauseMenu.open {
			pauseMenu.open = true
			pauseMenu.justOpenedMenu = true
		}

		// Pressing enter will refill the seed by a constant percentage.
		if win.JustPressed(pixelgl.KeyEnter) && !pauseMenu.open {
			context.Seed().refill()
		}

		// Determine new birds / remove birds.
		birds = resolveBirds(birds)

		// Keep the camera position towards the center of the feeder.
		camPos = pixel.Lerp(camPos, context.Seed().center, 1)
		cam := pixel.IM.Moved(camPos.Scaled(-1))
		canvas.SetMatrix(cam)

		for _, bird := range birds {
			// Update the physics and animation.
			bird.update(elapsed)
			bird.animation.update(elapsed, bird)
		}

		// Clear the scene to be re-drawn.
		canvas.Clear(colornames.Black)
		globalImd.Clear()
		if backgroundImd != nil {
			backgroundImd.Clear()
		}
		if seedsImd != nil {
			seedsImd.Clear()
		}

		for _, bird := range birds {
			bird.animation.imd.Clear()
		}

		// Reset the background as it could change based on time.
		backgroundPicture := resolveBackgroundPicture()
		backgroundImd = imdraw.New(backgroundPicture)
		backgroundSprite := pixel.NewSprite(nil, pixel.Rect{})
		backgroundSprite.Set(backgroundPicture, pixel.Rect{Min: pixel.Vec{X: 0, Y: 0}, Max: pixel.Vec{X: 1600, Y: 900}})

		// Draw the background sprite. Hardcoding the pixel calculations, cry about it if you want.
		backgroundSprite.Draw(backgroundImd, pixel.IM.ScaledXY(pixel.ZV, pixel.V(1.25, 1.25)).Moved(pixel.Vec{X: 200, Y: 112.5}))

		// Draw the seed. Can change based on time so always re-set.
		seedsPicture := resolveSeedsPicture()
		seedsImd = imdraw.New(seedsPicture)
		context.Seed().draw(seedsImd, birds, seedsPicture)

		// Draw stationary birds first.
		for _, bird := range birds {
			if !bird.entering && !bird.exiting {
				bird.animation.draw(bird.physics)
			}
		}

		// ...then draw the flying birds.
		for _, bird := range birds {
			if bird.entering || bird.exiting {
				bird.animation.draw(bird.physics)
			}
		}

		// Render the pause menu if it is open.
		if pauseMenu.open {
			pauseMenu.Render(win, globalImd, canvas, &birds)
		}

		// Draw to the canvas.
		seedsImd.Draw(canvas)
		backgroundImd.Draw(canvas)
		globalImd.Draw(canvas)

		// Draw birds to the canvas now.
		for _, bird := range birds {
			if !bird.entering && !bird.exiting {
				bird.animation.imd.Draw(canvas)
			}
		}

		// ...then draw the flying birds to the canvas.
		for _, bird := range birds {
			if bird.entering || bird.exiting {
				bird.animation.imd.Draw(canvas)
			}
		}

		// Draw the pause menu when open.
		if pauseMenu.open {
			pauseMenu.Show(canvas)
		}

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
}

func initializeFeederContext(win *pixelgl.Window, canvas *pixelgl.Canvas, imd *imdraw.IMDraw) {
	// Establish the new feeder context. Default to standard house feeder.
	context = feederContextMappings[standardHouseFeederName]
	context.Initialize()

	// Show feeder context selection menu.
	showMenu(win, canvas, imd, &[]*bird{})
}

func resolveBirds(birds []*bird) []*bird {
	// Remove birds due for removal.
	birds = removeBirds(birds)

	// If it's night time, birds don't show up. But it 'can' happen.
	if getTimeOfDay() == "night" && nextRandomInt(0, defaultNumChancesOfNightBird) != 1 {
		return birds
	}

	// Add any new birds.
	if success, newBird := resolveNewBird(birds); success {
		birds = append(birds, newBird)
		setNextBirdSpawnTime()
	}

	// Return the new birds list.
	return birds
}

func resolveNewBird(birds []*bird) (bool, *bird) {
	// Don't span a bird until the next bird spawn time has been achieved.
	if time.Now().Before(nextBirdSpawnTime) {
		return false, &bird{}
	}

	// Create a bird given the remaining space.
	return birdFactory()
}

func resolveNewSpawnLengthInSeconds() (int, error) {
	// Enumerate all spawn length category choices.
	choices := []wr.Choice{}
	for category, likelihood := range context.TimeLikelihoods() {
		choices = append(choices, wr.Choice{Item: category, Weight: likelihood})
	}

	// Initialize a weighted probability spawn length chooser.
	chooser, _ := wr.NewChooser(choices...)

	// Pick a random spawn length category.
	lengthCategory := chooser.Pick().(timeLength)

	// Depending on the spawn length category picked, pick a random length of between its spawn length range.
	switch lengthCategory {
	case veryShort:
		minMaxPair := context.SpawnLengthRanges()[veryShort]
		return nextRandomInt(minMaxPair.min, minMaxPair.max), nil
	case short:
		minMaxPair := context.SpawnLengthRanges()[short]
		return nextRandomInt(minMaxPair.min, minMaxPair.max), nil
	case medium:
		minMaxPair := context.SpawnLengthRanges()[medium]
		return nextRandomInt(minMaxPair.min, minMaxPair.max), nil
	case long:
		minMaxPair := context.SpawnLengthRanges()[long]
		return nextRandomInt(minMaxPair.min, minMaxPair.max), nil
	case veryLong:
		minMaxPair := context.SpawnLengthRanges()[veryLong]
		return nextRandomInt(minMaxPair.min, minMaxPair.max), nil
	case insane:
		minMaxPair := context.SpawnLengthRanges()[insane]
		return nextRandomInt(minMaxPair.min, minMaxPair.max), nil
	}

	return 0, errors.New("Invalid spawn length category selected")
}

func birdFactory() (bool, *bird) {
	// Enumerate all bird choices.
	choices := []wr.Choice{}
	for bird, likelihood := range context.BirdLikelihoods() {
		choices = append(choices, wr.Choice{Item: bird, Weight: likelihood})
	}

	// Initialize a weighted probability bird chooser.
	chooser, _ := wr.NewChooser(choices...)

	// Pick a random bird.
	birdSpeciesPick := chooser.Pick().(species)
	birdSpeciesPick.Initialize()

	// Pick a random (available) perch.
	perchFound, newPerch := getRandomPerch(birdSpeciesPick)

	// Determine if it's possible to fit this bird on the feeder.
	if !perchFound {
		// Set back bird spawn time when all spots are filled. NOTE: Potentially lower this to a set amount of delay.
		setNextBirdSpawnTime()
		return false, &bird{}
	}

	// Load the sprite/animation.
	animationSheet, birdAnimations, err := loadAnimationSheet(birdSpeciesPick.Animation(), animationMappingsFile, standardSpriteWidth)
	if err != nil {
		panic(err)
	}

	// Create the bird, along with animations and new default flight script.
	return true, newBird(defaultBirdFrameRate, birdSpeciesPick, birdAnimations, animationSheet, newPerch, imdraw.New(animationSheet))
}

func removeBirds(birds []*bird) []*bird {
	indexesForDeletion := []int{}
	for index, bird := range birds {
		// Birds that have hit their removal time should begin the exit process.
		if birdShouldExit(bird) {
			bird.setExitStatus()
		}

		// Mark the fully removed birds to be deleted.
		if bird.removed {
			indexesForDeletion = append(indexesForDeletion, index)
		}
	}

	// Delete the indexes of each bird that should be removed.
	for _, index := range indexesForDeletion {
		birds = append(birds[:index], birds[index+1:]...)
	}

	return birds
}

func birdShouldExit(bird *bird) bool {
	return !bird.entering && !bird.eating && !bird.removed && !bird.exiting && !bird.singing && time.Now().After(bird.removalTime)
}

func setNextBirdSpawnTime() {
	// Establish when the first bird will be spawned.
	newBirdSpawnLength, err := resolveNewSpawnLengthInSeconds()
	if err != nil {
		newBirdSpawnLength = defaultSpawnLength
	}

	nextBirdSpawnTime = time.Now().Add(time.Second * time.Duration(newBirdSpawnLength))
}

func getRandomPerch(species species) (bool, *perch) {
	availablePerches := []*perch{}

	// Enumerate all available (unoccupied) perches.
	for _, perch := range context.Perches() {
		perchWidth := perch.X.max - perch.X.min
		perchHeight := perch.Y.max - perch.Y.min

		// Perch must be unoccupied and spacious enough to accomodate this species.
		if !perch.occupied && perchWidth >= species.Width() && perchHeight >= species.Height() {
			availablePerches = append(availablePerches, perch)
		}
	}

	// No perches found, return unsuccessful.
	if len(availablePerches) == 0 {
		return false, &perch{}
	}

	// A perch exists, choose a random one.
	perchIndex := 0

	if len(availablePerches) > 1 {
		perchIndex = nextRandomInt(0, len(availablePerches)-1)
	}

	return true, availablePerches[perchIndex]
}

func showMenu(win *pixelgl.Window, canvas *pixelgl.Canvas, imd *imdraw.IMDraw, currentBirds *[]*bird) {
	menu := &mainMenu{}
	menu.ShowMainMenu(win, canvas, imd, currentBirds)
}

func bufferContextSounds() {
	// Buffer all sounds.
	for sound, path := range context.Sounds() {
		soundBuffers[sound] = bufferSound(path)
	}
}

func initializeSounds() {
	// Reset sound buffers.
	soundBuffers = make(map[string]*beep.Buffer)

	// Stop playing current sounds.
	speaker.Clear()
	speaker.Init(44100, int(time.Second/60*time.Duration(44100)/time.Second))

	// Initialize the new sounds.
	bufferContextSounds()

	// Terminate any current background sound through the current channel for it.
	if len(backgroundSoundControllers) != 0 {
		close(backgroundSoundControllers[backgroundSoundControllerIndex])
		backgroundSoundControllerIndex++
	}

	// Play background sound. Include terminator channel.
	backgroundSoundController := make(chan bool)
	backgroundSoundControllers = append(backgroundSoundControllers, backgroundSoundController)
	go playLoopingSound(soundBuffers["Background"], backgroundSoundControllers[backgroundSoundControllerIndex])
}

func enableSounds() {
	soundDisabled = false
}

func disableSounds() {
	speaker.Clear()
	soundDisabled = true
}

func resolveBackgroundPicture() pixel.Picture {
	return context.Backgrounds()[getTimeOfDay()]
}

func resolveSeedsPicture() pixel.Picture {
	return context.Seeds()[getTimeOfDay()]
}

func getTimeOfDay() string {
	currentHour := time.Now().Hour()
	var timeOfDay string

	// Whatever man. Just show the dusk image an hour before the sunset/sunrise time, and also during that hour.
	if float64(defaultSunriseHour-currentHour) == 1 ||
		float64(defaultSunriseHour-currentHour) == 0 ||
		float64(defaultSunsetHour-currentHour) == 1 ||
		float64(defaultSunsetHour-currentHour) == 0 {
		timeOfDay = "dusk"
	} else if defaultSunriseHour < currentHour && defaultSunsetHour > currentHour {
		timeOfDay = "day"
	} else {
		timeOfDay = "night"
	}

	return timeOfDay
}

func needsUpdate() bool {
	resp, err := http.Get(updateAvailableEndpoint + "?V=" + fmt.Sprint(versionNumber))

	// This is fine, they aren't connected to the internet.
	if err != nil {
		return false
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	result, err := strconv.ParseBool(string(body))

	if err != nil {
		panic(err)
	}

	return result
}

func main() {
	pixelgl.Run(run)
}
