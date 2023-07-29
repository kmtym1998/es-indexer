package model

import (
	"fmt"

	"github.com/kmtym1998/es-indexer/elasticsearch"
	"github.com/samber/lo"
)

type Address struct {
	ID               int    `json:"id"`
	ZipCode          string `json:"zipCode"`
	Prefecture       string `json:"prefecture"`
	Municipality     string `json:"municipality"`
	Town             string `json:"town"`
	Concat           string `json:"concat"`
	PrefectureKana   string `json:"prefectureKana"`
	MunicipalityKana string `json:"municipalityKana"`
	TownKana         string `json:"townKana"`
	ConcatKana       string `json:"concatKana"`
}

type AddressList []Address

func (a Address) NodeIdentifier() string {
	return fmt.Sprintf("%d", a.ID)
}

func (a AddressList) IndexName() string {
	return "addresses"
}

func (a AddressList) ToList() []elasticsearch.DocumentNode {
	return lo.Map(a, func(item Address, _ int) elasticsearch.DocumentNode {
		return item
	})
}
