package audio

import (
	"errors"

	"github.com/dh1tw/remoteAudio/icd"
	"github.com/golang/protobuf/proto"
)

func (ad *AudioDevice) deserializeAudioMsg(data []byte) error {

	msg := icd.AudioData{}
	err := proto.Unmarshal(data, &msg)
	if err != nil {
		return err
	}

	// var channels, samplingrate, bitrate, frames int
	var samplingrate, bitrate int

	// if msg.Channels != nil {
	// 	channels = int(msg.GetChannels())
	// }

	// if msg.FrameLength != nil {
	// 	frames = int(msg.GetFrameLength())
	// }

	if msg.String != nil {
		samplingrate = int(msg.GetSamplingRate())
	}

	// only accept 8 or 16 bit streams
	if msg.Bitrate != nil {
		bitrate = int(msg.GetBitrate())
		if bitrate != 8 && bitrate != 16 && bitrate != 32 {
			return errors.New("incompatible bitrate")
		}
	} else {
		return errors.New("unknown bitrate")
	}

	if bitrate == 8 {
		// if len(msg.Audio) != int(frames*channels) {
		// 	fmt.Println("msg length: ", len(msg.Audio), int(frames*channels), frames*channels)
		// 	return errors.New("audio data does not match frame buffer * channels")
		// }
		// } else if bitrate == 16 {
		// 	if len(msg.Audio) != int(frames*channels)*2 {
		// 		fmt.Println("msg length: ", len(msg.Audio), int(frames*channels), frames*channels)
		// 		return errors.New("audio data does not match frame buffer * channels")
		// 	}
		// } else if bitrate == 32 {
		// 	if len(msg.Audio) != int(frames*channels)*4 {
		// 		fmt.Println("msg length: ", len(msg.Audio), int(frames*channels), frames*channels)
		// 		return errors.New("audio data does not match frame buffer * channels")
		// 	}
	}

	if float64(samplingrate) != ad.Samplingrate {
		return errors.New("unequal sampling rate")
	}

	if msg.Audio != nil || msg.Audio2 != nil {
		if bitrate == 16 {
			// for i := 0; i < len(msg.Audio)/2; i++ {
			// 	sample := binary.LittleEndian.Uint16(msg.Audio[i*2 : i*2+2])
			// 	ad.out.Data16[i] = int16(sample)
			// }
			for i, sample := range msg.Audio2 {
				ad.out.Data16[i] = int16(sample)
			}
		} else if bitrate == 8 {
			// for i, sample := range msg.Audio2 {
			// 	ad.out.Data8[i] = int8(sample)
			// }

			for i := 0; i < len(msg.Audio); i++ {
				ad.out.Data8[i] = int8(msg.Audio[i])
			}
		} else if bitrate == 32 {
			// for i := 0; i < len(msg.Audio)/4; i++ {
			// 	sample := binary.LittleEndian.Uint32(msg.Audio[i*4 : i*4+4])
			// 	ad.out.Data32[i] = int32(sample)
			// }
			for i, sample := range msg.Audio2 {
				ad.out.Data32[i] = sample
			}

		}
	}

	return nil
}