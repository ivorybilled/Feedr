package main

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	wr "github.com/mroth/weightedrand"
)

type bird struct {
	species species

	entranceTime     time.Time
	removalTime      time.Time
	eatingStartTime  time.Time
	eatingEndTime    time.Time
	singingStartTime time.Time

	physics   *birdPhysics
	animation *birdAnimation

	entering bool
	exiting  bool
	perched  bool
	eating   bool
	removed  bool
	singing  bool

	perch      *perch
	exitTarget pixel.Rect

	songControl *beep.Ctrl
	doneSinging chan bool

	// Birds will only be set to sing sometimes, regardless of the scheduled time.
	chosenToSing bool

	xPerchDistance float64
	yPerchDistance float64

	xExitDistance float64
	yExitDistance float64

	totalEntryMoves int
	totalExitMoves  int
}

func newBird(frameRate float64, species species, birdAnimations map[string][]pixel.Rect, animationSheet pixel.Picture, perch *perch, imd *imdraw.IMDraw) *bird {
	// Resolve spawn location and exit target location.
	spawnLocation, err := getOutsideLocation(species)
	exitTarget, err := getOutsideLocation(species)

	// Panic on any error here.
	if err != nil {
		panic("An error occured while creating a new bird: " + err.Error())
	}

	phys := &birdPhysics{
		rect:        spawnLocation,
		flightSpeed: species.FlightSpeed(),
	}

	anim := &birdAnimation{
		sheet: animationSheet,
		anims: birdAnimations,
		rate:  frameRate,
		imd:   imd,
	}

	newBird := &bird{species: species, entranceTime: time.Now(), physics: phys, animation: anim, perch: perch, exitTarget: exitTarget, doneSinging: make(chan bool)}

	// Bird should be set to 'entering' status.
	newBird.setEntranceStatus()

	// Set the time for the bird to be exited.
	newBird.setRemovalTime()

	// Set the next time the bird should eat.
	newBird.setEatingStartTime()

	// Set properties for entrance and exit moves.
	newBird.initEntryMoves()
	newBird.initExitMoves()

	return newBird
}

func (bird *bird) update(elapsed float64) {
	if bird.exiting || bird.entering {
		// Move the flying bird accordingly.
		nextX, nextY, adjustForFlightSpeed := bird.getNextFlightMove()
		bird.physics.update(elapsed, nextX, nextY, adjustForFlightSpeed)
	} else if bird.singing {
		// If singing has completed, allow the bird to assume a 'perched' status again.
		// And set the next singing time, as well as resetting the next eating time.
		select {
		case done := <-bird.doneSinging:
			if done {
				bird.stopSinging()
			}
		default:
		}

		if soundDisabled {
			bird.stopSinging()
		}
	} else if bird.eating {
		// If we've reached the time to stop eating OR seed ran out, revert back to simply 'perched' status.
		// And set the next eating start time.
		if time.Now().After(bird.eatingEndTime) || context.Seed().Finished {
			bird.setPerchedStatus()
			bird.setEatingStartTime()
		}
	} else if bird.perched {
		// Establish the next singing time if it has not yet been established.
		if bird.singingStartTime.IsZero() {
			bird.setSingingStartTime()
		}

		// Take care of events that can occur while perched.
		if time.Now().After(bird.singingStartTime) {
			if soundDisabled || !bird.chosenToSing {
				// Don't sing, reset singing time.
				bird.setSingingStartTime()
			} else {
				// If we've reached the next set singing time, have the bird start singing.
				bird.setSingingStatus()
				go playSound(soundBuffers[bird.species.Song()], bird.doneSinging)
			}
		} else if time.Now().After(bird.eatingStartTime) {
			// Bird can't eat if seed is finished. Eat later.
			if context.Seed().Finished {
				bird.setEatingStartTime()
			} else {
				// If we've reached the next set eating start time, and seed is available, have the bird start eating.
				// And set the eating end time.
				bird.setEatingStatus()
				bird.setEatingEndTime()
			}
		}
	}
}

func (bird *bird) getNextFlightMove() (X, Y float64, adjustForFlightSpeed bool) {
	if bird.exiting {
		// Calculate the exit location's coordinates
		emptySpaceOnBothSides := ((bird.exitTarget.Max.X - bird.exitTarget.Min.X) - bird.species.Width()) / 2
		targetMinY := bird.exitTarget.Min.Y
		targetMaxX := bird.exitTarget.Max.X - emptySpaceOnBothSides

		// Get the next move.
		return bird.nextExitMove(targetMinY, targetMaxX)
	} else if bird.entering {
		// Calculate the target perch coordinates.
		emptySpaceOnBothSides := ((bird.perch.X.max - bird.perch.X.min) - bird.species.Width()) / 2
		targetMinY := bird.perch.Y.min
		targetMaxX := bird.perch.X.max - emptySpaceOnBothSides

		// Get the next move.
		return bird.nextEntryMove(targetMinY, targetMaxX)
	}

	return 0, 0, true
}

func (bird *bird) setRemovedStatus() {
	fmt.Println("i am removed as a" + bird.species.Animation())

	bird.exiting = false
	bird.entering = false
	bird.eating = false
	bird.singing = false
	bird.perched = false
	bird.removed = true
}

