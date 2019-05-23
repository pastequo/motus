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

// ErrNoSound is returned if the method to play the sound timeouted once.
var ErrNoSound = errors.New("failed to play sound")

type position int8

const (
	ok position = iota
	outOfPlace
	out
)

// DeWinter is the type to use to display text, motus (lingo) style.
// It has a maxTimeout member that refers to the maxDuration to wait to play a single sound.
type DeWinter struct {
	maxTimeout time.Duration
	withSound  bool
}

// NewDeWinter creates a DeWinter struct that plays sound
func NewDeWinter(maxTimeout time.Duration) *DeWinter {
	return &DeWinter{maxTimeout: maxTimeout, withSound: true}
}

// NewMutedDeWinter creates a DeWinter struct that is muted
func NewMutedDeWinter() *DeWinter {
	return &DeWinter{}
}

// DisplayText displays text as it should be in motus (lingo) game.
// If the DeWinter timeout is reached once, ErrNoSound is returned and the sound is disable for this displayer.
func (d *DeWinter) DisplayText(txt string, okCount, outOfPlaceCount int) error {

	var ret error = nil

	if txt == "" {
		return ret
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

	var buf *beep.Buffer
	var bgFunc func(arg interface{}) aurora.Value

	for i, c := range txt {
		switch mask[i] {
		case ok:
			buf = bufferOK
			bgFunc = aurora.BgRed
		case outOfPlace:
			buf = bufferOOP
			bgFunc = aurora.BgYellow
		case out:
			buf = bufferKO
			bgFunc = aurora.BgBlue
		}

		if d.withSound {
			err = playSound(buf, d.maxTimeout)
			if err != nil {
				d.withSound = false
				ret = ErrNoSound
			}
		}

		displayChar := bgFunc(aurora.White(string(c)))
		fmt.Printf("%v", displayChar)
		time.Sleep(50 * time.Millisecond)
	}

	fmt.Printf("\n")

	return ret
}

// IsMuted returns true is DeWinter can't play sound.
func (d *DeWinter) IsMuted() bool {
	return !d.withSound
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

func playSound(buffer *beep.Buffer, timeout time.Duration) error {

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	done := make(chan bool)

	sound := buffer.Streamer(0, buffer.Len())
	speaker.Play(beep.Seq(sound, beep.Callback(func() {
		done <- true
	})))

	select {
	case <-done:
		return nil
	case <-timer.C:
		return ErrNoSound
	}
}
