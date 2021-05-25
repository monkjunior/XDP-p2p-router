package common

import "errors"

var (
	ErrFailedToConvertIP     = errors.New("failed to parse ip")
	ErrFailedToQueryGeoLite2 = errors.New("failed to query geolite2")
	ErrFailedToCloseGeoLite2 = errors.New("failed to close geolite2")
)
