package main

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/shirou/gopsutil/mem"
)

// Data structures for memory representation
type page struct {
	used      bool
	processID string
}

type segment struct {
	size      int
	processID string
}

// Model represents the application state
type model struct {
	pages    []page
	segments []segment
	stats    struct {
		totalAllocations   int
		totalDeallocations int
		memoryUsage        float64
		totalMemory        uint64
		usedMemory         uint64
		freeMemory         uint64
		currentTime        time.Time
		pageSize           uint64
		fragmentationRate  float64
		peakMemoryUsage    float64
	}
	app *tview.Application
}

// Color scheme for the UI
var (
	// Matrix-inspired color scheme
	colorPrimary   = tcell.ColorLime
	colorSecondary = tcell.ColorAqua
	colorAccent    = tcell.ColorPink
	colorWarning   = tcell.ColorYellow
	colorError     = tcell.ColorRed
	colorBg        = tcell.ColorBlack
	colorText      = tcell.ColorWhite
	colorGrayText  = tcell.ColorGray
)

// Initialize the model
func initialModel() *model {
	pageCount := 64 // Increased number of memory pages to display
	m := &model{
		pages:    make([]page, pageCount),
		segments: []segment{},
		app:      tview.NewApplication(),
	}
	m.stats.pageSize = 4 // 4KB page size
	return m
}

