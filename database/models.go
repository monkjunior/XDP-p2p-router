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
	Ip          string `gorm:"primaryKey"`
	Asn         uint
	Isp         string
	CountryCode string
	Longitude   float64
	Latitude    float64
	Distance    float64
}

type Limits struct {
	Ip        string `gorm:"primaryKey"`
	Bandwidth float64
}
