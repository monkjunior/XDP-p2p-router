package geolite2

import (
	"math"
	"net"

	"go.uber.org/zap"

	bpf "github.com/iovisor/gobpf/bcc"
	"github.com/oschwald/geoip2-golang"
	"github.com/vu-ngoc-son/XDP-p2p-router/database"
	"github.com/vu-ngoc-son/XDP-p2p-router/internal/common"
	"github.com/vu-ngoc-son/XDP-p2p-router/internal/logger"
)

type GeoLite2 struct {
	ASN, City, Country          *geoip2.Reader
	HostPublicIP                string
	HostLongitude, HostLatitude float64
}

func NewGeoLite2(asnDBPath, cityDBPath, countryDBPath, hostPublicIP string) *GeoLite2 {
	asnDB, err := geoip2.Open(asnDBPath)
	if err != nil {
		return nil
	}

	cityDB, err := geoip2.Open(cityDBPath)
	if err != nil {
		return nil
	}

	countryDB, err := geoip2.Open(countryDBPath)
	if err != nil {
		return nil
	}

	cityRecord, err := cityDB.City(net.ParseIP(hostPublicIP))
	if err != nil {
		return nil
	}

	return &GeoLite2{
		ASN:           asnDB,
		City:          cityDB,
		Country:       countryDB,
		HostPublicIP:  hostPublicIP,
		HostLatitude:  cityRecord.Location.Latitude,
		HostLongitude: cityRecord.Location.Longitude,
	}
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

func (g *GeoLite2) IPInfo(ipNumber uint32) (*database.Peers, error) {
	myLogger := logger.GetLogger()
	IPRaw := make([]byte, 4)
	bpf.GetHostByteOrder().PutUint32(IPRaw, ipNumber)

	ipAddress, err := common.ConvertUint8ToIP(IPRaw)
	if err != nil {
		return nil, err
	}
	IP := net.ParseIP(ipAddress)
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

	distance := g.DistanceToHost(latitude, longitude)

	myLogger.Info("get peer info successfully", zap.String("peer_address", ipAddress), zap.Float64("distance", distance))
	return &database.Peers{
		IpAddress:   ipAddress,
		IpNumber:    ipNumber,
		Asn:         asnRecord.AutonomousSystemNumber,
		Isp:         asnRecord.AutonomousSystemOrganization,
		CountryCode: countryRecord.Country.IsoCode,
		Longitude:   longitude,
		Latitude:    latitude,
		Distance:    distance,
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

func (g *GeoLite2) DistanceToHost(latitude, longitude float64) (distance float64) {
	myLogger := logger.GetLogger()
	deltaLong := longitude - g.HostLongitude
	deltaLat := latitude - g.HostLatitude

	a := math.Pow(math.Sin(deltaLat/2), 2) + math.Pow(math.Cos(deltaLong/2), 2)*math.Cos(latitude)*math.Cos(longitude)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	R := 6373.0

	distance = R * c
	if distance == math.NaN() {
		myLogger.Debug("got result NaN when calculate host distance",
			zap.Float64("longitude", longitude),
			zap.Float64("latitude", latitude),
			zap.Float64("delta_long", deltaLong),
			zap.Float64("delta_lat", deltaLat),
			zap.Float64("a", a),
			zap.Float64("c", c),
			zap.Float64("distance", distance),
		)
		return 0
	}

	return distance
}
