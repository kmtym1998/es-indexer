package model

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