package geolite2

import (
	"math"
	"net"

	"go.uber.org/zap"

	"github.com/oschwald/geoip2-golang"
	"github.com/vu-ngoc-son/XDP-p2p-router/database"
	"github.com/vu-ngoc-son/XDP-p2p-router/internal/common"
	"github.com/vu-ngoc-son/XDP-p2p-router/internal/logger"
)

type GeoLite2 struct {
	ASN, City, Country          *geoip2.Reader
	HostPublicIP                string
	HostLongitude, HostLatitude float64
	HostASN                     uint
	HostISP                     string
	HostCountryCode             string
}

func NewGeoLite2(asnDBPath, cityDBPath, countryDBPath, hostPublicIP string) (*GeoLite2, error) {
	asnDB, err := geoip2.Open(asnDBPath)
	if err != nil {
		return nil, err
	}

	cityDB, err := geoip2.Open(cityDBPath)
	if err != nil {
		return nil, err
	}

	countryDB, err := geoip2.Open(countryDBPath)
	if err != nil {
		return nil, err
	}

	asnRecord, err := asnDB.ASN(net.ParseIP(hostPublicIP))
	if err != nil {
		return nil, err
	}

	cityRecord, err := cityDB.City(net.ParseIP(hostPublicIP))
	if err != nil {
		return nil, err
	}

	countryRecord, err := countryDB.Country(net.ParseIP(hostPublicIP))
	if err != nil {
		return nil, err
	}

	return &GeoLite2{
		ASN:             asnDB,
		City:            cityDB,
		Country:         countryDB,
		HostPublicIP:    hostPublicIP,
		HostLatitude:    cityRecord.Location.Latitude,
		HostLongitude:   cityRecord.Location.Longitude,
		HostASN:         asnRecord.AutonomousSystemNumber,
		HostISP:         asnRecord.AutonomousSystemOrganization,
		HostCountryCode: countryRecord.Country.IsoCode,
	}, nil
}

func (g *GeoLite2) Close() {
	myLogger := logger.GetLogger()
	err := g.ASN.Close()
	if err != nil {
		myLogger.Error("failed to close asn db", zap.Error(common.ErrFailedToCloseGeoLite2))
	}

	err = g.Country.Close()
	if err != nil {
		myLogger.Error("failed to close country db", zap.Error(common.ErrFailedToCloseGeoLite2))
	}

	err = g.City.Close()
	if err != nil {
		myLogger.Error("failed to close city db", zap.Error(common.ErrFailedToCloseGeoLite2))
	}
	myLogger.Info("close geolite2 dbs successfully")
}

func (g *GeoLite2) IPInfo(IP net.IP, IPNumber uint32, rxPkt, rxByte uint64) (*database.Peers, error) {
	myLogger := logger.GetLogger()

	if common.IsPrivateIP(IP) {
		return &database.Peers{
			IpAddress:    IP.String(),
			IpNumber:     IPNumber,
			Asn:          g.HostASN,
			Isp:          g.HostISP,
			CountryCode:  g.HostCountryCode,
			Longitude:    g.HostLongitude,
			Latitude:     g.HostLatitude,
			Distance:     0.0,
			TotalBytes:   rxByte,
			TotalPackets: rxPkt,
		}, nil
	}

	asnRecord, err := g.ASN.ASN(IP)
	if err != nil {
		myLogger.Error("error while querying asn", zap.Error(common.ErrFailedToQueryGeoLite2))
		return nil, err
	}
	cityRecord, err := g.City.City(IP)
	if err != nil {
		myLogger.Error("error while querying city", zap.Error(common.ErrFailedToQueryGeoLite2))
		return nil, err
	}
	latitude := cityRecord.Location.Latitude
	longitude := cityRecord.Location.Longitude
	distance := g.DistanceToHost(latitude, longitude)

	countryRecord, err := g.Country.Country(IP)
	if err != nil {
		myLogger.Error("error while querying country", zap.Error(common.ErrFailedToQueryGeoLite2))
		return nil, err
	}
	countryCode := countryRecord.Country.IsoCode
	if countryCode == "" {
		countryCode = "OTHER"
	}

	myLogger.Info("get peer info successfully", zap.String("peer_address", IP.String()), zap.Float64("distance", distance))
	return &database.Peers{
		IpAddress:    IP.String(),
		IpNumber:     IPNumber,
		Asn:          asnRecord.AutonomousSystemNumber,
		Isp:          asnRecord.AutonomousSystemOrganization,
		CountryCode:  countryCode,
		Longitude:    longitude,
		Latitude:     latitude,
		Distance:     distance,
		TotalBytes:   rxByte,
		TotalPackets: rxPkt,
	}, nil
}

func (g *GeoLite2) HostInfo() (*database.Hosts, error) {
	myLogger := logger.GetLogger()
	IP := net.ParseIP(g.HostPublicIP)
	asnRecord, err := g.ASN.ASN(IP)
	if err != nil {
		myLogger.Error("error while querying asn", zap.Error(common.ErrFailedToQueryGeoLite2))
		return nil, err
	}
	cityRecord, err := g.City.City(IP)
	if err != nil {
		myLogger.Error("error while querying city", zap.Error(common.ErrFailedToQueryGeoLite2))
		return nil, err
	}
	countryRecord, err := g.Country.Country(IP)
	if err != nil {
		myLogger.Error("error while querying country", zap.Error(common.ErrFailedToQueryGeoLite2))
		return nil, err
	}

	latitude := cityRecord.Location.Latitude
	longitude := cityRecord.Location.Longitude

	distance := 0.0

	myLogger.Info("get host info successfully")
	return &database.Hosts{
		Ip:          g.HostPublicIP,
		Asn:         asnRecord.AutonomousSystemNumber,
		Isp:         asnRecord.AutonomousSystemOrganization,
		CountryCode: countryRecord.Country.IsoCode,
		Longitude:   longitude,
		Latitude:    latitude,
		Distance:    distance,
	}, nil
}

// DistanceToHost references: https://www.geodatasource.com/developers/go
func (g *GeoLite2) DistanceToHost(latitude, longitude float64) float64 {
	myLogger := logger.GetLogger()

	const PI float64 = math.Pi

	radianLat1 := PI * latitude / 180
	radianLat2 := PI * g.HostLatitude / 180

	theta := longitude - g.HostLongitude
	radianTheta := PI * theta / 180

	dist := math.Sin(radianLat1)*math.Sin(radianLat2) + math.Cos(radianLat1)*math.Cos(radianLat2)*math.Cos(radianTheta)

	if dist > 1 {
		dist = 1
	}

	dist = math.Acos(dist)
	dist = dist * 180 / PI
	dist = dist * 60 * 1.1515
	dist = dist * 1.609344 //Kms

	if dist == math.NaN() {
		myLogger.Debug("got result NaN when calculate host distance",
			zap.Float64("longitude", longitude),
			zap.Float64("latitude", latitude),
			zap.Float64("distance (Kms)", dist),
		)
		return 0.0
	}
	myLogger.Debug("distance calculated",
		zap.Float64("longitude", longitude),
		zap.Float64("latitude", latitude),
		zap.Float64("distance (Kms)", dist),
	)
	return dist
}
