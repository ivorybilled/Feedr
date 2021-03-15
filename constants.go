package main

type pair struct {
	min, max int
}

type coordinatePair struct {
	min, max float64
}

const (
	winHeight                   = 900
	winWidth                    = 1600
	defaultSpawnLength          = 100
	defaultFeederLength         = 10
	defaultEatingGap            = 10
	defaultEatingLength         = 10
	defaultSingingGap           = 60
	defaultSeedRefillMultiplier = .05
	defaultBirdFrameRate        = 1.0 / 10
	standardHouseFeederName     = "Backyard Sunflower Feeder"
	pauseMenuHeight             = float64(450)
	pauseMenuWidth              = float64(1000)
	spawnRandomnessOffset       = 500
	likelihoodMaxPercent        = 1000
	animationMappingsFile       = "animationMap/animationMappings.csv"
	standardSpriteWidth         = 43

	// Extremely rough 'pretend' times.
	defaultSunriseHour = 7
	defaultSunsetHour  = 19

	// Chances of a bird showing up at night is 1 in this number.
	defaultNumChancesOfNightBird = 1000000000000
)

type animState int

const (
	perched animState = iota
	eating
	flying
	singing
)

type timeLength int

const (
	veryShort timeLength = iota
	short
	medium
	long
	veryLong
	insane
)

// List of each available feeder 'context' in the game.
var feederContextMappings = map[string]feederContext{
	standardHouseFeederName: &standardHouseFeeder{},
}

// List of each available feeder 'context' in the game.
var feederContexts = []string{
	standardHouseFeederName,
}

// Menu title text.
var menuTitleText = "Welcome to Feeder."

// Menu subtitle text.
var menuSubtitleText = "Select a feeder scene with the arrow keys, then press enter.\n"

// The text to show at the bottom of the menu.
var menuAdditionalText = "\nPress enter repeatedly to refill the feeder. \nPress escape to access/exit this menu.\n"

// When an update is found.
var updateRequiredText = "\nNew version with more birds/feeders available at feeder.com!\n"

// Duh.
var creditText = "\nDeveloped by David Bennett. Art by Camila Canuto."
