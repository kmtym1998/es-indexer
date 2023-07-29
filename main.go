package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	es "kmtym1998/es-zip-codes/elasticsearch"
	"kmtym1998/es-zip-codes/model"

	"github.com/ktnyt/go-moji"
)

func main() {
	f, err := os.Open("./data/ken_all.csv")
	if err != nil {
		log.Fatalln("os.Open", err)
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatalln("ioutil.ReadAll", err)
	}
	content := string(b)
	r := csv.NewReader(strings.NewReader(content))

	results := []*model.Address{}
	for i := 0; ; i++ {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		results = append(results, &model.Address{
			ID:               i,
			ZipCode:          record[2],
			PrefectureKana:   moji.Convert(record[3], moji.HK, moji.ZK),
			MunicipalityKana: moji.Convert(record[4], moji.HK, moji.ZK),
			TownKana:         moji.Convert(record[5], moji.HK, moji.ZK),
			Prefecture:       record[6],
			Municipality:     record[7],
			Town:             record[8],
			Concat:           fmt.Sprintf("%s%s%s", record[6], record[7], record[8]),
			ConcatKana:       fmt.Sprintf("%s%s%s", moji.Convert(record[3], moji.HK, moji.ZK), moji.Convert(record[4], moji.HK, moji.ZK), moji.Convert(record[5], moji.HK, moji.ZK)),
		})
	}

	e, err := es.NewClient()
	if err != nil {
		log.Fatalln("es.NewClient(): ", err)
	}
	if err := e.BulkInsertProducts("addresses", results); err != nil {
		log.Fatalln("e.BulkInsertProducts: ", err)
	}
}
