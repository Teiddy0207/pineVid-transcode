package util

import (
	"encoding/json"
	"go.uber.org/zap"
	"os/exec"
)

type MediaInfo struct {
	Media struct {
		Ref   string `json:"@ref"`
		Track []struct {
			Type                  string `json:"@type"`
			VideoCount            string `json:"VideoCount,omitempty"`
			AudioCount            string `json:"AudioCount,omitempty"`
			FileExtension         string `json:"FileExtension,omitempty"`
			Format                string `json:"Format"`
			FormatProfile         string `json:"Format_Profile,omitempty"`
			CodecID               string `json:"CodecID"`
			CodecIDCompatible     string `json:"CodecID_Compatible,omitempty"`
			FileSize              string `json:"FileSize,omitempty"`
			Duration              string `json:"Duration"`
			OverallBitRateMode    string `json:"OverallBitRate_Mode,omitempty"`
			OverallBitRate        string `json:"OverallBitRate,omitempty"`
			FrameRate             string `json:"FrameRate"`
			FrameCount            string `json:"FrameCount"`
			StreamSize            string `json:"StreamSize"`
			HeaderSize            string `json:"HeaderSize,omitempty"`
			DataSize              string `json:"DataSize,omitempty"`
			FooterSize            string `json:"FooterSize,omitempty"`
			IsStreamable          string `json:"IsStreamable,omitempty"`
			EncodedDate           string `json:"Encoded_Date"`
			TaggedDate            string `json:"Tagged_Date"`
			FileModifiedDate      string `json:"File_Modified_Date,omitempty"`
			FileModifiedDateLocal string `json:"File_Modified_Date_Local,omitempty"`
			Extra                 struct {
				CodecConfigurationBox string `json:"CodecConfigurationBox"`
				TIM                   string `json:"TIM"`
				TSC                   string `json:"TSC"`
				TSZ                   string `json:"TSZ"`
			} `json:"extra,omitempty"`
			StreamOrder                    string `json:"StreamOrder,omitempty"`
			ID                             string `json:"ID,omitempty"`
			FormatLevel                    string `json:"Format_Level,omitempty"`
			FormatSettingsCABAC            string `json:"Format_Settings_CABAC,omitempty"`
			FormatSettingsRefFrames        string `json:"Format_Settings_RefFrames,omitempty"`
			BitRate                        string `json:"BitRate,omitempty"`
			Width                          string `json:"Width,omitempty"`
			Height                         string `json:"Height,omitempty"`
			StoredHeight                   string `json:"Stored_Height,omitempty"`
			SampledWidth                   string `json:"Sampled_Width,omitempty"`
			SampledHeight                  string `json:"Sampled_Height,omitempty"`
			PixelAspectRatio               string `json:"PixelAspectRatio,omitempty"`
			DisplayAspectRatio             string `json:"DisplayAspectRatio,omitempty"`
			Rotation                       string `json:"Rotation,omitempty"`
			FrameRateMode                  string `json:"FrameRate_Mode,omitempty"`
			Standard                       string `json:"Standard,omitempty"`
			ColorSpace                     string `json:"ColorSpace,omitempty"`
			ChromaSubsampling              string `json:"ChromaSubsampling,omitempty"`
			BitDepth                       string `json:"BitDepth,omitempty"`
			ScanType                       string `json:"ScanType,omitempty"`
			Language                       string `json:"Language,omitempty"`
			ColourDescriptionPresent       string `json:"colour_description_present,omitempty"`
			ColourDescriptionPresentSource string `json:"colour_description_present_Source,omitempty"`
			ColourRange                    string `json:"colour_range,omitempty"`
			ColourRangeSource              string `json:"colour_range_Source,omitempty"`
			ColourPrimaries                string `json:"colour_primaries,omitempty"`
			ColourPrimariesSource          string `json:"colour_primaries_Source,omitempty"`
			TransferCharacteristics        string `json:"transfer_characteristics,omitempty"`
			TransferCharacteristicsSource  string `json:"transfer_characteristics_Source,omitempty"`
			MatrixCoefficients             string `json:"matrix_coefficients,omitempty"`
			MatrixCoefficientsSource       string `json:"matrix_coefficients_Source,omitempty"`
			FormatAdditionalFeatures       string `json:"Format_AdditionalFeatures,omitempty"`
			SourceDuration                 string `json:"Source_Duration,omitempty"`
			BitRateMode                    string `json:"BitRate_Mode,omitempty"`
			BitRateMaximum                 string `json:"BitRate_Maximum,omitempty"`
			Channels                       string `json:"Channels,omitempty"`
			ChannelPositions               string `json:"ChannelPositions,omitempty"`
			ChannelLayout                  string `json:"ChannelLayout,omitempty"`
			SamplesPerFrame                string `json:"SamplesPerFrame,omitempty"`
			SamplingRate                   string `json:"SamplingRate,omitempty"`
			SamplingCount                  string `json:"SamplingCount,omitempty"`
			SourceFrameCount               string `json:"Source_FrameCount,omitempty"`
			CompressionMode                string `json:"Compression_Mode,omitempty"`
			StreamSizeProportion           string `json:"StreamSize_Proportion,omitempty"`
			SourceStreamSize               string `json:"Source_StreamSize,omitempty"`
			SourceStreamSizeProportion     string `json:"Source_StreamSize_Proportion,omitempty"`
		} `json:"track"`
	} `json:"media"`
}

func GetMediaInfo(p string) MediaInfo {
	var mediaInfo MediaInfo
	cmd := exec.Command("mediainfo", "--output=JSON", p)
	if out, err := cmd.Output(); err == nil {
		if err := json.Unmarshal(out, &mediaInfo); err != nil {
			Logger.Error("GetMediaInfo json.Unmarshal", zap.String("error", err.Error()))
		}
	} else {
		Logger.Error("GetMediaInfo cmd.Output", zap.String("error", err.Error()))
	}
	return mediaInfo
}