// Create the main layout
func (m *model) createLayout() *tview.Grid {
	// Create the main grid layout
	grid := tview.NewGrid().
		SetRows(3, 1, 10, 1, 3, 1, 0).
		SetColumns(0, 0).
		SetBorders(false)

	// Create the title bar with Matrix-style animation
	titleBar := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText(renderMatrixTitle("MEMORY ALLOCATION TRACKER"))
	titleBar.SetBorderPadding(0, 0, 0, 0)
	titleBar.SetBackgroundColor(colorBg)

	// Create the status bar
	statusBar := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetChangedFunc(func() {
			m.app.Draw()
		})
	statusBar.SetBackgroundColor(colorBg)

	// Create the memory stats panel
	statsPanel := tview.NewGrid().
		SetRows(0).
		SetColumns(0, 0, 0).
		SetBorders(false)

	// System memory stats
	sysMemBox := tview.NewTextView().
		SetDynamicColors(true).
		SetChangedFunc(func() {
			m.app.Draw()
		})
	sysMemBox.SetBorder(true).
		SetBorderColor(colorSecondary).
		SetTitle(" SYSTEM MEMORY ").
		SetTitleColor(colorPrimary).
		SetTitleAlign(tview.AlignCenter).
		SetBackgroundColor(colorBg)

	// Memory metrics
	memMetricsBox := tview.NewTextView().
		SetDynamicColors(true).
		SetChangedFunc(func() {
			m.app.Draw()
		})
	memMetricsBox.SetBorder(true).
		SetBorderColor(colorPrimary).
		SetTitle(" MEMORY METRICS ").
		SetTitleColor(colorSecondary).
		SetTitleAlign(tview.AlignCenter).
		SetBackgroundColor(colorBg)

	// Operations stats
	opsBox := tview.NewTextView().
		SetDynamicColors(true).
		SetChangedFunc(func() {
			m.app.Draw()
		})
	opsBox.SetBorder(true).
		SetBorderColor(colorAccent).
		SetTitle(" OPERATIONS ").
		SetTitleColor(colorPrimary).
		SetTitleAlign(tview.AlignCenter).
		SetBackgroundColor(colorBg)

	statsPanel.AddItem(sysMemBox, 0, 0, 1, 1, 0, 0, false)
	statsPanel.AddItem(memMetricsBox, 0, 1, 1, 1, 0, 0, false)
	statsPanel.AddItem(opsBox, 0, 2, 1, 1, 0, 0, false)

	// Create the memory bar
	memBar := tview.NewTextView().
		SetDynamicColors(true).
		SetChangedFunc(func() {
			m.app.Draw()
		})
	memBar.SetBorder(true).
		SetBorderColor(colorPrimary).
		SetTitle(" MEMORY USAGE ").
		SetTitleColor(colorSecondary).
		SetTitleAlign(tview.AlignCenter).
		SetBackgroundColor(colorBg)

	// Create the memory visualization panel
	memVisPanel := tview.NewGrid().
		SetRows(0).
		SetColumns(0, 0).
		SetBorders(false)

	// Paging visualization
	pagingBox := tview.NewTextView().
		SetDynamicColors(true).
		SetChangedFunc(func() {
			m.app.Draw()
		})
	pagingBox.SetBorder(true).
		SetBorderColor(colorSecondary).
		SetTitle(" PAGING ").
		SetTitleColor(colorPrimary).
		SetTitleAlign(tview.AlignCenter).
		SetBackgroundColor(colorBg)

	// Segmentation visualization
	segmentationBox := tview.NewTextView().
		SetDynamicColors(true).
		SetChangedFunc(func() {
			m.app.Draw()
		})
	segmentationBox.SetBorder(true).
		SetBorderColor(colorAccent).
		SetTitle(" SEGMENTATION ").
		SetTitleColor(colorPrimary).
		SetTitleAlign(tview.AlignCenter).
		SetBackgroundColor(colorBg)

	memVisPanel.AddItem(pagingBox, 0, 0, 1, 1, 0, 0, false)
	memVisPanel.AddItem(segmentationBox, 0, 1, 1, 1, 0, 0, false)

	// Add all components to the main grid
	grid.AddItem(titleBar, 0, 0, 1, 2, 0, 0, false)
	grid.AddItem(tview.NewTextView().SetText(""), 1, 0, 1, 2, 0, 0, false) // Spacer
	grid.AddItem(statsPanel, 2, 0, 1, 2, 0, 0, false)
	grid.AddItem(tview.NewTextView().SetText(""), 3, 0, 1, 2, 0, 0, false) // Spacer
	grid.AddItem(memBar, 4, 0, 1, 2, 0, 0, false)
	grid.AddItem(tview.NewTextView().SetText(""), 5, 0, 1, 2, 0, 0, false) // Spacer
	grid.AddItem(memVisPanel, 6, 0, 1, 2, 0, 0, false)

	// Set up the update function for real-time stats
	go func() {
		for {
			m.updateStats()
			sysMemBox.SetText(m.renderSystemMemory())
			memMetricsBox.SetText(m.renderMemoryMetrics())
			opsBox.SetText(m.renderOperations())
			memBar.SetText(m.renderMemoryBar())
			pagingBox.SetText(m.renderPages())
			segmentationBox.SetText(m.renderSegments())
			statusBar.SetText(m.renderStatusBar())
			time.Sleep(500 * time.Millisecond)
		}
	}()

	// Set up key bindings
	m.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape, tcell.KeyCtrlC:
			m.app.Stop()
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q':
				m.app.Stop()
			case 'a':
				m.allocateMemory()
			case 'd':
				m.deallocateMemory()
			}
		}
		return event
	})

	return grid
}

// Matrix-style title rendering
func renderMatrixTitle(title string) string {
	result := "[green::b]"
	for _, char := range title {
		if rand.Intn(10) < 8 { // 80% chance to show the character
			result += string(char)
		} else {
			// Use a random Matrix-like character instead
			matrixChars := []rune{'0', '1', '|', '/', '\\', '_', '-', '+', '=', '*', '#', '@'}
			result += string(matrixChars[rand.Intn(len(matrixChars))])
		}
	}
	return result
}

// Update real-time memory stats
func (m *model) updateStats() {
	// Update current time
	m.stats.currentTime = time.Now()

	// Calculate memory usage for pages
	usedPages := 0
	for _, p := range m.pages {
		if p.used {
			usedPages++
		}
	}
	memoryUsage := float64(usedPages) / float64(len(m.pages)) * 100

	// Calculate fragmentation
	fragmentation := 0.0
	if len(m.segments) > 0 {
		gaps := 0
		for i := 1; i < len(m.segments); i++ {
			if m.segments[i-1].processID != m.segments[i].processID {
				gaps++
			}
		}
		fragmentation = float64(gaps) / float64(len(m.segments)) * 100
	}
	m.stats.fragmentationRate = fragmentation

	// Update peak memory usage
	if memoryUsage > m.stats.peakMemoryUsage {
		m.stats.peakMemoryUsage = memoryUsage
	}

	// Update system memory stats
	if v, err := mem.VirtualMemory(); err == nil {
		m.stats.memoryUsage = v.UsedPercent
		m.stats.totalMemory = v.Total / 1024 / 1024
		m.stats.usedMemory = v.Used / 1024 / 1024
		m.stats.freeMemory = v.Free / 1024 / 1024
	}
}

