package main

import (
	"math"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
)

type birdAnimation struct {
	sheet pixel.Picture
	anims map[string][]pixel.Rect
	rate  float64

	state         animState
	counter       float64
	direction     float64
	lockDirection bool

	frame pixel.Rect

	sprite *pixel.Sprite
	imd    *imdraw.IMDraw
}

func (animation *birdAnimation) LockDirection() {
	animation.lockDirection = true
}

func (animation *birdAnimation) UnlockDirection() {
	animation.lockDirection = false
}

func (animation *birdAnimation) update(elapsed float64, bird *bird) {
	animation.counter += elapsed

	// determine the new animation state
	var newState animState
	switch {
	case bird.singing:
		newState = singing
	case bird.eating:
		newState = eating
	case bird.perched:
		newState = perched
	case bird.entering || bird.exiting:
		newState = flying
	}

	// reset the time counter if the state changed
	if animation.state != newState {
		animation.state = newState
		animation.counter = 0
	}

	// determine the correct animation frame
	switch animation.state {
	case perched:
		animation.frame = animation.anims["Perch"][0]
	case flying:
		i := int(math.Floor(animation.counter / animation.rate))
		animation.frame = animation.anims["Fly"][i%len(animation.anims["Fly"])]
	case eating:
		i := int(math.Floor(animation.counter / animation.rate))
		animation.frame = animation.anims["Eat"][i%len(animation.anims["Eat"])]
	case singing:
		i := int(math.Floor(animation.counter / animation.rate))
		animation.frame = animation.anims["Sing"][i%len(animation.anims["Sing"])]
	}

	// set the facing direction of the bird
	if bird.physics.vel.X != 0 && !animation.lockDirection {
		if bird.physics.vel.X > 0 {
			animation.direction = -1
		} else {
			animation.direction = +1
		}
	}
}

func (animation *birdAnimation) draw(phys *birdPhysics) {
	if animation.sprite == nil {
		animation.sprite = pixel.NewSprite(nil, pixel.Rect{})
	}
	// draw the correct frame with the correct position and direction
	animation.sprite.Set(animation.sheet, animation.frame)
	animation.sprite.Draw(animation.imd, pixel.IM.
		ScaledXY(pixel.ZV, pixel.V(
			phys.rect.W()/animation.sprite.Frame().W(),
			phys.rect.H()/animation.sprite.Frame().H(),
		)).
		ScaledXY(pixel.ZV, pixel.V(-animation.direction, 1)).
		Moved(phys.rect.Center()),
	)
}
