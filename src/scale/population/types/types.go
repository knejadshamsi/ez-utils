package types

import "encoding/xml"

// Population represents the root XML element containing all persons
type Population struct {
	XMLName xml.Name `xml:"population"`
	Desc    string   `xml:"desc,attr"`
	Persons []Person `xml:"person"`
}

// Person represents an individual agent in the simulation
type Person struct {
	XMLName xml.Name `xml:"person"`
	ID      string   `xml:"id,attr"`
	Plans   []Plan   `xml:"plan"`
}

// Plan represents a person's activity plan
type Plan struct {
	Selected   string     `xml:"selected,attr"`
	Activities []Activity `xml:"activity"`
}

// Activity represents a single activity in a plan
type Activity struct {
	X float64 `xml:"x,attr"`
	Y float64 `xml:"y,attr"`
}

// Agent represents the database storage format for agents
type Agent struct {
	ID        string
	Longitude float64
	Latitude  float64
	XML       string
}