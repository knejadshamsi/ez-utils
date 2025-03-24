package population

import (
	"sync/atomic"
)

type Person struct {
	ID    string `xml:"id,attr"`
	Plans []Plan `xml:"plan"`
}

type Plan struct {
	Selected   string     `xml:"selected,attr"`
	Activities []Activity `xml:"activity"`
}

type Activity struct {
	Type string  `xml:"type,attr"`
	X    float64 `xml:"x,attr"`
	Y    float64 `xml:"y,attr"`
}

type Agent struct {
	ID        string
	Longitude float64
	Latitude  float64
	XML       string
}

// PopulationModel holds the state for population processing
type PopulationModel struct {
	// Counters for display and tracking
	personsFound uint64
	chunkCount   int
	agentCounter uint64
	chunkCounter int
	dbCounter    int
	dbTotal      int

	// Flags
	UseDatabase bool
}

// SetPersonsFound sets the total number of persons found
func (m *PopulationModel) SetPersonsFound(count uint64) {
	atomic.StoreUint64(&m.personsFound, count)
}

// SetChunkCount sets the total number of chunks
func (m *PopulationModel) SetChunkCount(count int) {
	m.chunkCount = count
}

// SetAgentCounter updates the agent counter
func (m *PopulationModel) SetAgentCounter(count uint64) {
	atomic.StoreUint64(&m.agentCounter, count)
}

// SetChunkCounter updates the current chunk counter
func (m *PopulationModel) SetChunkCounter(current int, total int) {
	m.chunkCounter = current
	m.chunkCount = total
}

// SetDbCounter updates the database operation counters
func (m *PopulationModel) SetDbCounter(current int, total int) {
	m.dbCounter = current
	m.dbTotal = total
}

// GetStats returns the current model stats
func (m *PopulationModel) GetStats() (personsFound uint64, chunkCount int, agentCounter uint64, chunkCounter int, dbCounter int, dbTotal int) {
	return atomic.LoadUint64(&m.personsFound), m.chunkCount, atomic.LoadUint64(&m.agentCounter), m.chunkCounter, m.dbCounter, m.dbTotal
}
