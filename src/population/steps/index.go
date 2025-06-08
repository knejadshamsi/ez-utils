package steps

import (
"bufio"
"context"
"database/sql"
"encoding/xml"
"fmt"
"io"
"os"
"path/filepath"
"runtime"
"strings"
"sync"
"sync/atomic"

"ez-utils/src/display"
"ez-utils/src/population/types"
)

// Use type aliases for better readability
type Person = types.Person
type Plan = types.Plan
type Agent = types.Agent

// ExtractAgentLocations processes XML chunks to extract agent locations and create indexes
func ExtractAgentLocations() error {
// Input directory containing chunks from Step 1
inputDir := filepath.Join("temp", "population", "01_split_population_chunks", "chunks")

// Output directory for index files
outputDir := filepath.Join("temp", "population", "02_extract_agent_locations", "indexes")
if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
return fmt.Errorf("failed to create output directory: %w", err)
}

// Get list of XML chunks from Step 1
chunkFiles, err := getChunkFiles(inputDir)
if err != nil {
return fmt.Errorf("failed to get chunk files: %w", err)
}

// Update display with total chunks
display.SetChunkCounter(0, len(chunkFiles))

// Set up context for coordinating workers
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// Create worker pool based on available CPUs
numWorkers := runtime.NumCPU()
var jobWg sync.WaitGroup
jobChan := make(chan string, numWorkers*2) // Buffer for smoother processing

// Track total agents extracted across all chunks
var totalAgentsExtracted int32

// Launch workers to process chunks
for i := 0; i < numWorkers; i++ {
jobWg.Add(1)
go func() {
defer jobWg.Done()
processChunks(ctx, jobChan, outputDir, &totalAgentsExtracted)
}()
}

// Send all chunks to workers
for _, chunkFile := range chunkFiles {
jobChan <- chunkFile
}
close(jobChan)

// Wait for all processing to complete
jobWg.Wait()

return nil
}

// Get all XML chunk files from the input directory
func getChunkFiles(inputDir string) ([]string, error) {
files, err := os.ReadDir(inputDir)
if err != nil {
return nil, fmt.Errorf("failed to read directory: %w", err)
}

var chunkFiles []string
for _, file := range files {
if !file.IsDir() && strings.HasSuffix(file.Name(), ".xml") {
chunkFiles = append(chunkFiles, filepath.Join(inputDir, file.Name()))
}
}

return chunkFiles, nil
}

// processChunks reads XML chunks and extracts agent data
func processChunks(ctx context.Context, jobChan <-chan string, outputDir string, totalAgentsPtr *int32) {
for {
select {
case <-ctx.Done():
return
case chunkFile, ok := <-jobChan:
if !ok {
return // Channel closed, no more jobs
}

// Process a single chunk
agentsExtracted, err := processChunk(chunkFile, outputDir)
if err != nil {
// Log error but continue processing other chunks
fmt.Printf("Error processing chunk %s: %v\n", chunkFile, err)
continue
}

// Update total agents count
totalAgents := int(atomic.AddInt32(totalAgentsPtr, int32(agentsExtracted)))

// Use the proper agent counter function for Step 2
display.SetAgentCounter(totalAgents)

// Extract chunk number from filename for tracking
baseName := filepath.Base(chunkFile)
var chunkNum int
fmt.Sscanf(baseName, "chunk_%04d.xml", &chunkNum)

// Update processed chunks (current and total)
processedChunks := getProcessedChunks() + 1
incrementProcessedChunks()

// Get the total number of chunks we're processing
totalChunks := len(jobChan) + processedChunks

// Use the proper chunk counter function for Step 2
display.SetChunkCounter(processedChunks, totalChunks)
}
}
}

// Track processed chunks with an atomic counter
var processedChunksCounter int32

// Get current processed chunks count
func getProcessedChunks() int {
return int(atomic.LoadInt32(&processedChunksCounter))
}

// Increment processed chunks count
func incrementProcessedChunks() {
atomic.AddInt32(&processedChunksCounter, 1)
}

