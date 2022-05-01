package resource

import (
	"fmt"
	"image"
	"io"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

// Loader is used to load and cache game resources like images and audio files.
type Loader struct {
	// OpenAssetFunc is used to open an asset resource identified by its path.
	// The returned resource will be closed after it will be loaded.
	OpenAssetFunc func(path string) io.ReadCloser

	ImageRegistry ImageRegistry
	AudioRegistry AudioRegistry

	wavDecoder wavDecoder
	oggDecoder oggDecoder

	images map[ID]*ebiten.Image
	wavs   map[ID]*wav.Stream
	oggs   map[ID]*vorbis.Stream
}

type wavDecoder interface {
	DecodeWAV(r io.Reader) (*wav.Stream, error)
}

type oggDecoder interface {
	DecodeOGG(r io.Reader) (*vorbis.Stream, error)
}

func NewLoader(wd wavDecoder, od oggDecoder) *Loader {
	l := &Loader{
		images:     make(map[ID]*ebiten.Image),
		wavs:       make(map[ID]*wav.Stream),
		oggs:       make(map[ID]*vorbis.Stream),
		wavDecoder: wd,
		oggDecoder: od,
	}
	l.AudioRegistry.mapping = make(map[ID]Audio)
	l.ImageRegistry.mapping = make(map[ID]Image)
	return l
}

func (l *Loader) PreloadImage(id ID) {
	l.LoadImage(id)
}

func (l *Loader) PreloadAudio(id ID) {
	audioInfo := l.GetAudioInfo(id)
	if strings.HasSuffix(audioInfo.Path, ".ogg") {
		l.LoadOGG(id)
	} else {
		l.LoadWAV(id)
	}
}

func (l *Loader) PreloadWAV(id ID) {
	l.LoadWAV(id)
}

func (l *Loader) PreloadOGG(id ID) {
	l.LoadOGG(id)
}

func (l *Loader) LoadWAV(id ID) *wav.Stream {
	stream, ok := l.wavs[id]
	if !ok {
		wavInfo := l.GetAudioInfo(id)
		r := l.OpenAssetFunc(wavInfo.Path)
		defer func() {
			if err := r.Close(); err != nil {
				panic(fmt.Sprintf("closing %q wav reader: %v", wavInfo.Path, err))
			}
		}()
		var err error
		stream, err = l.wavDecoder.DecodeWAV(r)
		if err != nil {
			panic(fmt.Sprintf("decode %q wav: %v", wavInfo.Path, err))
		}
		l.wavs[id] = stream
	}
	return stream
}

func (l *Loader) GetAudioInfo(id ID) Audio {
	info, ok := l.AudioRegistry.mapping[id]
	if !ok {
		panic(fmt.Sprintf("unregistered audio with id=%d", id))
	}
	return info
}

func (l *Loader) LoadOGG(id ID) *vorbis.Stream {
	stream, ok := l.oggs[id]
	if !ok {
		oggInfo := l.GetAudioInfo(id)
		r := l.OpenAssetFunc(oggInfo.Path)
		defer func() {
			if err := r.Close(); err != nil {
				panic(fmt.Sprintf("closing %q ogg reader: %v", oggInfo.Path, err))
			}
		}()
		var err error
		stream, err = l.oggDecoder.DecodeOGG(r)
		if err != nil {
			panic(fmt.Sprintf("decode %q ogg: %v", oggInfo.Path, err))
		}
		l.oggs[id] = stream
	}
	return stream
}

func (l *Loader) LoadImage(id ID) *ebiten.Image {
	img, ok := l.images[id]
	if !ok {
		imageInfo, ok := l.ImageRegistry.mapping[id]
		if !ok {
			panic(fmt.Sprintf("unregistered image with id=%d", id))
		}
		r := l.OpenAssetFunc(imageInfo.Path)
		defer func() {
			if err := r.Close(); err != nil {
				panic(fmt.Sprintf("closing %q image reader: %v", imageInfo.Path, err))
			}
		}()
		rawImage, _, err := image.Decode(r)
		if err != nil {
			panic(fmt.Sprintf("decode %q image: %v", imageInfo.Path, err))
		}
		img = ebiten.NewImageFromImage(rawImage)
		l.images[id] = img
	}
	return img
}