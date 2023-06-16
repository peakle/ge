package ge

import (
	"io"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	resource "github.com/quasilyte/ebitengine-resource"
)

type AudioSystem struct {
	loader *resource.Loader

	currentMusic resource.Audio

	audioContext *audio.Context

	currentQueueSound resource.Audio
	soundQueue        []resource.AudioID

	groupVolume [4]float64

	muted bool
}

type audioResource struct {
	player *audio.Player
	group  uint
	volume float64
}

func (sys *AudioSystem) init(audioContext *audio.Context, l *resource.Loader) {
	sys.loader = l
	sys.audioContext = audioContext
	sys.soundQueue = make([]resource.AudioID, 0, 4)

	for i := range sys.groupVolume {
		sys.groupVolume[i] = 1.0
	}

	// Audio player factory has lazy initialization that may lead
	// to a ~0.2s delay before the first sound can be played.
	// To avoid that delay, we force that factory to initialize
	// right now, before the game is started.
	dummy := sys.audioContext.NewPlayerFromBytes(nil)
	dummy.Rewind()
}

func (sys *AudioSystem) GetContext() *audio.Context {
	return sys.audioContext
}

func (sys *AudioSystem) Update() {
	if sys.currentQueueSound.Player == nil {
		if len(sys.soundQueue) == 0 {
			// Nothing to play in the queue.
			return
		}
		// Do a dequeue.
		sys.currentQueueSound = sys.playSound(sys.soundQueue[0])
		for i, id := range sys.soundQueue[1:] {
			sys.soundQueue[i] = id
		}
		sys.soundQueue = sys.soundQueue[:len(sys.soundQueue)-1]
		return
	}
	if !sys.currentQueueSound.Player.IsPlaying() {
		// Finished playing the current enqueued sound.
		sys.currentQueueSound = resource.Audio{}
	}
}

func (sys *AudioSystem) SetGroupVolume(groupID uint, multiplier float64) {
	if groupID >= uint(len(sys.groupVolume)) {
		panic("invalid group ID")
	}
	sys.groupVolume[groupID] = multiplier
}

func (sys *AudioSystem) DecodeWAV(r io.Reader) (*wav.Stream, error) {
	return wav.Decode(sys.audioContext, r)
}

func (sys *AudioSystem) DecodeOGG(r io.Reader) (*vorbis.Stream, error) {
	return vorbis.Decode(sys.audioContext, r)
}

func (sys *AudioSystem) PauseCurrentMusic() {
	if sys.muted {
		return
	}
	if sys.currentMusic.Player == nil {
		return
	}
	sys.currentMusic.Player.Pause()
}

func (sys *AudioSystem) ContinueCurrentMusic() {
	if sys.muted {
		return
	}
	if sys.currentMusic.Player == nil || sys.currentMusic.Player.IsPlaying() {
		return
	}
	sys.currentMusic.Player.SetVolume(sys.currentMusic.Volume * sys.groupVolume[sys.currentMusic.Group])
	sys.currentMusic.Player.Play()
}

func (sys *AudioSystem) ContinueMusic(id resource.AudioID) {
	if sys.muted {
		return
	}
	sys.continueMusic(sys.loader.LoadAudio(id))
}

func (sys *AudioSystem) continueMusic(res resource.Audio) {
	if res.Player.IsPlaying() {
		return
	}
	sys.currentMusic = res
	res.Player.SetVolume(res.Volume * sys.groupVolume[res.Group])
	res.Player.Play()
}

func (sys *AudioSystem) PlayMusic(id resource.AudioID) {
	if sys.muted {
		return
	}
	res := sys.loader.LoadOGG(id)
	if sys.currentMusic.Player != nil && res.Player == sys.currentMusic.Player && res.Player.IsPlaying() {
		return
	}
	sys.currentMusic = res
	res.Player.SetVolume(res.Volume * sys.groupVolume[res.Group])
	res.Player.Rewind()
	res.Player.Play()
}

func (sys *AudioSystem) ResetQueue() {
	sys.soundQueue = sys.soundQueue[:0]
}

func (sys *AudioSystem) EnqueueSound(id resource.AudioID) {
	if sys.muted {
		return
	}
	sys.soundQueue = append(sys.soundQueue, id)
}

func (sys *AudioSystem) PlaySound(id resource.AudioID) {
	if sys.muted {
		return
	}
	sys.playSound(id)
}

func (sys *AudioSystem) playSound(id resource.AudioID) resource.Audio {
	res := sys.loader.LoadAudio(id)
	volume := sys.groupVolume[res.Group]
	if volume != 0 {
		res.Player.SetVolume(res.Volume * sys.groupVolume[res.Group])
		res.Player.Rewind()
		res.Player.Play()
	}
	return res
}