func (bird *bird) setExitStatus() {
	fmt.Println("i am exiting as a" + bird.species.Animation())

	bird.exiting = true
	bird.entering = false
	bird.eating = false
	bird.singing = false
	bird.perched = false
	bird.removed = false
	bird.perch.occupied = false
}

func (bird *bird) setPerchedStatus() {
	fmt.Println("i am perched as a" + bird.species.Animation())

	bird.exiting = false
	bird.entering = false
	bird.eating = false
	bird.singing = false
	bird.perched = true
	bird.removed = false
	bird.perch.occupied = true
}

func (bird *bird) setEatingStatus() {
	bird.exiting = false
	bird.entering = false
	bird.eating = true
	bird.singing = false
	bird.perched = true
	bird.removed = false
	bird.perch.occupied = true
}

func (bird *bird) setEntranceStatus() {
	fmt.Println("i am entering as a" + bird.species.Animation())
	bird.exiting = false
	bird.entering = true
	bird.eating = false
	bird.singing = false
	bird.perched = false
	bird.removed = false
	bird.perch.occupied = true
}

func (bird *bird) setSingingStatus() {
	bird.exiting = false
	bird.entering = false
	bird.eating = false
	bird.singing = true
	bird.perched = true
	bird.removed = false
	bird.perch.occupied = true
}

func resolveNewLengthInSeconds(timeMap map[timeLength]pair) (int, error) {
	// Enumerate all time length category choices.
	choices := []wr.Choice{}
	for category, likelihood := range context.TimeLikelihoods() {
		choices = append(choices, wr.Choice{Item: category, Weight: likelihood})
	}

	// Initialize a weighted probability time length chooser.
	chooser, _ := wr.NewChooser(choices...)

	// Pick a random time length category.
	lengthCategory := chooser.Pick().(timeLength)

	// Depending on the time length category picked, pick a random length of between its time length range.
	switch lengthCategory {
	case veryShort:
		minMaxPair := timeMap[veryShort]
		return nextRandomInt(minMaxPair.min, minMaxPair.max), nil
	case short:
		minMaxPair := timeMap[short]
		return nextRandomInt(minMaxPair.min, minMaxPair.max), nil
	case medium:
		minMaxPair := timeMap[medium]
		return nextRandomInt(minMaxPair.min, minMaxPair.max), nil
	case long:
		minMaxPair := timeMap[long]
		return nextRandomInt(minMaxPair.min, minMaxPair.max), nil
	case veryLong:
		minMaxPair := timeMap[veryLong]
		return nextRandomInt(minMaxPair.min, minMaxPair.max), nil
	case insane:
		minMaxPair := timeMap[insane]
		return nextRandomInt(minMaxPair.min, minMaxPair.max), nil
	}

	return 0, errors.New("Invalid feeder length category selected")
}

func getOutsideLocation(species species) (pixel.Rect, error) {
	// Randomly choose from one of four 'areas' to find a location.
	random := nextRandomInt(0, 3)

	switch random {
	case 0:
		// Left entry.
		maxMinXLocation := -(winWidth / 2) - species.Width()
		leftMinX := maxMinXLocation
		leftMinY := nextRandomFloat64(-(winHeight/2)-spawnRandomnessOffset, (winHeight/2)+spawnRandomnessOffset)
		return pixel.R(leftMinX, leftMinY, leftMinX+species.Width(), leftMinY+species.Height()), nil
	case 1:
		// Right entry.
		minMaxXLocation := (winWidth / 2) + species.Width()
		rightMaxX := minMaxXLocation
		rightMinY := nextRandomFloat64(-(winHeight/2)-spawnRandomnessOffset, (winHeight/2)+spawnRandomnessOffset)
		return pixel.R(rightMaxX-species.Width(), rightMinY, rightMaxX, rightMinY+species.Height()), nil
	case 2:
		// Top entry.
		minMaxYLocation := (winHeight / 2) + species.Height()
		topMaxY := minMaxYLocation
		topMinX := nextRandomFloat64(-(winWidth/2)-spawnRandomnessOffset, (winWidth/2)+spawnRandomnessOffset)
		return pixel.R(topMinX, topMaxY-species.Height(), topMinX+species.Width(), topMaxY), nil
	case 3:
		// Bottom entry.
		maxMinYLocation := -(winHeight / 2) - species.Height()
		bottomMinY := maxMinYLocation
		bottomMinX := nextRandomFloat64(-(winWidth/2)-spawnRandomnessOffset, (winWidth/2)+spawnRandomnessOffset)
		return pixel.R(bottomMinX, bottomMinY, bottomMinX+species.Width(), bottomMinY+species.Height()), nil
	}

	return pixel.Rect{}, errors.New("no outside location could be determined for new bird")
}

func (bird *bird) setRemovalTime() {
	feederLength, err := resolveNewLengthInSeconds(context.FeederLengthRanges())
	if err != nil {
		feederLength = defaultFeederLength
	}

	bird.removalTime = time.Now().Add(time.Second * time.Duration(feederLength))
}

