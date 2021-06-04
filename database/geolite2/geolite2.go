package geolite2

import (
	"fmt"
	"github.com/oschwald/geoip2-golang"
	"github.com/vu-ngoc-son/XDP-p2p-router/database"
	"github.com/vu-ngoc-son/XDP-p2p-router/internal/common"
	"math"
	"net"
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
	err := g.ASN.Close()
	if err != nil {
		fmt.Println(common.ErrFailedToCloseGeoLite2)
	}

	err = g.ASN.Close()
	if err != nil {
		fmt.Println(common.ErrFailedToCloseGeoLite2)
	}

	err = g.ASN.Close()
	if err != nil {
		fmt.Println(common.ErrFailedToCloseGeoLite2)
	}
}

func (g *GeoLite2) IPInfo(ipAddress string) (*database.Peers, error) {
	IP := net.ParseIP(ipAddress)
	asnRecord, err := g.ASN.ASN(IP)
	if err != nil {
		fmt.Println("error while querying asn", common.ErrFailedToQueryGeoLite2)
		return nil, err
	}
	cityRecord, err := g.City.City(IP)
	if err != nil {
		fmt.Println("error while querying city", common.ErrFailedToQueryGeoLite2)
		return nil, err
	}
	countryRecord, err := g.Country.Country(IP)
	if err != nil {
		fmt.Println("error while querying country", common.ErrFailedToQueryGeoLite2)
		return nil, err
	}

	latitude := cityRecord.Location.Latitude
	longitude := cityRecord.Location.Longitude

	distance := g.DistanceToHost(latitude, longitude)

	return &database.Peers{
		Ip:          ipAddress,
		Asn:         asnRecord.AutonomousSystemNumber,
		Isp:         asnRecord.AutonomousSystemOrganization,
		CountryCode: countryRecord.Country.IsoCode,
		Longitude:   longitude,
		Latitude:    latitude,
		Distance:    distance,
	}, nil
}

func (g *GeoLite2) HostInfo() (*database.Hosts, error) {
	IP := net.ParseIP(g.HostPublicIP)
	asnRecord, err := g.ASN.ASN(IP)
	if err != nil {
		fmt.Println("error while querying asn", common.ErrFailedToQueryGeoLite2)
		return nil, err
	}
	cityRecord, err := g.City.City(IP)
	if err != nil {
		fmt.Println("error while querying city", common.ErrFailedToQueryGeoLite2)
		return nil, err
	}
	countryRecord, err := g.Country.Country(IP)
	if err != nil {
		fmt.Println("error while querying country", common.ErrFailedToQueryGeoLite2)
		return nil, err
	}

	latitude := cityRecord.Location.Latitude
	longitude := cityRecord.Location.Longitude

	distance := 0.0

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
	deltaLong := longitude - g.HostLongitude
	deltaLat := latitude - g.HostLatitude

	a := math.Pow(math.Sin(deltaLat/2), 2) + math.Pow(math.Cos(deltaLong/2), 2)*math.Cos(latitude)*math.Cos(longitude)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	R := 6373.0

	distance = R * c
	if distance == math.NaN() {
		return 0
	}

	return distance
}
