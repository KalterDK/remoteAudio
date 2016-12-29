package audio

import (
	"errors"
	"fmt"

	sbAudio "github.com/dh1tw/remoteAudio/sb_audio"
	"github.com/gogo/protobuf/proto"
	"github.com/hraban/opus"
)

type deserializer struct {
	*AudioDevice
	opusDecoder *opus.Decoder
	opusBuffer  []float32
}

func (d *deserializer) DeserializeOpusAudioMsg(data []byte) error {

	msg := sbAudioDataPool.Get().(*sbAudio.AudioData)
	defer sbAudioDataPool.Put(msg)

	err := proto.Unmarshal(data, msg)
	if err != nil {
		return err
	}

	len, err := d.opusDecoder.DecodeFloat32(msg.GetAudioRaw(), d.opusBuffer)
	if err != nil {
		return err
	}

	fmt.Println("out bytes decoded:", len)

	d.out = d.opusBuffer[:len*d.Channels]

	return nil
}

// DeserializeAudioMsg deserializes protocol buffers containing audio frames with
// the corresponding meta data. In case the Audio channels and / or Samplerate
// doesn't match with the hosts values, both will be converted to match.
func (ad *AudioDevice) DeserializeAudioMsg(data []byte) error {

	msg := sbAudioDataPool.Get().(*sbAudio.AudioData)
	defer sbAudioDataPool.Put(msg)

	err := proto.Unmarshal(data, msg)
	if err != nil {
		return err
	}

	var samplingrate float64
	var channels, bitrate int

	channels = int(msg.GetChannels())
	if channels == 0 {
		return errors.New("invalid amount of channels")
	}

	samplingrate = float64(msg.GetSamplingRate())
	if samplingrate == 0 {
		return errors.New("invalid samplerate")
	}

	// only accept 8, 12, 16 or 32 bit streams
	bitrate = int(msg.GetBitrate())

	if bitrate != 8 && bitrate != 12 && bitrate != 16 && bitrate != 32 {
		return errors.New("incompatible bitrate")
	}

	if len(msg.AudioPacked) == 0 {
		return errors.New("empty audio buffer")
	}

	// convert the data to float32 (8bit, 12bit, 16bit, 32bit)
	convertedAudio := make([]float32, 0, len(msg.AudioPacked))
	for _, sample := range msg.AudioPacked {
		convertedAudio = append(convertedAudio, float32(sample)/bitMapToFloat32[bitrate])
	}

	if msg.Codec == sbAudio.Codec_PCM {
		// if necessary, adjust the channels
		if channels != ad.Channels {

			// audio device is STEREO but we received MONO
			if channels == MONO && ad.Channels == STEREO {
				expanded := make([]float32, 0, len(convertedAudio)*2)
				// left channel = right channel
				for _, sample := range convertedAudio {
					expanded = append(expanded, sample)
					expanded = append(expanded, sample)
				}
				convertedAudio = expanded

			} else if channels == STEREO && ad.Channels == MONO {
				// audio device is MONO but we received STEREO
				reduced := make([]float32, 0, len(convertedAudio)/2)
				// chop of the right channel
				for i := 0; i < len(convertedAudio); i += 2 {
					reduced = append(reduced, convertedAudio[i])
				}
				convertedAudio = reduced
			}
		}

		var resampledAudio []float32

		// if necessary, resample the audio
		if samplingrate != ad.Samplingrate {
			ratio := ad.Samplingrate / samplingrate // output samplerate / input samplerate
			resampledAudio, err = ad.Converter.Process(convertedAudio, ratio, false)
			if err != nil {
				return err
			}
			ad.out = resampledAudio
		} else {
			ad.out = convertedAudio
		}
	}
	return nil
}
