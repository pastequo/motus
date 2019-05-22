// Package motus contains the entrypoints for the project.
package motus

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"sync"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/gobuffalo/packr/v2"
	"github.com/logrusorgru/aurora"
	"github.com/pkg/errors"
)

// ErrInvalidArg is returned if the arguments are invalid.
var ErrInvalidArg = errors.New("invalid argument")

type position int8

const (
	ok position = iota
	outOfPlace
	out
)

// DisplayText displays text as it should be in motus (lingo) game.
func DisplayText(txt string, okCount, outOfPlaceCount int) error {

	if txt == "" {
		return nil
	}

	if okCount < 0 {
		okCount = 0
	}
	if outOfPlaceCount < 0 {
		outOfPlaceCount = 0
	}

	if len(txt) < okCount+outOfPlaceCount {
		return ErrInvalidArg
	}

	mask := make([]position, len(txt))
	for i := 0; i < len(txt); i++ {
		switch {
		case i < okCount:
			mask[i] = ok
		case i < okCount+outOfPlaceCount:
			mask[i] = outOfPlace
		default:
			mask[i] = out

		}
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(mask), func(i, j int) {
		if i*j == 0 {
			return
		}
		mask[i], mask[j] = mask[j], mask[i]
	})

	err := loadBuffer()
	if err != nil {
		return err
	}

	fmt.Printf("%v\r", aurora.White(txt).BgBlue())

	var d aurora.Value

	for i, c := range txt {
		switch mask[i] {
		case ok:
			playSound(bufferOK)
			d = aurora.White(string(c)).BgRed()
		case outOfPlace:
			playSound(bufferOOP)
			d = aurora.White(string(c)).BgYellow()
		case out:
			playSound(bufferKO)
			d = aurora.White(string(c)).BgBlue()
		}
		fmt.Printf("%v", d)
		time.Sleep(50 * time.Millisecond)
	}

	fmt.Printf("\n")

	return nil
}

var (
	once sync.Once
	box  *packr.Box
)

var (
	bufferOK  *beep.Buffer
	bufferOOP *beep.Buffer
	bufferKO  *beep.Buffer
)

func loadBuffer() error {
	var err error

	bufferOK, err = initBuffer("ok.mp3", true)
	if err != nil {
		return errors.Wrap(err, "failed to load ok.mp3")
	}

	bufferOOP, err = initBuffer("oop.mp3", false)
	if err != nil {
		return errors.Wrap(err, "failed to load oop.mp3")
	}

	bufferKO, err = initBuffer("ko.mp3", false)
	if err != nil {
		return errors.Wrap(err, "failed to load ko.mp3")
	}

	return nil
}

func initBuffer(filePath string, initSpeaker bool) (*beep.Buffer, error) {

	once.Do(func() {
		box = packr.New("sound", "./resources")
	})

	b, err := box.Find(filePath)
	if err != nil {
		return nil, err
	}

	readCloser := ioutil.NopCloser(bytes.NewReader(b))
	streamer, format, err := mp3.Decode(readCloser)
	if err != nil {
		return nil, err
	}
	defer streamer.Close()

	ret := beep.NewBuffer(format)
	ret.Append(streamer)

	if !initSpeaker {
		return ret, nil
	}

	err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/100))

	return ret, err
}

func playSound(buffer *beep.Buffer) {

	done := make(chan bool)

	sound := buffer.Streamer(0, buffer.Len())
	speaker.Play(beep.Seq(sound, beep.Callback(func() {
		done <- true
	})))

	<-done
}
