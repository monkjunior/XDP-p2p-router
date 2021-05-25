package database

type Hosts struct {
	Ip          string
	Asn         uint
	Isp         string
	CountryCode string
	Longitude   float64
	Latitude    float64
	Distance    float64
}

type Peers struct {
	Ip          string
	Asn         uint
	Isp         string
	CountryCode string
	Longitude   float64
	Latitude    float64
	Distance    float64
}

type Limits struct {
	Ip        string
	Bandwidth float64
}
