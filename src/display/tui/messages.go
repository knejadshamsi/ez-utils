package tui

// Message types for state updates
type StepMsg int
type ProcessCompleteMsg bool
type SystemStatsMsg struct {
	Cpu float64
	Ram float64
}
type TickMsg struct{}
type BlinkMsg struct{}

// Message types for XML processing  (steps 0-1)
type ChunkCountMsg int
type PersonsFoundMsg int
type BytesReadMsg int
type FirstChunkFixedMsg bool
type LastChunkFixedMsg bool

// Message types for location processing phase (steps 2-3)
type AgentCounterMsg int
type ChunkCounterMsg struct {
	Current int
	Total   int
}
type DbCounterMsg struct {
	Current int
	Total   int
}

// Message types for data transformation (steps 4-18)
type CoordinateCounterMsg int
type BinCounterMsg struct {
	Current int
	Total   int
}
type AgentBinCounterMsg int
type ScaleCounterMsg struct {
	Current int
	Total   int
}
type OutputScaleMsg int
type OutputCounterMsg struct {
	Current int
	Total   int
}
type CleanupCounterMsg struct {
	Files int
	Dirs  int
	Bytes int
}

// CustomMsg for generic messages
type CustomMsg struct {
	Type int
	Data map[string]interface{}
}