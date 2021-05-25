package ip2location

import (
	"github.com/iovisor/gobpf/bcc"
	"gorm.io/gorm"
)

type Locator struct {
	BPFTable *bcc.Table
	DB       *gorm.DB
}

func NewLocator(t *bcc.Table, db *gorm.DB) *Locator {
	return &Locator{
		BPFTable: t,
		DB:       db,
	}
}
