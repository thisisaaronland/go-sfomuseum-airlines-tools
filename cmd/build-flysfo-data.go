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

	seen := new(sync.Map)

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

		is_current := utils.Int64Property(f.Bytes(), []string{"properties.mz:is_current"}, -1)

		if is_current != 1 {
			return nil
		}

		concordances, err := whosonfirst.Concordances(f)

		if err != nil {
			return err
		}

		iata_code, ok := concordances["iata:code"]

		if !ok {
			return nil
		}

		wof_id := whosonfirst.Id(f)
		name := whosonfirst.Name(f)

		w, ok := seen.Load(iata_code)

		if ok {
			log.Println("WARNING", iata_code, w, wof_id)
		}

		sfom_id := utils.Int64Property(f.Bytes(), []string{"properties.sfomuseum:airline_id"}, -1)

		a := sfomuseum.Airline{
			WOFID:       wof_id,
			SFOMuseumID: int(sfom_id),
			Name:        name,
			IATACode:    iata_code,
		}

		mu.Lock()
		defer mu.Unlock()

		lookup = append(lookup, a)
		seen.Store(iata_code, wof_id)

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
		Package: "flysfo",
		Data:    string(enc),
	}

	err = template.RenderAirlineData(os.Stdout, &vars)

	if err != nil {
		log.Fatal(err)
	}
}
