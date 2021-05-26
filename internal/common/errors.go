package common

import "errors"

var (
	ErrFailedToConvertIP     = errors.New("failed to parse ip")
	ErrEmptyResultConvertIP  = errors.New("got empty result while parsing ip")
	ErrFailedToConvertUint32 = errors.New("failed to parse uint32")
	ErrFailedToConvertUint64 = errors.New("failed to parse uint64")
	ErrFailedToQueryGeoLite2 = errors.New("failed to query geolite2")
	ErrFailedToCloseGeoLite2 = errors.New("failed to close geolite2")
)
