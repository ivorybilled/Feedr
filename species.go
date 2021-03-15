package main

type species interface {
	Initialize()
	ConsumptionRate() float64
	Width() float64
	Height() float64
	FlightSpeed() float64
	Song() string
	SingingLikelihood() uint
	Animation() string
}

// Northern Cardinal species.
type cardinal struct {
	consumptionRate   float64
	width             float64
	height            float64
	flightSpeed       float64
	songName          string
	singingLikelihood uint
	animation         string
}

func (cardinal *cardinal) Initialize() {
	cardinal.consumptionRate = .002
	cardinal.width = 190
	cardinal.height = 221
	cardinal.flightSpeed = 30
	cardinal.songName = "Downy Woodpecker"
	cardinal.singingLikelihood = 100
	cardinal.animation = "sprites/northernCardinal.png"
}

func (cardinal *cardinal) ConsumptionRate() float64 {
	return cardinal.consumptionRate
}

func (cardinal *cardinal) Width() float64 {
	return cardinal.width
}

func (cardinal *cardinal) Height() float64 {
	return cardinal.height
}

func (cardinal *cardinal) FlightSpeed() float64 {
	return cardinal.flightSpeed
}

func (cardinal *cardinal) Animation() string {
	return cardinal.animation
}

func (cardinal *cardinal) Song() string {
	return cardinal.songName
}

func (cardinal *cardinal) SingingLikelihood() uint {
	return cardinal.singingLikelihood
}

// Downy Woodpecker species.
type downyWoodpecker struct {
	consumptionRate   float64
	width             float64
	height            float64
	flightSpeed       float64
	songName          string
	singingLikelihood uint
	animation         string
}

func (downyWoodpecker *downyWoodpecker) Initialize() {
	downyWoodpecker.consumptionRate = .003
	downyWoodpecker.width = 190
	downyWoodpecker.height = 221
	downyWoodpecker.flightSpeed = 30
	downyWoodpecker.songName = "Downy Woodpecker"
	downyWoodpecker.singingLikelihood = 30
	downyWoodpecker.animation = "sprites/downyWoodpecker.png"
}

func (downyWoodpecker *downyWoodpecker) ConsumptionRate() float64 {
	return downyWoodpecker.consumptionRate
}

func (downyWoodpecker *downyWoodpecker) Width() float64 {
	return downyWoodpecker.width
}

func (downyWoodpecker *downyWoodpecker) Height() float64 {
	return downyWoodpecker.height
}

func (downyWoodpecker *downyWoodpecker) FlightSpeed() float64 {
	return downyWoodpecker.flightSpeed
}

func (downyWoodpecker *downyWoodpecker) Song() string {
	return downyWoodpecker.songName
}

func (downyWoodpecker *downyWoodpecker) Animation() string {
	return downyWoodpecker.animation
}

func (downyWoodpecker *downyWoodpecker) SingingLikelihood() uint {
	return downyWoodpecker.singingLikelihood
}

// Black-capped Chickadee species.
type chickadee struct {
	consumptionRate   float64
	width             float64
	height            float64
	flightSpeed       float64
	songName          string
	singingLikelihood uint
	animation         string
}

func (chickadee *chickadee) Initialize() {
	chickadee.consumptionRate = .005
	chickadee.width = 190
	chickadee.height = 221
	chickadee.flightSpeed = 30
	chickadee.songName = "Downy Woodpecker"
	chickadee.singingLikelihood = 100
	chickadee.animation = "sprites/blackCappedChickadee.png"
}

func (chickadee *chickadee) ConsumptionRate() float64 {
	return chickadee.consumptionRate
}

func (chickadee *chickadee) Width() float64 {
	return chickadee.width
}

func (chickadee *chickadee) Height() float64 {
	return chickadee.height
}

func (chickadee *chickadee) FlightSpeed() float64 {
	return chickadee.flightSpeed
}

func (chickadee *chickadee) Song() string {
	return chickadee.songName
}

func (chickadee *chickadee) Animation() string {
	return chickadee.animation
}

func (chickadee *chickadee) SingingLikelihood() uint {
	return chickadee.singingLikelihood
}

// Tufted Titmouse species.
type titmouse struct {
	consumptionRate   float64
	width             float64
	height            float64
	flightSpeed       float64
	songName          string
	singingLikelihood uint
	animation         string
}

func (titmouse *titmouse) Initialize() {
	titmouse.consumptionRate = .002
	titmouse.width = 190
	titmouse.height = 221
	titmouse.flightSpeed = 30
	titmouse.songName = "Downy Woodpecker"
	titmouse.singingLikelihood = 80
	titmouse.animation = "sprites/tuftedTitmouse.png"
}

func (titmouse *titmouse) ConsumptionRate() float64 {
	return titmouse.consumptionRate
}

func (titmouse *titmouse) Width() float64 {
	return titmouse.width
}

func (titmouse *titmouse) Height() float64 {
	return titmouse.height
}

func (titmouse *titmouse) FlightSpeed() float64 {
	return titmouse.flightSpeed
}

func (titmouse *titmouse) Song() string {
	return titmouse.songName
}

func (titmouse *titmouse) Animation() string {
	return titmouse.animation
}

func (titmouse *titmouse) SingingLikelihood() uint {
	return titmouse.singingLikelihood
}
