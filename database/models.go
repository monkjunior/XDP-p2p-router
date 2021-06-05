package database

type Hosts struct {
	Ip          string `gorm:"primaryKey"`
	Asn         uint
	Isp         string
	CountryCode string
	Longitude   float64
	Latitude    float64
	Distance    float64
}

type Peers struct {
	IpNumber     uint32 `gorm:"primaryKey"`
	IpAddress    string
	Asn          uint
	Isp          string
	CountryCode  string
	Longitude    float64
	Latitude     float64
	Distance     float64
	Bandwidth    float64
	TotalPackets uint64
	TotalBytes   uint64
}

type Limits struct {
	Ip        string `gorm:"primaryKey"`
	Bandwidth float64
}
