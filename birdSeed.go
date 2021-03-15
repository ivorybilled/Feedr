package main

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
)

type birdSeed struct {
	Finished bool

	center pixel.Vec

	height            float64
	width             float64
	seedCount         float64
	originalSeedCount float64
	seedsPerRow       float64

	// How much to scale the seed in terms of size.
	scaleX float64
	scaleY float64

	// Adjust the seed's positioning.
	adjustedX float64
	adjustedY float64

	// The Y coordinate to reach (progressing from low to high) for the seed to be 'done'.
	doneLowerY float64
}

func (seed *birdSeed) getRemainingSeedCount(birds []*bird) float64 {
	newSeedCount := seed.seedCount

	for _, bird := range birds {
		if bird.eating {
			newSeedCount -= bird.species.ConsumptionRate()
		}
	}

	return newSeedCount
}

func (seed *birdSeed) draw(imd *imdraw.IMDraw, birds []*bird, picture pixel.Picture) {
	seed.seedCount = seed.getRemainingSeedCount(birds)

	// Don't let seed count dip into negatives.
	if seed.seedCount <= 0 {
		seed.seedCount = 0
	}

	// Deplete the feeder at the given intervals.
	seed.height = seed.seedCount / seed.seedsPerRow

	// Set the top and bottom Y coordinates for the seed pile rectangle.
	upperY := seed.center.Y + (seed.originalSeedCount/seed.seedsPerRow)/2
	lowerY := upperY - seed.height

	// Ensure bird can be aware when seed is finished.
	if lowerY >= seed.doneLowerY {
		seed.Finished = true
	} else {
		seed.Finished = false
	}

	// Top right point.
	maxVec := pixel.Vec{X: seed.center.X + seed.width/2, Y: upperY}

	// Bottom left point.
	minVec := pixel.Vec{X: seed.center.X - seed.width/2, Y: lowerY}

	seedSprite := pixel.NewSprite(nil, pixel.Rect{})
	seedSprite.Set(picture, pixel.Rect{Min: minVec, Max: maxVec})

	// Draw the seeds sprite.
	seedSprite.Draw(imd, pixel.IM.ScaledXY(pixel.ZV, pixel.V(seed.scaleX, seed.scaleY)).Moved(pixel.Vec{X: seed.adjustedX, Y: seed.adjustedY}))
}

func (seed *birdSeed) refill() {
	incrementAmount := seed.originalSeedCount * defaultSeedRefillMultiplier

	if seed.seedCount+incrementAmount > seed.originalSeedCount {
		incrementAmount = seed.originalSeedCount - seed.seedCount
	}

	seed.seedCount += incrementAmount
}

func newBirdSeed(center pixel.Vec, height, width, seedCount, scaleX, scaleY, adjustedX, adjustedY, doneLowerY float64) *birdSeed {
	seed := &birdSeed{
		center:            center,
		width:             width,
		height:            height,
		seedCount:         seedCount,
		originalSeedCount: seedCount,
		seedsPerRow:       seedCount / height,
		scaleX:            scaleX,
		scaleY:            scaleY,
		adjustedX:         adjustedX,
		adjustedY:         adjustedY,
		doneLowerY:        doneLowerY,
	}

	return seed
}
