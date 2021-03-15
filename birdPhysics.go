package main

import "github.com/faiface/pixel"

type birdPhysics struct {
	rect        pixel.Rect
	vel         pixel.Vec
	flightSpeed float64
}

func (physics *birdPhysics) update(elapsed, x, y float64, adjustForFlightSpeed bool) {
	flightSpeed := physics.flightSpeed

	if !adjustForFlightSpeed {
		elapsed = 1
		flightSpeed = 1
	}

	// apply flight script movements
	physics.vel.X = x * flightSpeed
	physics.vel.Y = y * flightSpeed

	// apply velocity
	physics.rect = physics.rect.Moved(physics.vel.Scaled(elapsed))
}