// Render system memory stats
func (m *model) renderSystemMemory() string {
	return fmt.Sprintf(
		"\n[yellow]Total:[white] %d MB\n"+
			"[yellow]Used:[white] %d MB\n"+
			"[yellow]Free:[white] %d MB\n"+
			"[yellow]Usage:[white] %.1f%%",
		m.stats.totalMemory,
		m.stats.usedMemory,
		m.stats.freeMemory,
		m.stats.memoryUsage,
	)
}

// Render memory metrics
func (m *model) renderMemoryMetrics() string {
	// Calculate memory usage for pages
	usedPages := 0
	for _, p := range m.pages {
		if p.used {
			usedPages++
		}
	}
	// Use this variable in the return statement instead of calculating again
	memoryUsage := float64(usedPages) / float64(len(m.pages)) * 100

	return fmt.Sprintf(
		"\n[yellow]Page Size:[white] %d KB\n"+
			"[yellow]Fragmentation:[white] %.1f%%\n"+
			"[yellow]Peak Usage:[white] %.1f%%\n"+
			"[yellow]Free Pages:[white] %d",
		m.stats.pageSize,
		m.stats.fragmentationRate,
		memoryUsage, // Use the calculated value here instead of recalculating
		len(m.pages)-usedPages,
	)
}

// In the renderSegments function, fix the undefined variable p
func (m *model) renderSegments() string {
	if len(m.segments) == 0 {
		return "[gray::i]No active memory segments"
	}

	result := "[yellow]Variable-size memory blocks[white]\n\n"
	result += "[yellow]Segments:[white] \n\n"

	// Visual representation of segments
	var row string
	for i, s := range m.segments {
		if i > 0 && i%8 == 0 {
			result += row + "\n"
			row = ""
		}

		// Size-proportional block with process ID
		blockSize := s.size / 4
		if blockSize < 1 {
			blockSize = 1
		}
		if blockSize > 5 {
			blockSize = 5
		}

		// Use different colors for different process IDs
		processNum := 0
		if len(s.processID) > 1 {
			processNum = int(s.processID[1] - '0') // Changed p.processID to s.processID
		}

		var blockColor string
		switch processNum % 5 {
		case 0:
			blockColor = "[green]"
		case 1:
			blockColor = "[aqua]"
		case 2:
			blockColor = "[pink]"
		case 3:
			blockColor = "[blue]"
		case 4:
			blockColor = "[yellow]"
		}

		block := strings.Repeat("█", blockSize)
		row += blockColor + block + "(" + s.processID + ") "
	}

	if row != "" {
		result += row + "\n"
	}

	// Add detailed table
	result += "\n[aqua::b]ID | Size | Status[white]\n"
	result += "[white]------------------[white]\n"

	for i, s := range m.segments {
		if i >= 8 && len(m.segments) > 9 {
			result += fmt.Sprintf("[gray]... and %d more[white]", len(m.segments)-8)
			break
		}

		// Use different colors for different process IDs
		processNum := 0
		if len(s.processID) > 1 {
			processNum = int(s.processID[1] - '0')
		}

		var idColor string
		switch processNum % 5 {
		case 0:
			idColor = "[green]"
		case 1:
			idColor = "[aqua]"
		case 2:
			idColor = "[pink]"
		case 3:
			idColor = "[blue]"
		case 4:
			idColor = "[yellow]"
		}

		result += fmt.Sprintf("%s%s[white] | [gray]%dKB[white] | [lime]Active[white]\n",
			idColor, s.processID, s.size)
	}

	return result
}