func (bird *bird) setEatingStartTime() {
	eatingGap, err := resolveNewLengthInSeconds(context.EatingGapRanges())
	if err != nil {
		eatingGap = defaultEatingGap
	}

	bird.eatingStartTime = time.Now().Add(time.Second * time.Duration(eatingGap))
}

func (bird *bird) setEatingEndTime() {
	eatingLength, err := resolveNewLengthInSeconds(context.EatingLengthRanges())
	if err != nil {
		eatingLength = defaultEatingLength
	}

	bird.eatingEndTime = time.Now().Add(time.Second * time.Duration(eatingLength))
}

func (bird *bird) setSingingStartTime() {
	singingGap, err := resolveNewLengthInSeconds(context.SingingGapRanges())
	if err != nil {
		singingGap = defaultSingingGap
	}

	bird.singingStartTime = time.Now().Add(time.Second * time.Duration(singingGap))

	// Create two choices: true to sing, false to not sing.
	choices := []wr.Choice{}
	choices = append(choices, wr.Choice{Item: true, Weight: bird.species.SingingLikelihood()})
	choices = append(choices, wr.Choice{Item: false, Weight: likelihoodMaxPercent - bird.species.SingingLikelihood()})

	// Initialize a weighted probability boolean chooser.
	chooser, _ := wr.NewChooser(choices...)

	// Pick a random to-sing-or-not-to-sing result.
	bird.chosenToSing = chooser.Pick().(bool)
}

func (bird *bird) stopSinging() {
	bird.setPerchedStatus()
	bird.setSingingStartTime()
	bird.setEatingStartTime()
}

func (bird *bird) nextEntryMove(targetMinY, targetMaxX float64) (float64, float64, bool) {
	currentMinY := bird.physics.rect.Min.Y
	currentMaxX := bird.physics.rect.Max.X

	if currentMinY == targetMinY && currentMaxX == targetMaxX {
		// If the bird has finished its entry moves, set it as 'perched' and end movement.
		bird.setPerchedStatus()

		// Allow directional changes now that we're perched.
		bird.animation.UnlockDirection()
		return 0, 0, true
	}

	xMove := bird.xPerchDistance / float64(bird.totalEntryMoves)
	yMove := bird.yPerchDistance / float64(bird.totalEntryMoves)

	xDifference := targetMaxX - currentMaxX
	yDifference := targetMinY - currentMinY

	adjustForFlightSpeed := true
	if math.Abs(xDifference) <= math.Abs(xMove) && math.Abs(yDifference) <= math.Abs(yMove) {
		adjustForFlightSpeed = false

		xMove = xDifference
		yMove = yDifference

		// Forcing the move direction into place may alter direction abruptly. We prevent this here.
		bird.animation.LockDirection()
	}

	return xMove, yMove, adjustForFlightSpeed
}

func (bird *bird) nextExitMove(targetMinY, targetMaxX float64) (float64, float64, bool) {
	currentMinY := bird.physics.rect.Min.Y
	currentMaxX := bird.physics.rect.Max.X

	if currentMinY == targetMinY && currentMaxX == targetMaxX {
		// If the bird has finished its entry moves, set it as 'perched' and end movement.
		bird.setRemovedStatus()
		return 0, 0, true
	}

	xMove := bird.xExitDistance / float64(bird.totalExitMoves)
	yMove := bird.yExitDistance / float64(bird.totalExitMoves)

	xDifference := targetMaxX - currentMaxX
	yDifference := targetMinY - currentMinY

	adjustForFlightSpeed := true
	if math.Abs(xDifference) <= 1 && math.Abs(yDifference) <= 1 {
		adjustForFlightSpeed = false
		xMove = xDifference
		yMove = yDifference
	}

	return xMove, yMove, adjustForFlightSpeed
}

func (bird *bird) initEntryMoves() {
	currentMinY := bird.physics.rect.Min.Y
	currentMaxX := bird.physics.rect.Max.X

	emptySpaceOnBothSides := ((bird.perch.X.max - bird.perch.X.min) - bird.species.Width()) / 2
	targetMinY := bird.perch.Y.min
	targetMaxX := bird.perch.X.max - emptySpaceOnBothSides

	bird.yPerchDistance = targetMinY - currentMinY
	bird.xPerchDistance = targetMaxX - currentMaxX
	bird.totalEntryMoves = int(math.Round(math.Max(math.Abs(bird.xPerchDistance), math.Abs(bird.yPerchDistance)) / 10))
}

func (bird *bird) initExitMoves() {
	exitEmptySpaceOnBothSides := ((bird.exitTarget.Max.X - bird.exitTarget.Min.X) - bird.species.Width()) / 2
	exitTargetMinY := bird.exitTarget.Min.Y
	exitTargetMaxX := bird.exitTarget.Max.X - exitEmptySpaceOnBothSides

	bird.yExitDistance = exitTargetMinY - bird.perch.Y.min
	bird.xExitDistance = exitTargetMaxX - bird.perch.X.max
	bird.totalExitMoves = int(math.Round(math.Max(math.Abs(bird.xExitDistance), math.Abs(bird.yExitDistance)) / 10))
}
