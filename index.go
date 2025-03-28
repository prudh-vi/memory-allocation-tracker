package main

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
}

// Color scheme for the UI
var (
	// Professional color scheme
	primaryColor   = lipgloss.Color("#0366d6")
	secondaryColor = lipgloss.Color("#28a745")
	accentColor    = lipgloss.Color("#6f42c1")
	warningColor   = lipgloss.Color("#f9c513")
	errorColor     = lipgloss.Color("#d73a49")
	darkBg         = lipgloss.Color("#1e1e1e")
	lightText      = lipgloss.Color("#f0f6fc")
	grayText       = lipgloss.Color("#8b949e")

	// UI component styles
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor).
		Background(darkBg).
		Padding(1, 3).
		Border(lipgloss.DoubleBorder()).
		BorderForeground(accentColor)

	boxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(secondaryColor).
		Padding(1, 2).
		Background(darkBg)

	infoStyle = lipgloss.NewStyle().
		Foreground(lightText).
		Bold(true).
		Padding(1, 0).
		Border(lipgloss.NormalBorder()).
		BorderForeground(primaryColor)

	statsStyle = lipgloss.NewStyle().
		Foreground(lightText).
		Italic(true).
		Border(lipgloss.HiddenBorder()).
		Padding(0, 1)

	headerStyle = lipgloss.NewStyle().
		Foreground(accentColor).
		Bold(true).
		Padding(0, 1)

	warningStyle = lipgloss.NewStyle().
		Foreground(warningColor).
		Background(darkBg).
		Bold(true)

	memoryBarStyle = lipgloss.NewStyle().
		Foreground(secondaryColor).
		Background(darkBg).
		Bold(true)
)

// Initialize the model
func initialModel() model {
	pageCount := 32 // Number of memory pages to display
	m := model{
		pages:    make([]page, pageCount),
		segments: []segment{},
	}
	m.stats.pageSize = 4 // 4KB page size
	return m
}

// Update function handles user input and updates the model
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "a":
			return m.allocateMemory(), nil
		case "d":
			return m.deallocateMemory(), nil
		}
	case tickMsg:
		// Update real-time memory stats
		if v, err := mem.VirtualMemory(); err == nil {
			m.stats.memoryUsage = v.UsedPercent
			m.stats.totalMemory = v.Total / 1024 / 1024
			m.stats.usedMemory = v.Used / 1024 / 1024
			m.stats.freeMemory = v.Free / 1024 / 1024
		}
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return tickMsg{}
		})
	}
	return m, nil
}

// View renders the current state of the model
func (m model) View() string {
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

	// Build UI components
	header := headerStyle.Render("[ MEMORY TRACKER v1.0 ][ SYSTEM MONITOR ACTIVE ][ REAL-TIME ANALYSIS ]")

	title := titleStyle.Render("📊 MEMORY ALLOCATION TRACKER")
	dateTime := statsStyle.Render(fmt.Sprintf("⏱️ %s", m.stats.currentTime.Format("15:04:05")))

	info := infoStyle.Render("[ A ] ALLOCATE MEMORY | [ D ] DEALLOCATE MEMORY | [ Q ] QUIT")

	// Add explanatory text for paging and segmentation
	pagingInfo := lipgloss.NewStyle().
		Foreground(lightText).
		Border(lipgloss.NormalBorder()).
		BorderForeground(secondaryColor).
		Padding(1).
		Render("📄 PAGING: Fixed-size memory blocks (pages) allocated to processes. Each page shows its process ID when allocated.")

	segmentationInfo := lipgloss.NewStyle().
		Foreground(lightText).
		Border(lipgloss.NormalBorder()).
		BorderForeground(primaryColor).
		Padding(1).
		Render("🧩 SEGMENTATION: Variable-size memory blocks (segments) showing process ID and size in KB.")

	warning := ""
	if m.stats.memoryUsage > 80 {
		warning = warningStyle.Render("⚠ WARNING: HIGH SYSTEM MEMORY USAGE DETECTED ⚠")
	}

	systemStats := statsStyle.Render(fmt.Sprintf(
		"💻 SYSTEM MEMORY\n"+
			"├─ Total: %d MB\n"+
			"├─ Used: %d MB\n"+
			"└─ Free: %d MB",
		m.stats.totalMemory,
		m.stats.usedMemory,
		m.stats.freeMemory,
	))

	memoryStats := statsStyle.Render(fmt.Sprintf(
		"📈 MEMORY METRICS\n"+
			"├─ Page Size: %d KB\n"+
			"├─ Fragmentation: %.1f%%\n"+
			"├─ Peak Usage: %.1f%%\n"+
			"└─ Available Pages: %d",
		m.stats.pageSize,
		fragmentation,
		m.stats.peakMemoryUsage,
		len(m.pages)-usedPages,
	))

	operations := statsStyle.Render(fmt.Sprintf(
		"🔄 OPERATIONS\n"+
			"├─ Allocations: %d\n"+
			"└─ Deallocations: %d",
		m.stats.totalAllocations,
		m.stats.totalDeallocations,
	))

	// Memory usage bar
	memBar := createMemoryBar(memoryUsage)

	// Pages and segments visualization
	// Pages visualization with clearer title
	pages := boxStyle.Render("📄 MEMORY PAGES (PAGING)\n" + renderPages(m.pages))
	
	// Segments visualization with clearer title
	segments := boxStyle.Render("🧩 MEMORY SEGMENTS (SEGMENTATION)\n" + renderSegments(m.segments))

	// Separator line
	separator := strings.Repeat("─", 80)

	// Combine all components
	return lipgloss.JoinVertical(lipgloss.Center,
		header,
		separator,
		title,
		dateTime,
		separator,
		info,
		warning,
		separator,
		lipgloss.JoinHorizontal(lipgloss.Top, systemStats, "   ", memoryStats, "   ", operations),
		separator,
		memBar,
		separator,
		pagingInfo,  // Add paging explanation
		pages,
		separator,
		segmentationInfo,  // Add segmentation explanation
		segments,
	)
}

