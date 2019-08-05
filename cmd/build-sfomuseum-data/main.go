package main

import (
	"context"
	"encoding/json"
	"flag"
	"github.com/sfomuseum/go-sfomuseum-airlines-tools/template"
	"github.com/sfomuseum/go-sfomuseum-airlines/sfomuseum"
	"github.com/sfomuseum/go-sfomuseum-geojson/feature"
	sfomuseum_props "github.com/sfomuseum/go-sfomuseum-geojson/properties/sfomuseum"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/properties/whosonfirst"
	"github.com/whosonfirst/go-whosonfirst-geojson-v2/utils"
	"github.com/whosonfirst/go-whosonfirst-index"
	index_utils "github.com/whosonfirst/go-whosonfirst-index/utils"
	"io"
	"log"
	"os"
	"sync"
)

func main() {

	data := flag.String("data", "/usr/local/data/sfomuseum-data-enterprise", "...")

	flag.Parse()

	lookup := make([]sfomuseum.Airline, 0)
	mu := new(sync.RWMutex)

	cb := func(fh io.Reader, ctx context.Context, args ...interface{}) error {

		is_principal, err := index_utils.IsPrincipalWOFRecord(fh, ctx)

		if err != nil {
			return err
		}

		if !is_principal {
			return nil
		}

		f, err := feature.LoadFeatureFromReader(fh)

		if err != nil {
			return err
		}

		pt := sfomuseum_props.Placetype(f)

		if pt != "airline" {
			return nil
		}

		wof_id := whosonfirst.Id(f)
		name := whosonfirst.Name(f)

		sfom_id := utils.Int64Property(f.Bytes(), []string{"properties.sfomuseum:airline_id"}, -1)

		concordances, err := whosonfirst.Concordances(f)

		if err != nil {
			return err
		}

		a := sfomuseum.Airline{
			WOFID:       wof_id,
			SFOMuseumID: int(sfom_id),
			Name:        name,
		}

		iata_code, ok := concordances["iata:code"]

		if ok {
			a.IATACode = iata_code
		}

		icao_code, ok := concordances["icao:code"]

		if ok {
			a.ICAOCode = icao_code
		}

		callsign, ok := concordances["icao:callsign"]

		if ok {
			a.ICAOCallsign = callsign
		}

		id, ok := concordances["wd:id"]

		if ok {
			a.WikidataID = id
		}

		mu.Lock()
		lookup = append(lookup, a)
		mu.Unlock()

		return nil
	}

	idx, err := index.NewIndexer("repo", cb)

	if err != nil {
		log.Fatal(err)
	}

	err = idx.IndexPath(*data)

	if err != nil {
		log.Fatal(err)
	}

	enc, err := json.Marshal(lookup)

	if err != nil {
		log.Fatal(err)
	}

	vars := template.AirlineDataVars{
		Package: "sfomuseum",
		Data:    string(enc),
	}

	err = template.RenderAirlineData(os.Stdout, &vars)

	if err != nil {
		log.Fatal(err)
	}
}
