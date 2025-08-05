package tui

import "fmt"

// RenderHeader creates the title section with elapsed time
func RenderHeader(state *State) string {
	ezPart := "[EZ-UTILS]"
	descPart := ": PT Processing "
	timeStr := fmt.Sprintf("[ELAPSED TIME %02d:%02d]",
		int(state.ElapsedTime.Minutes()),
		int(state.ElapsedTime.Seconds())%60)

	titleContent := Styles.TitleEZ.Render(ezPart) + Styles.TitleText.Render(descPart)
	rightContent := Styles.TimeInfo.Render(timeStr)

	return titleContent + rightContent
}
