package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"sync"

	"github.com/thomersch/gosmparse"
)

func main() {
	filename := flag.String("filename", "extract.osm.pbf", "file to be measured")
	flag.Parse()
	Parse(*filename)
}

type Output struct {
	FileName string
	Lengths  map[string]float64
	Counts   map[string]int
}

const EARTH_RADIUS = 6371

type point struct {
	Lat, Lon float64
}

func (p point) GreatCircleDistance(p2 point) float64 {
	dLat := (p2.Lat - p.Lat) * (math.Pi / 180.0)
	dLon := (p2.Lon - p.Lon) * (math.Pi / 180.0)

	lat1 := p.Lat * (math.Pi / 180.0)
	lat2 := p2.Lat * (math.Pi / 180.0)

	a1 := math.Sin(dLat/2) * math.Sin(dLat/2)
	a2 := math.Sin(dLon/2) * math.Sin(dLon/2) * math.Cos(lat1) * math.Cos(lat2)

	a := a1 + a2

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return EARTH_RADIUS * c
}

type dataHandler struct {
	Nodes    map[int64]point
	NodesMtx sync.Mutex
	Ways     []gosmparse.Way
	WaysMtx  sync.Mutex

	NdCounts    map[string]int
	NdCountsMtx sync.Mutex
}

func (d *dataHandler) ReadNode(n gosmparse.Node) {
	d.NodesMtx.Lock()
	d.Nodes[n.ID] = point{float64(n.Lat), float64(n.Lon)}
	d.NodesMtx.Unlock()

	if v, ok := n.Tags["railway"]; ok {
		if v == "milestone" || v == "signal" || v == "halt" || v == "station" {
			d.NdCountsMtx.Lock()
			d.NdCounts[fmt.Sprintf("railway=%v", v)] += 1
			d.NdCountsMtx.Unlock()
		}
	}
	for _, kn := range []string{"railway:signal:speed_limit", "railway:signal:speed_limit_distant", "railway:signal:combined", "railway:signal:main", "railway:signal:distant"} {
		if _, ok := n.Tags[kn]; ok {
			d.NdCountsMtx.Lock()
			d.NdCounts[kn] += 1
			d.NdCountsMtx.Unlock()
		}
	}
}

func (d *dataHandler) ReadWay(w gosmparse.Way) {
	d.WaysMtx.Lock()
	d.Ways = append(d.Ways, w)
	d.WaysMtx.Unlock()
}

func (d *dataHandler) ReadRelation(r gosmparse.Relation) {
	// relations are not supported
}

func (d *dataHandler) CalculateLength() map[string]float64 {
	var (
		sums   = map[string]float64{}
		cp, lp point
	)

	for _, w := range d.Ways {
		for pos, nd := range w.NodeIDs {
			cp = d.Nodes[nd]
			if pos != 0 {
				if v, ok := w.Tags["railway"]; ok {
					sums[fmt.Sprintf("railway=%v", v)] += lp.GreatCircleDistance(cp)

					if v == "rail" {
						if _, ok := w.Tags["maxspeed"]; ok {
							sums[fmt.Sprintf("railway=%v and maxspeed=*", v)] += lp.GreatCircleDistance(cp)
						}
					}
				}
			}
			lp = d.Nodes[nd]
		}
	}
	return sums
}

func Parse(filename string) {
	r, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	dec := gosmparse.NewDecoder(r)
	// Parse will block until it is done or an error occurs.
	dh := dataHandler{
		Nodes:    make(map[int64]point),
		NdCounts: make(map[string]int),
	}
	err = dec.Parse(&dh)
	if err != nil {
		panic(err)
	}
	log.Printf("nodes: %v, ways: %v", len(dh.Nodes), len(dh.Ways))

	o := Output{
		FileName: filename,
		Lengths:  dh.CalculateLength(),
		Counts:   dh.NdCounts,
	}
	err = json.NewEncoder(os.Stdout).Encode(o)
	if err != nil {
		log.Fatal(err)
	}
}
