package main

import (
	"log"
	"os"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

// Call as a go-routine.
func playSound(buffer *beep.Buffer, completeFlag chan bool) {
	streamer := buffer.Streamer(0, buffer.Len())

	done := make(chan bool)

	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		completeFlag <- true
		done <- true
	})))

	<-done
}

// Call as a go-routine.
func playLoopingSound(buffer *beep.Buffer, controller chan bool) {
	for {
		// Loop while sound is disabled. Exit if controller is closed.
		if soundDisabled {
			select {
			case _, ok := <-controller:
				if !ok {
					return
				}
			default:
			}

			continue
		}

		streamer := buffer.Streamer(0, buffer.Len())

		done := make(chan bool)

		speaker.Play(beep.Seq(streamer, beep.Callback(func() {
			done <- true
		})))

	playback:
		for {
			select {
			case _, ok := <-controller:
				if !ok {
					return
				}
			case finished, ok := <-done:
				if finished || !ok {
					break playback
				}
			default:
			}

			if soundDisabled {
				break playback
			}
		}
	}
}

func bufferSound(path string) *beep.Buffer {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		log.Fatal(err)
	}

	buffer := beep.NewBuffer(format)
	buffer.Append(streamer)
	streamer.Close()

	return buffer
}
