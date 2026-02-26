package handlers

import (
	"context"
	"encoding/base64"

	humatypes "github.com/Cogwheel-Validator/spectra-gnoland-indexer/api/huma-types"
	"github.com/danielgtaylor/huma/v2"
)

func ConvertFromBase64toBase64Url(ctx context.Context, input *humatypes.ConvertFromBase64toBase64UrlInput) (*humatypes.ConvertFromBase64toBase64UrlOutput, error) {
	if len(input.TxHash64) != 44 {
		return nil, huma.Error400BadRequest("Input is not a valid tx hash")
	}
	decoded, err := base64.StdEncoding.DecodeString(input.TxHash64)
	if err != nil {
		return nil, huma.Error400BadRequest("Input is not valid base64 encoded", err)
	}
	response := humatypes.ConvertFromBase64toBase64UrlOutput{
		Body: humatypes.ConvertFromBase64toBase64UrlBody{
			TxHash64Url: base64.URLEncoding.EncodeToString(decoded),
		},
	}
	return &response, nil
}

func ConvertFromBase64UrlToBase64(ctx context.Context, input *humatypes.ConvertFromBase64UrlToBase64Input) (*humatypes.ConvertFromBase64UrlToBase64Output, error) {
	if len(input.TxHash64Url) != 44 {
		return nil, huma.Error400BadRequest("Input is not a valid tx hash")
	}
	decoded, err := base64.URLEncoding.DecodeString(input.TxHash64Url)
	if err != nil {
		return nil, huma.Error400BadRequest("Input is not valid base64url encoded", err)
	}
	response := humatypes.ConvertFromBase64UrlToBase64Output{
		Body: humatypes.ConvertFromBase64UrlToBase64Body{
			TxHash64: base64.StdEncoding.EncodeToString(decoded),
		},
	}
	return &response, nil
}