// Initialize the application
func (m model) Init() tea.Cmd {
	return tea.Batch(
		tick(),
		tea.EnterAltScreen,
	)
}

// Ticker for real-time updates
func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

type tickMsg struct{}

// Allocate memory with contiguous pages
func (m model) allocateMemory() model {
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

	return m
}

// Deallocate memory
func (m model) deallocateMemory() model {
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
	return m
}

// Render memory pages visualization with enhanced differentiation
func renderPages(pages []page) string {
	var output []string
	output = append(output, lipgloss.NewStyle().
		Foreground(lightText).
		Render("Each [ ] represents a fixed-size page (4KB):"))
	
	for i, p := range pages {
		if i > 0 && i%8 == 0 {
			output = append(output, "\n")
		}
		
		if p.used {
			output = append(output, lipgloss.NewStyle().
				Foreground(secondaryColor).
				Bold(true).
				Render(fmt.Sprintf("[%s]", p.processID)))
		} else {
			output = append(output, lipgloss.NewStyle().
				Foreground(grayText).
				Render("[ ]"))
		}
	}
	return strings.Join(output, " ")
}

// Render memory segments visualization with enhanced differentiation
func renderSegments(segments []segment) string {
	if len(segments) == 0 {
		return lipgloss.NewStyle().
			Foreground(grayText).
			Italic(true).
			Render("No active memory segments")
	}
	
	var output []string
	output = append(output, lipgloss.NewStyle().
		Foreground(lightText).
		Render("Each ⟨Process:Size⟩ represents a variable-size segment:"))
	
	for i, s := range segments {
		if i > 0 && i%4 == 0 {
			output = append(output, "\n")
		}
		
		output = append(output, lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			Render(fmt.Sprintf("⟨%s:%dKB⟩", s.processID, s.size)))
	}
	return strings.Join(output, " ")
}

// Create a visual memory usage bar
func createMemoryBar(usage float64) string {
	width := 70
	filled := int(usage * float64(width) / 100)
	
	// Create the bar with gradient colors
	var bar strings.Builder
	for i := 0; i < width; i++ {
		if i < filled {
			if i < width/3 {
				bar.WriteString(lipgloss.NewStyle().Foreground(secondaryColor).Render("█"))
			} else if i < width*2/3 {
				bar.WriteString(lipgloss.NewStyle().Foreground(warningColor).Render("█"))
			} else {
				bar.WriteString(lipgloss.NewStyle().Foreground(errorColor).Render("█"))
			}
		} else {
			bar.WriteString(lipgloss.NewStyle().Foreground(grayText).Render("░"))
		}
	}
	
	// Add percentage markers
	markers := fmt.Sprintf("[0%%]%s[50%%]%s[100%%]", 
		strings.Repeat(" ", width/2-3),
		strings.Repeat(" ", width/2-4))
	
	return boxStyle.Render(fmt.Sprintf("MEMORY USAGE: %.1f%%\n%s\n%s", 
		usage, bar.String(), markers))
}

func main() {
	rand.Seed(time.Now().UnixNano())
	
	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		log.Fatal("Error running program:", err)
	}
}