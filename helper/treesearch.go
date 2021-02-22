package main

import (
	"embed"
	"encoding/json"
	"fmt"
)

//Region region struct
type Region struct {
	RegionID       string `json:"RegionID,omitempty"`
	parentRegionID string
	RegionName     string            `json:"RegionName,omitempty"`
	LevelType      string            `json:"LevelType,omitempty"`
	ChildRegions   map[string]Region `json:"ChildRegions,omitempty"`
}

var (
	//go:embed asgs.json
	asgs embed.FS
	//go:embed lga.json
	lga embed.FS
	//go:embed poa.json
	poa embed.FS
	//go:embed ssc.json
	ssc embed.FS
	//go:embed gccsa.json
	gccsa embed.FS
)

var asgsRegion Region
var lgaRegion Region
var poaRegion Region
var sscRegion Region
var gccsaRegion Region

func init() {
	data, _ := asgs.ReadFile("asgs.json")
	err := json.Unmarshal(data, &asgsRegion)
	if err != nil {
		fmt.Printf("Error %s", err)
	}

	data, _ = lga.ReadFile("lga.json")
	err = json.Unmarshal(data, &lgaRegion)
	if err != nil {
		fmt.Printf("Error %s", err)
	}
	data, _ = poa.ReadFile("poa.json")
	err = json.Unmarshal(data, &poaRegion)
	if err != nil {
		fmt.Printf("Error %s", err)
	}
	data, _ = ssc.ReadFile("ssc.json")
	err = json.Unmarshal(data, &sscRegion)
	if err != nil {
		fmt.Printf("Error %s", err)
	}
	data, _ = gccsa.ReadFile("gccsa.json")
	err = json.Unmarshal(data, &gccsaRegion)
	if err != nil {
		fmt.Printf("Error %s", err)
	}

}

func main() {
	fmt.Println("Start Search")
	fmt.Println(len(asgsRegion.ChildRegions))
	fmt.Println(len(lgaRegion.ChildRegions))
	fmt.Println(len(poaRegion.ChildRegions))
	fmt.Println(len(sscRegion.ChildRegions))
	fmt.Println(len(gccsaRegion.ChildRegions))
}
