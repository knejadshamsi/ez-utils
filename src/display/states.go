package display

// Population module step labels
var PopulationStepLabels = map[int]string{
	1:  "Create Population Chunks",
	2:  "Fix XML Structure",
	3:  "Extract Agent Locations",
	4:  "Store Agent Data in Database",
	5:  "Create Network",
	6:  "Analyze Network",
	7:  "Create Boundary Polygon",
	8:  "Divide into Bins",
	9:  "Load Bin Configuration",
	10: "Assign Agents to Bins",
	11: "Validate Bins",
	12: "Calculate Bin Statistics",
	13: "Determine Scale Requirements",
	14: "Create Scaled Selection",
	15: "Create Scale Index",
	16: "Process Scaled Chunks",
	17: "Generate Final Files",
	18: "Update Database Scale Records",
	19: "Clean Temporary Files",
}

func GetStepLabel(module string, step int) string {
	switch module {
	case "POPULATION":
		if label, exists := PopulationStepLabels[step]; exists {
			return label
		}
		// TODO:  impliment addtional modules here
	}
	return "Processing Step"
}
