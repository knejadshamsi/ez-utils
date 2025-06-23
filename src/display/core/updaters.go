package core

import (
	"fmt"

	"ez-utils/src/display/tui"
)

// SetLiveUpdate sets a live update value for a specific step
func (d *DisplayInstance) SetLiveUpdate(stepNum int, key string, value interface{}) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if !d.running {
		return fmt.Errorf("display instance is not running")
	}

	// Check if this step supports this live update
	if !d.config.HasLiveUpdate(stepNum, key) {
		return fmt.Errorf("step %d does not support live update '%s'", stepNum, key)
	}

	// Store the update
	if d.liveUpdates[stepNum] == nil {
		d.liveUpdates[stepNum] = make(map[string]interface{})
	}
	d.liveUpdates[stepNum][key] = value

	// Send async update
	return d.asyncManager.SendLiveUpdate(stepNum, key, value)
}

// SetStepMessage sets a custom message for a specific step
func (d *DisplayInstance) SetStepMessage(stepNum int, template string, data map[string]interface{}) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if !d.running {
		return fmt.Errorf("display instance is not running")
	}

	// Store the message
	d.stepMessages[stepNum] = template

	// Send async update
	return d.asyncManager.SendStepMessage(stepNum, template, data)
}

// SetStep updates the current step number (async)
func (d *DisplayInstance) SetStep(stepNumber int) error {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if !d.running {
		return fmt.Errorf("display instance is not running")
	}

	// Validate step transition
	if err := d.businessLogic.ValidateStepTransition(stepNumber); err != nil {
		// Log warning but continue
		fmt.Printf("Step transition warning: %v\n", err)
	}

	// Update business logic
	d.businessLogic.UpdateStep(stepNumber)

	// Send async update
	return d.asyncManager.SendStepUpdate(stepNumber)
}

// SetProcessComplete marks the process as complete
func (d *DisplayInstance) SetProcessComplete(complete bool) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if !d.running || d.program == nil {
		return fmt.Errorf("display instance is not running")
	}

	// Update business logic
	d.businessLogic.SetProcessComplete(complete)

	// Send UI update
	d.program.Send(tui.ProcessCompleteMsg(complete))
	return nil
}

// XML Processing Functions (Steps 0-1) - Chunk parsing and validation

// SetChunkCount updates the chunk count
func (d *DisplayInstance) SetChunkCount(count int) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if !d.running || d.program == nil {
		return fmt.Errorf("display instance is not running")
	}

	// Update business logic
	d.businessLogic.SetChunkCount(count)

	// Send UI update
	d.program.Send(tui.ChunkCountMsg(count))
	return nil
}

// SetPersonsFound updates the persons found count
func (d *DisplayInstance) SetPersonsFound(count int) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if !d.running || d.program == nil {
		return fmt.Errorf("display instance is not running")
	}

	// Update business logic
	d.businessLogic.SetPersonsFound(count)

	// Send UI update
	d.program.Send(tui.PersonsFoundMsg(count))
	return nil
}

// SetBytesReadMB updates the bytes read in MB
func (d *DisplayInstance) SetBytesReadMB(mb int) error {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if !d.running || d.program == nil {
		return fmt.Errorf("display instance is not running")
	}

	d.program.Send(tui.BytesReadMsg(mb))
	return nil
}

// SetFirstChunkStatus updates the first chunk fixed status
func (d *DisplayInstance) SetFirstChunkStatus(fixed bool) error {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if !d.running || d.program == nil {
		return fmt.Errorf("display instance is not running")
	}

	d.program.Send(tui.FirstChunkFixedMsg(fixed))
	return nil
}

// SetLastChunkStatus updates the last chunk fixed status
func (d *DisplayInstance) SetLastChunkStatus(fixed bool) error {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if !d.running || d.program == nil {
		return fmt.Errorf("display instance is not running")
	}

	d.program.Send(tui.LastChunkFixedMsg(fixed))
	return nil
}

// Location Extraction Functions (Steps 2-3) - Geographic data processing

// SetAgentCounter updates the agent counter
func (d *DisplayInstance) SetAgentCounter(count int) error {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if !d.running || d.program == nil {
		return fmt.Errorf("display instance is not running")
	}

	d.program.Send(tui.AgentCounterMsg(count))
	return nil
}

// SetChunkCounter updates the chunk processing counter
func (d *DisplayInstance) SetChunkCounter(current, total int) error {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if !d.running || d.program == nil {
		return fmt.Errorf("display instance is not running")
	}

	d.program.Send(tui.ChunkCounterMsg{Current: current, Total: total})
	return nil
}

// SetDbCounter updates the database processing counter
func (d *DisplayInstance) SetDbCounter(current, total int) error {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if !d.running || d.program == nil {
		return fmt.Errorf("display instance is not running")
	}

	d.program.Send(tui.DbCounterMsg{Current: current, Total: total})
	return nil
}

// Processing Functions (Steps 4-18) - Data transformation and cleanup

// SetCoordinateCounter updates the coordinate counter
func (d *DisplayInstance) SetCoordinateCounter(count int) error {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if !d.running || d.program == nil {
		return fmt.Errorf("display instance is not running")
	}

	d.program.Send(tui.CoordinateCounterMsg(count))
	return nil
}

// SetBinCounter updates the bin processing counter
func (d *DisplayInstance) SetBinCounter(current, total int) error {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if !d.running || d.program == nil {
		return fmt.Errorf("display instance is not running")
	}

	d.program.Send(tui.BinCounterMsg{Current: current, Total: total})
	return nil
}

// SetAgentBinCounter updates the agent bin counter
func (d *DisplayInstance) SetAgentBinCounter(count int) error {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if !d.running || d.program == nil {
		return fmt.Errorf("display instance is not running")
	}

	d.program.Send(tui.AgentBinCounterMsg(count))
	return nil
}

// SetScaleCounter updates the scale processing counter
func (d *DisplayInstance) SetScaleCounter(current, total int) error {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if !d.running || d.program == nil {
		return fmt.Errorf("display instance is not running")
	}

	d.program.Send(tui.ScaleCounterMsg{Current: current, Total: total})
	return nil
}

// SetOutputScale updates the output scale value
func (d *DisplayInstance) SetOutputScale(scale int) error {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if !d.running || d.program == nil {
		return fmt.Errorf("display instance is not running")
	}

	d.program.Send(tui.OutputScaleMsg(scale))
	return nil
}

// SetOutputCounter updates the output processing counter
func (d *DisplayInstance) SetOutputCounter(current, total int) error {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if !d.running || d.program == nil {
		return fmt.Errorf("display instance is not running")
	}

	d.program.Send(tui.OutputCounterMsg{Current: current, Total: total})
	return nil
}

// SetCleanupCounter updates the cleanup counter
func (d *DisplayInstance) SetCleanupCounter(files, dirs, bytes int) error {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if !d.running || d.program == nil {
		return fmt.Errorf("display instance is not running")
	}

	d.program.Send(tui.CleanupCounterMsg{Files: files, Dirs: dirs, Bytes: bytes})
	return nil
}