package processing

import (
	"fmt"
	"os"
	"path/filepath"
	
)

func (pto *PTOrchestrator) cleanupFiles() error {
	pto.logMessage("Starting Step 9: Clean Temporary Files")
	
	// Initialize live updates for cleanup
	if pto.tuiEnabled && pto.displayInstance != nil {
		_ = pto.displayInstance.SetLiveUpdate(8, "files", "0")
		_ = pto.displayInstance.SetLiveUpdate(8, "dirs", "0")
		_ = pto.displayInstance.SetLiveUpdate(8, "bytes", "0 MB")
	}
	
	// Clean up temporary directory and all its contents
	tempDir := pto.config.TempDir
	
	// Calculate directory size before cleanup (for stats)
	var totalSize int64
	var fileCount, dirCount int
	
	err := filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			dirCount++
		} else {
			fileCount++
			totalSize += info.Size()
		}
		return nil
	})
	
	if err != nil {
		pto.logMessage(fmt.Sprintf("Error calculating cleanup stats: %v", err))
		// Continue with cleanup even if stats calculation fails
	}
	
	// Remove the entire temporary directory
	if err := os.RemoveAll(tempDir); err != nil {
		pto.logMessage(fmt.Sprintf("Failed to remove temporary directory %s: %v", tempDir, err))
		return NewDirectoryError(tempDir, "remove")
	}
	
	// Update cleanup stats for display integration
	pto.stats.CleanupFiles = fileCount
	pto.stats.CleanupDirs = dirCount
	pto.stats.CleanupBytes = totalSize
	
	// Final live update with cleanup results
	if pto.tuiEnabled && pto.displayInstance != nil {
		_ = pto.displayInstance.SetLiveUpdate(8, "files", fmt.Sprintf("%d", fileCount))
		_ = pto.displayInstance.SetLiveUpdate(8, "dirs", fmt.Sprintf("%d", dirCount))
		sizeMB := float64(totalSize) / (1024 * 1024)
		_ = pto.displayInstance.SetLiveUpdate(8, "bytes", fmt.Sprintf("%.1f MB", sizeMB))
	}
	
	pto.logMessage(fmt.Sprintf("Cleanup completed successfully. Removed %d files and %d directories (%.2f KB)", 
		fileCount, dirCount, float64(totalSize)/1024.0))
	return nil
}