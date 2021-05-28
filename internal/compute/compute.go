package compute

import (
	"math"

	"github.com/vu-ngoc-son/XDP-p2p-router/database"
	dbSqlite "github.com/vu-ngoc-son/XDP-p2p-router/database/db-sqlite"
)

const (
	e      = 0.1
	e1     = 1.0
	alpha1 = 1000.0
	alpha2 = 1000.0
	alpha3 = 2000.0
	B      = 2000000.0
	minB   = 200.0
)

type ServiceCalculator struct {
	DB *dbSqlite.SQLiteDB
}

func NewCalculator(db *dbSqlite.SQLiteDB) ServiceCalculator {
	return ServiceCalculator{
		DB: db,
	}
}

//UpdatePeersLimit calculate limit of all peers and update to database.
func (c ServiceCalculator) UpdatePeersLimit() error {
	peers, err := c.DB.GetPeers()
	if err != nil {
		return err
	}

	for _, p := range peers {
		_, err = c.LimitByIP(p, true)
		if err != nil {
			return err
		}
	}

	return nil
}

//LimitByIP calculate limit bandwidth of a specific ip address
func (c ServiceCalculator) LimitByIP(p database.Peers, updateDB bool) (*database.Limits, error) {
	limit := B
	n1, n2, n3, f1, f2, f3 := c.prepareArgs(p)

	logicalDistance := f1*math.Exp(-1/(n1+e)) + f2*math.Exp(-1/(n2+e)) + f3*math.Exp(-1/n3+e)
	limit = B / (logicalDistance + e1)

	if limit == math.NaN() {
		limit = B
	}

	if limit < minB {
		limit = minB
	}

	l := database.Limits{
		Ip:        p.Ip,
		Bandwidth: limit,
	}

	if updateDB {
		err := c.DB.UpdatePeerLimit(&l)
		if err != nil {
			return nil, err
		}
	}

	return &l, nil
}

func (c ServiceCalculator) prepareArgs(p database.Peers) (n1, n2, n3 float64, f1, f2, f3 float64) {
	f1 = 0
	f2 = 0
	f3 = 0
	n1, n2, n3, err := c.DB.FindNearByPeers()
	if err != nil {
		return math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN(), math.NaN()
	}
	sameASN, sameISP, sameCountry := c.DB.CompareToHost(p)

	if !sameASN {
		f1 = alpha1
		f3 = p.Distance
	}
	if !sameISP {
		f2 = alpha2
	}
	if !sameCountry {
		f3 = f3 + alpha3
	}

	return n1, n2, n3, f1, f2, f3
}
