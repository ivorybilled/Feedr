package main

import (
	"github.com/faiface/pixel"
)

// I understand making an interface for what are essentially static properties is not best practice.
// But the polymorphism that this provides is currently needed.
type feederContext interface {
	Name() string
	Initialize()
	TimeLikelihoods() map[timeLength]uint
	SpawnLengthRanges() map[timeLength]pair
	FeederLengthRanges() map[timeLength]pair
	EatingLengthRanges() map[timeLength]pair
	EatingGapRanges() map[timeLength]pair
	SingingGapRanges() map[timeLength]pair
	BirdLikelihoods() map[species]uint
	Perches() []*perch
	Seed() *birdSeed
	Sounds() map[string]string
	Backgrounds() map[string]pixel.Picture
	Seeds() map[string]pixel.Picture
}

type standardHouseFeeder struct {
	timeLikelihoods    map[timeLength]uint
	spawnLengthRanges  map[timeLength]pair
	feederLengthRanges map[timeLength]pair
	eatingLengthRanges map[timeLength]pair
	eatingGapRanges    map[timeLength]pair
	singingGapRanges   map[timeLength]pair
	birdLikelihoods    map[species]uint
	perches            []*perch
	seed               *birdSeed
	sounds             map[string]string
	backgrounds        map[string]pixel.Picture
	seeds              map[string]pixel.Picture
}

func (standardHouseFeeder *standardHouseFeeder) Initialize() {
	standardHouseFeeder.timeLikelihoods = map[timeLength]uint{
		veryShort: 50,
		short:     400,
		medium:    400,
		long:      144,
		veryLong:  5,
		insane:    1,
	}

	standardHouseFeeder.spawnLengthRanges = map[timeLength]pair{
		veryShort: {5, 30},
		short:     {30, 120},
		medium:    {120, 600},
		long:      {600, 3000},
		veryLong:  {3000, 7000},
		insane:    {7000, 15000},
	}

	standardHouseFeeder.feederLengthRanges = map[timeLength]pair{
		veryShort: {10, 20},
		short:     {20, 90},
		medium:    {90, 200},
		long:      {200, 400},
		veryLong:  {400, 700},
		insane:    {700, 2000},
	}

	standardHouseFeeder.eatingLengthRanges = map[timeLength]pair{
		veryShort: {1, 3},
		short:     {3, 5},
		medium:    {5, 7},
		long:      {7, 9},
		veryLong:  {9, 15},
		insane:    {15, 20},
	}

	standardHouseFeeder.eatingGapRanges = map[timeLength]pair{
		veryShort: {5, 10},
		short:     {10, 15},
		medium:    {15, 25},
		long:      {25, 35},
		veryLong:  {35, 80},
		insane:    {80, 150},
	}

	standardHouseFeeder.singingGapRanges = map[timeLength]pair{
		veryShort: {45, 80},
		short:     {80, 100},
		medium:    {120, 240},
		long:      {250, 350},
		veryLong:  {350, 500},
		insane:    {500, 1000},
	}

	standardHouseFeeder.birdLikelihoods = map[species]uint{
		&cardinal{}:        120,
		&downyWoodpecker{}: 300,
		&chickadee{}:       300,
		&titmouse{}:        280,
	}

	standardHouseFeeder.perches = []*perch{
		{X: &coordinatePair{90, 280}, Y: &coordinatePair{-340, -119}, occupied: false},
		{X: &coordinatePair{270, 460}, Y: &coordinatePair{-350, -129}, occupied: false},
	}

	standardHouseFeeder.seed = newBirdSeed(pixel.V(0, 0), 300, 700, 3000, 1.15, 1.15, 100, -300, 100)

	standardHouseFeeder.sounds = map[string]string{
		"Black-capped Chickadee": "sounds/songs/downyWoodpeckerSong.mp3",
		"Tufted Titmouse":        "sounds/songs/downyWoodpeckerSong.mp3",
		"Northern Cardinal":      "sounds/songs/downyWoodpeckerSong.mp3",
		"Downy Woodpecker":       "sounds/songs/downyWoodpeckerSong.mp3",
		"Background":             "sounds/background/ambience.mp3",
	}

	standardHouseFeeder.backgrounds = map[string]pixel.Picture{
		"night": getPixelPicture("sprites/backgrounds/backyardSunflowerNightEmpty.png"),
		"day":   getPixelPicture("sprites/backgrounds/backyardSunflowerDayEmpty.png"),
		"dusk":  getPixelPicture("sprites/backgrounds/backyardSunflowerDuskEmpty.png"),
	}

	standardHouseFeeder.seeds = map[string]pixel.Picture{
		"night": getPixelPicture("sprites/seeds/sunflowerSeedPileNight.png"),
		"day":   getPixelPicture("sprites/seeds/sunflowerSeedPileDay.png"),
		"dusk":  getPixelPicture("sprites/seeds/sunflowerSeedPileDusk.png"),
	}
}

func (standardHouseFeeder *standardHouseFeeder) Name() string {
	return standardHouseFeederName
}

func (standardHouseFeeder *standardHouseFeeder) TimeLikelihoods() map[timeLength]uint {
	return standardHouseFeeder.timeLikelihoods
}

func (standardHouseFeeder *standardHouseFeeder) SpawnLengthRanges() map[timeLength]pair {
	return standardHouseFeeder.spawnLengthRanges
}

func (standardHouseFeeder *standardHouseFeeder) FeederLengthRanges() map[timeLength]pair {
	return standardHouseFeeder.feederLengthRanges
}

func (standardHouseFeeder *standardHouseFeeder) EatingLengthRanges() map[timeLength]pair {
	return standardHouseFeeder.eatingLengthRanges
}

func (standardHouseFeeder *standardHouseFeeder) EatingGapRanges() map[timeLength]pair {
	return standardHouseFeeder.eatingGapRanges
}

func (standardHouseFeeder *standardHouseFeeder) SingingGapRanges() map[timeLength]pair {
	return standardHouseFeeder.singingGapRanges
}

func (standardHouseFeeder *standardHouseFeeder) BirdLikelihoods() map[species]uint {
	return standardHouseFeeder.birdLikelihoods
}

func (standardHouseFeeder *standardHouseFeeder) Perches() []*perch {
	return standardHouseFeeder.perches
}

func (standardHouseFeeder *standardHouseFeeder) Seed() *birdSeed {
	return standardHouseFeeder.seed
}

func (standardHouseFeeder *standardHouseFeeder) Sounds() map[string]string {
	return standardHouseFeeder.sounds
}

func (standardHouseFeeder *standardHouseFeeder) Backgrounds() map[string]pixel.Picture {
	return standardHouseFeeder.backgrounds
}

func (standardHouseFeeder *standardHouseFeeder) Seeds() map[string]pixel.Picture {
	return standardHouseFeeder.seeds
}