// Render operations stats
func (m *model) renderOperations() string {
	return fmt.Sprintf(
		"\n[yellow]Allocations:[white] %d\n"+
			"[yellow]Deallocations:[white] %d\n"+
			"[yellow]Time:[white] %s",
		m.stats.totalAllocations,
		m.stats.totalDeallocations,
		m.stats.currentTime.Format("15:04:05"),
	)
}

// Render memory bar
func (m *model) renderMemoryBar() string {
	// Calculate memory usage for pages
	usedPages := 0
	for _, p := range m.pages {
		if p.used {
			usedPages++
		}
	}
	memoryUsage := float64(usedPages) / float64(len(m.pages)) * 100

	width := 100
	filled := int(memoryUsage * float64(width) / 100)

	bar := "\n[yellow]Memory Usage: [white]" + fmt.Sprintf("%.1f%% ", memoryUsage) + "\n\n["

	for i := 0; i < width; i++ {
		if i < filled {
			if i < width/3 {
				bar += "[green]█"
			} else if i < width*2/3 {
				bar += "[yellow]█"
			} else {
				bar += "[red]█"
			}
		} else {
			bar += "[gray]░"
		}
	}
	bar += "]"

	return bar
}

// Render pages visualization
func (m *model) renderPages() string {
	var result string
	result = "[yellow]Fixed-size memory blocks (4KB each)[white]\n\n"

	// Create rows of pages with more per row and vibrant colors
	var row string
	for i, p := range m.pages {
		if i > 0 && i%16 == 0 {
			result += row + "\n"
			row = ""
		}

		if p.used {
			// Use different colors for different process IDs
			processNum := 0
			if len(p.processID) > 1 {
				processNum = int(p.processID[1] - '0')
			}

			var blockColor string
			switch processNum % 5 {
			case 0:
				blockColor = "[green]"
			case 1:
				blockColor = "[aqua]"
			case 2:
				blockColor = "[pink]"
			case 3:
				blockColor = "[blue]"
			case 4:
				blockColor = "[yellow]"
			}

			row += blockColor + "■ "
		} else {
			row += "[gray]□ "
		}
	}

	if row != "" {
		result += row + "\n"
	}

	// Add legend
	result += "\n[green]■[white] allocated page  [gray]□[white] free page"

	return result
}

// Render status bar
func (m *model) renderStatusBar() string {
	return "[green]a[white]:allocate | [green]d[white]:deallocate | [green]q[white]:quit"
}

// Allocate memory with contiguous pages
func (m *model) allocateMemory() {
	pid := fmt.Sprintf("P%d", rand.Intn(100))

	// Try to find contiguous free pages
	startIdx := -1
	contiguousCount := 0
	requiredPages := rand.Intn(3) + 1 // Allocate 1-3 pages at once

	for i := range m.pages {
		if !m.pages[i].used {
			if startIdx == -1 {
				startIdx = i
			}
			contiguousCount++
			if contiguousCount >= requiredPages {
				break
			}
		} else {
			startIdx = -1
			contiguousCount = 0
		}
	}

	// Allocate contiguous pages if found
	if startIdx != -1 && contiguousCount >= requiredPages {
		for i := startIdx; i < startIdx+requiredPages; i++ {
			m.pages[i] = page{true, pid}
		}
		m.segments = append(m.segments, segment{requiredPages * int(m.stats.pageSize), pid})
		m.stats.totalAllocations++
	}
}

// Deallocate memory
func (m *model) deallocateMemory() {
	if len(m.segments) > 0 {
		// Get the process ID of the last segment
		lastPID := m.segments[len(m.segments)-1].processID

		// Remove the segment
		m.segments = m.segments[:len(m.segments)-1]

		// Free pages with matching process ID
		for i := range m.pages {
			if m.pages[i].processID == lastPID {
				m.pages[i] = page{false, ""}
			}
		}

		m.stats.totalDeallocations++
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	m := initialModel()
	
	// Create the UI layout
	grid := m.createLayout()
	
	// Start the application
	if err := m.app.SetRoot(grid, true).EnableMouse(true).Run(); err != nil {
		log.Fatal("Error running application:", err)
	}
}