// Process a single XML chunk to extract agent data
func processChunk(chunkFile, outputDir string) (int, error) {
// Open the chunk file
file, err := os.Open(chunkFile)
if err != nil {
return 0, fmt.Errorf("failed to open chunk file: %w", err)
}
defer file.Close()

// Create a decoder for the XML
decoder := xml.NewDecoder(file)

// Extract chunk number from filename for the output file
baseName := filepath.Base(chunkFile)
var chunkNum int
fmt.Sscanf(baseName, "chunk_%04d.xml", &chunkNum)

// Prepare output index file
indexFileName := filepath.Join(outputDir, fmt.Sprintf("index_%04d.idx", chunkNum))
indexFile, err := os.Create(indexFileName)
if err != nil {
return 0, fmt.Errorf("failed to create index file: %w", err)
}
defer indexFile.Close()

writer := bufio.NewWriter(indexFile)
defer writer.Flush()

var (
currentPerson Person
inPerson     bool
agentCount   int
personXML    strings.Builder
)

// Process XML tokens
for {
token, err := decoder.Token()
if err == io.EOF {
break
}
if err != nil {
return agentCount, fmt.Errorf("error reading XML token: %w", err)
}

switch t := token.(type) {
case xml.StartElement:
if t.Name.Local == "person" {
inPerson = true
personXML.Reset()
personXML.WriteString(fmt.Sprintf("<%s", t.Name.Local))
for _, attr := range t.Attr {
personXML.WriteString(fmt.Sprintf(" %s=\"%s\"", attr.Name.Local, attr.Value))
if attr.Name.Local == "id" {
currentPerson.ID = attr.Value
}
}
personXML.WriteString(">")

// Unmarshal the person element
if err := decoder.DecodeElement(&currentPerson, &t); err != nil {
// Log error but continue with next person
fmt.Printf("Error decoding person: %v\n", err)
inPerson = false
continue
}

// Extract selected plan and first activity coordinates
coords, err := extractCoordinates(&currentPerson)
if err != nil {
// Log error but continue with next person
fmt.Printf("Error extracting coordinates for agent %s: %v\n", currentPerson.ID, err)
continue
}

// Convert coordinates from MTM8 to WGS84
longitude, latitude := ConvertMTM8ToWGS84(coords[0], coords[1])

// Write to index file
_, err = writer.WriteString(fmt.Sprintf("%s: %f, %f\n", currentPerson.ID, longitude, latitude))
if err != nil {
return agentCount, fmt.Errorf("failed to write to index file: %w", err)
}

agentCount++

// Update display every 1000 agents
if agentCount%1000 == 0 {
// Use the proper agent counter function
display.SetAgentCounter(agentCount)
}

inPerson = false
} else if inPerson {
personXML.WriteString(fmt.Sprintf("<%s", t.Name.Local))
for _, attr := range t.Attr {
personXML.WriteString(fmt.Sprintf(" %s=\"%s\"", attr.Name.Local, attr.Value))
}
personXML.WriteString(">")
}
case xml.EndElement:
if inPerson && t.Name.Local == "person" {
personXML.WriteString(fmt.Sprintf("</%s>", t.Name.Local))
inPerson = false
} else if inPerson {
personXML.WriteString(fmt.Sprintf("</%s>", t.Name.Local))
}
case xml.CharData:
if inPerson {
personXML.Write(t)
}
}
}

return agentCount, nil
}

// Extract coordinates from a person's selected plan and first activity
func extractCoordinates(person *Person) ([]float64, error) {
// Find the selected plan
var selectedPlan *Plan
for i := range person.Plans {
if person.Plans[i].Selected == "yes" {
selectedPlan = &person.Plans[i]
break
}
}

if selectedPlan == nil && len(person.Plans) > 0 {
// If no plan is explicitly selected, use the first one
selectedPlan = &person.Plans[0]
}

if selectedPlan == nil || len(selectedPlan.Activities) == 0 {
return nil, fmt.Errorf("no valid plan or activities found")
}

// Get coordinates from the first activity
firstActivity := selectedPlan.Activities[0]
return []float64{firstActivity.X, firstActivity.Y}, nil
}

// StoreAgentsInDatabase stores agent data in the database
func StoreAgentsInDatabase(indexDir string) error {
// Connect to the database
db, err := connectToDatabase()
if err != nil {
return fmt.Errorf("failed to connect to database: %w", err)
}
defer db.Close()

// Create agents table if it doesn't exist
if err := createAgentsTable(db); err != nil {
return fmt.Errorf("failed to create agents table: %w", err)
}

// Get all index files
files, err := os.ReadDir(indexDir)
if err != nil {
return fmt.Errorf("failed to read index directory: %w", err)
}

var indexFiles []string
for _, file := range files {
if !file.IsDir() && strings.HasSuffix(file.Name(), ".idx") {
indexFiles = append(indexFiles, filepath.Join(indexDir, file.Name()))
}
}

// Track database operations progress
totalOperations := len(indexFiles)
currentOperation := 0

// Initialize the database counter
display.SetDbCounter(currentOperation, totalOperations)

// Process each index file and insert data into the database
for i, indexFile := range indexFiles {
if err := processIndexFile(db, indexFile); err != nil {
return fmt.Errorf("failed to process index file %s: %w", indexFile, err)
}

// Update progress
currentOperation = i + 1
display.SetDbCounter(currentOperation, totalOperations)
}

return nil
}

// Process a single index file and insert data into the database
func processIndexFile(db *sql.DB, indexFile string) error {
// Open the index file
file, err := os.Open(indexFile)
if err != nil {
return fmt.Errorf("failed to open index file: %w", err)
}
defer file.Close()

scanner := bufio.NewScanner(file)

// Collect agents in batches for efficient database operations
var agents []Agent
batchSize := 1000

for scanner.Scan() {
line := scanner.Text()
parts := strings.Split(line, ":")
if len(parts) != 2 {
continue // Skip malformed lines
}

agentID := strings.TrimSpace(parts[0])
coordParts := strings.Split(strings.TrimSpace(parts[1]), ",")
if len(coordParts) != 2 {
continue // Skip malformed lines
}

var longitude, latitude float64
_, err := fmt.Sscanf(coordParts[0], "%f", &longitude)
if err != nil {
continue // Skip malformed lines
}

_, err = fmt.Sscanf(coordParts[1], "%f", &latitude)
if err != nil {
continue // Skip malformed lines
}

// Create agent and add to batch
agent := Agent{
ID:        agentID,
Longitude: longitude,
Latitude:  latitude,
}

agents = append(agents, agent)

// When batch is full, insert into database
if len(agents) >= batchSize {
if err := insertAgentBatch(db, agents); err != nil {
return fmt.Errorf("failed to insert agent batch: %w", err)
}

// Clear batch for next round
agents = agents[:0]
}
}

// Insert any remaining agents
if len(agents) > 0 {
if err := insertAgentBatch(db, agents); err != nil {
return fmt.Errorf("failed to insert remaining agent batch: %w", err)
}
}

return nil
}
