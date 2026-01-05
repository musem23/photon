package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mahamedmuse/photon/internal/config"
	"github.com/mahamedmuse/photon/internal/image"
)

type state int

const (
	stateMenu state = iota
	stateSelectInput
	stateSelectOutput
	stateSelectFormat
	stateQuality
	stateConfirm
	stateConverting
	stateComplete
	stateSettings
	stateSelectOutputDir
	stateBatchSelect
	stateBatchConfirm
	stateBatchConverting
	stateBatchComplete
)

type fileEntry struct {
	name     string
	path     string
	isDir    bool
	isImg    bool
	selected bool
}

type Model struct {
	state         state
	config        config.Config
	width, height int

	// Menu
	menuIndex int
	menuItems []string

	// File browser
	currentDir   string
	files        []fileEntry
	fileIndex    int
	inputFile    string
	outputFile   string
	scrollOffset int

	// Format selection
	formats      []string
	formatIndex  int
	outputFormat string

	// Quality
	quality int

	// Conversion
	spinner    spinner.Model
	converting bool
	convErr    error
	convDone   bool

	// Settings
	settingIndex int

	// Batch mode
	batchMode      bool
	selectedFiles  []string
	batchOutputDir string
	batchResults   []batchResult
	batchIndex     int
}

type batchResult struct {
	input   string
	output  string
	success bool
	err     error
}

func NewModel() Model {
	cfg, _ := config.Load()

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED"))

	return Model{
		state:     stateMenu,
		config:    cfg,
		menuIndex: 0,
		menuItems: []string{
			"🖼  Convert Image",
			"📚 Batch Convert",
			"🕐 Recent Files",
			"⚙  Settings",
			"🚪 Quit",
		},
		formats:    []string{"png", "jpg", "gif", "webp", "bmp", "tiff", "avif"},
		quality:    cfg.DefaultQuality,
		spinner:    s,
		currentDir: cfg.LastInputDir,
	}
}

func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.state == stateMenu {
				m.config.Save()
				return m, tea.Quit
			}
			m.state = stateMenu
			return m, nil

		case "esc":
			if m.state != stateMenu {
				m.state = stateMenu
			}
			return m, nil
		}

		switch m.state {
		case stateMenu:
			return m.updateMenu(msg)
		case stateSelectInput:
			return m.updateFileBrowser(msg, true)
		case stateSelectOutput:
			return m.updateFileBrowser(msg, false)
		case stateSelectFormat:
			return m.updateFormatSelect(msg)
		case stateQuality:
			return m.updateQuality(msg)
		case stateConfirm:
			return m.updateConfirm(msg)
		case stateSettings:
			return m.updateSettings(msg)
		case stateSelectOutputDir:
			return m.updateOutputDirBrowser(msg)
		case stateBatchSelect:
			return m.updateBatchSelect(msg)
		case stateBatchConfirm:
			return m.updateBatchConfirm(msg)
		case stateBatchComplete:
			if msg.String() != "" {
				m.state = stateMenu
			}
			return m, nil
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case conversionDoneMsg:
		m.converting = false
		m.convDone = true
		m.convErr = msg.err
		m.state = stateComplete
		if msg.err == nil {
			m.config.AddRecentFile(m.inputFile)
			m.config.Save()
		}
		return m, nil

	case batchDoneMsg:
		m.converting = false
		m.batchResults = msg.results
		m.state = stateBatchComplete
		return m, nil
	}

	return m, nil
}

func (m Model) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.menuIndex > 0 {
			m.menuIndex--
		}
	case "down", "j":
		if m.menuIndex < len(m.menuItems)-1 {
			m.menuIndex++
		}
	case "enter":
		switch m.menuIndex {
		case 0: // Convert Image
			m.batchMode = false
			m.loadFiles(m.currentDir)
			m.state = stateSelectInput
		case 1: // Batch Convert
			m.batchMode = true
			m.selectedFiles = []string{}
			m.loadFiles(m.currentDir)
			m.state = stateBatchSelect
		case 3: // Settings
			m.state = stateSettings
		case 4: // Quit
			m.config.Save()
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *Model) loadFiles(dir string) {
	m.currentDir = dir
	m.files = []fileEntry{}
	m.fileIndex = 0
	m.scrollOffset = 0

	if dir != "/" {
		m.files = append(m.files, fileEntry{
			name:  "..",
			path:  filepath.Dir(dir),
			isDir: true,
		})
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	imageExts := map[string]bool{
		".png": true, ".jpg": true, ".jpeg": true, ".gif": true,
		".webp": true, ".bmp": true, ".tiff": true, ".tif": true,
		".avif": true, ".heic": true, ".heif": true,
	}

	var dirs, files []fileEntry
	for _, e := range entries {
		if !m.config.ShowHiddenFiles && strings.HasPrefix(e.Name(), ".") {
			continue
		}

		entry := fileEntry{
			name:  e.Name(),
			path:  filepath.Join(dir, e.Name()),
			isDir: e.IsDir(),
		}

		if !e.IsDir() {
			ext := strings.ToLower(filepath.Ext(e.Name()))
			entry.isImg = imageExts[ext]
		}

		if e.IsDir() {
			dirs = append(dirs, entry)
		} else {
			files = append(files, entry)
		}
	}

	m.files = append(m.files, dirs...)
	m.files = append(m.files, files...)
}

func (m Model) updateFileBrowser(msg tea.KeyMsg, isInput bool) (tea.Model, tea.Cmd) {
	maxVisible := m.height - 15
	if maxVisible < 5 {
		maxVisible = 5
	}

	switch msg.String() {
	case "up", "k":
		if m.fileIndex > 0 {
			m.fileIndex--
			if m.fileIndex < m.scrollOffset {
				m.scrollOffset = m.fileIndex
			}
		}
	case "down", "j":
		if m.fileIndex < len(m.files)-1 {
			m.fileIndex++
			if m.fileIndex >= m.scrollOffset+maxVisible {
				m.scrollOffset = m.fileIndex - maxVisible + 1
			}
		}
	case "enter":
		if m.fileIndex < len(m.files) {
			entry := m.files[m.fileIndex]
			if entry.isDir {
				m.loadFiles(entry.path)
			} else if isInput && entry.isImg {
				m.inputFile = entry.path
				m.config.LastInputDir = m.currentDir
				m.state = stateSelectFormat
			}
		}
	case "tab":
		m.config.ShowHiddenFiles = !m.config.ShowHiddenFiles
		m.loadFiles(m.currentDir)
	}
	return m, nil
}

func (m Model) updateFormatSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k", "left", "h":
		if m.formatIndex > 0 {
			m.formatIndex--
		}
	case "down", "j", "right", "l":
		if m.formatIndex < len(m.formats)-1 {
			m.formatIndex++
		}
	case "enter":
		m.outputFormat = m.formats[m.formatIndex]
		ext := filepath.Ext(m.inputFile)
		base := strings.TrimSuffix(filepath.Base(m.inputFile), ext)
		if err := m.config.EnsureOutputDir(); err == nil {
			m.outputFile = filepath.Join(m.config.OutputDir, base+"."+m.outputFormat)
		} else {
			m.outputFile = filepath.Join(filepath.Dir(m.inputFile), base+"."+m.outputFormat)
		}
		m.state = stateQuality
	}
	return m, nil
}

func (m Model) updateQuality(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "left", "h":
		if m.quality > 1 {
			m.quality -= 5
			if m.quality < 1 {
				m.quality = 1
			}
		}
	case "right", "l":
		if m.quality < 100 {
			m.quality += 5
			if m.quality > 100 {
				m.quality = 100
			}
		}
	case "enter":
		if m.batchMode {
			m.state = stateBatchConfirm
		} else {
			m.state = stateConfirm
		}
	}
	return m, nil
}

func (m Model) updateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "enter":
		m.state = stateConverting
		m.converting = true
		return m, m.doConvert()
	case "n", "esc":
		m.state = stateMenu
	}
	return m, nil
}

func (m Model) updateOutputDirBrowser(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	maxVisible := m.height - 15
	if maxVisible < 5 {
		maxVisible = 5
	}

	switch msg.String() {
	case "up", "k":
		if m.fileIndex > 0 {
			m.fileIndex--
			if m.fileIndex < m.scrollOffset {
				m.scrollOffset = m.fileIndex
			}
		}
	case "down", "j":
		if m.fileIndex < len(m.files)-1 {
			m.fileIndex++
			if m.fileIndex >= m.scrollOffset+maxVisible {
				m.scrollOffset = m.fileIndex - maxVisible + 1
			}
		}
	case "enter":
		if m.fileIndex < len(m.files) {
			entry := m.files[m.fileIndex]
			if entry.isDir {
				m.loadFiles(entry.path)
			}
		}
	case "s":
		m.config.OutputDir = m.currentDir
		m.config.Save()
		m.state = stateSettings
	case "tab":
		m.config.ShowHiddenFiles = !m.config.ShowHiddenFiles
		m.loadFiles(m.currentDir)
	}
	return m, nil
}

func (m Model) updateSettings(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	settings := []string{"Default Quality", "Default Format", "Output Directory", "Show Hidden Files", "Confirm Overwrite", "Back"}

	switch msg.String() {
	case "up", "k":
		if m.settingIndex > 0 {
			m.settingIndex--
		}
	case "down", "j":
		if m.settingIndex < len(settings)-1 {
			m.settingIndex++
		}
	case "left", "h":
		switch m.settingIndex {
		case 0:
			if m.config.DefaultQuality > 10 {
				m.config.DefaultQuality -= 5
			}
		}
	case "right", "l":
		switch m.settingIndex {
		case 0:
			if m.config.DefaultQuality < 100 {
				m.config.DefaultQuality += 5
			}
		}
	case "enter", " ":
		switch m.settingIndex {
		case 2:
			m.loadFiles(m.config.OutputDir)
			m.state = stateSelectOutputDir
		case 3:
			m.config.ShowHiddenFiles = !m.config.ShowHiddenFiles
		case 4:
			m.config.ConfirmOverwrite = !m.config.ConfirmOverwrite
		case 5:
			m.config.Save()
			m.state = stateMenu
		}
	}
	return m, nil
}

type conversionDoneMsg struct {
	err error
}

type batchDoneMsg struct {
	results []batchResult
}

func (m Model) doConvert() tea.Cmd {
	return func() tea.Msg {
		opts := image.Options{
			Quality: m.quality,
		}
		err := image.Convert(m.inputFile, m.outputFile, opts)
		return conversionDoneMsg{err: err}
	}
}

func (m Model) doBatchConvert() tea.Cmd {
	return func() tea.Msg {
		results := []batchResult{}
		opts := image.Options{
			Quality: m.quality,
		}

		for _, inputPath := range m.selectedFiles {
			ext := filepath.Ext(inputPath)
			base := strings.TrimSuffix(filepath.Base(inputPath), ext)
			outputPath := filepath.Join(m.batchOutputDir, base+"."+m.outputFormat)

			err := image.Convert(inputPath, outputPath, opts)
			results = append(results, batchResult{
				input:   inputPath,
				output:  outputPath,
				success: err == nil,
				err:     err,
			})
		}

		return batchDoneMsg{results: results}
	}
}

func (m Model) updateBatchSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	maxVisible := m.height - 15
	if maxVisible < 5 {
		maxVisible = 5
	}

	switch msg.String() {
	case "up", "k":
		if m.fileIndex > 0 {
			m.fileIndex--
			if m.fileIndex < m.scrollOffset {
				m.scrollOffset = m.fileIndex
			}
		}
	case "down", "j":
		if m.fileIndex < len(m.files)-1 {
			m.fileIndex++
			if m.fileIndex >= m.scrollOffset+maxVisible {
				m.scrollOffset = m.fileIndex - maxVisible + 1
			}
		}
	case " ": // Space to toggle selection
		if m.fileIndex < len(m.files) {
			entry := &m.files[m.fileIndex]
			if entry.isImg {
				entry.selected = !entry.selected
				if entry.selected {
					m.selectedFiles = append(m.selectedFiles, entry.path)
				} else {
					// Remove from selected
					for i, p := range m.selectedFiles {
						if p == entry.path {
							m.selectedFiles = append(m.selectedFiles[:i], m.selectedFiles[i+1:]...)
							break
						}
					}
				}
			}
		}
	case "a": // Select all images
		m.selectedFiles = []string{}
		for i := range m.files {
			if m.files[i].isImg {
				m.files[i].selected = true
				m.selectedFiles = append(m.selectedFiles, m.files[i].path)
			}
		}
	case "n": // Deselect all
		m.selectedFiles = []string{}
		for i := range m.files {
			m.files[i].selected = false
		}
	case "enter":
		if m.fileIndex < len(m.files) {
			entry := m.files[m.fileIndex]
			if entry.isDir {
				m.loadFiles(entry.path)
			} else if len(m.selectedFiles) > 0 {
				m.config.LastInputDir = m.currentDir
				m.state = stateSelectFormat
			}
		}
	case "c": // Continue with selection
		if len(m.selectedFiles) > 0 {
			m.config.LastInputDir = m.currentDir
			m.state = stateSelectFormat
		}
	case "tab":
		m.config.ShowHiddenFiles = !m.config.ShowHiddenFiles
		m.loadFiles(m.currentDir)
	}
	return m, nil
}

func (m Model) updateBatchConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "enter":
		// Create output directory in ~/Downloads/photon with timestamp
		if err := m.config.EnsureOutputDir(); err == nil {
			timestamp := time.Now().Format("2006-01-02_15-04-05")
			m.batchOutputDir = filepath.Join(m.config.OutputDir, fmt.Sprintf("batch_%s_%s", m.outputFormat, timestamp))
		} else {
			// Fallback to current directory
			timestamp := time.Now().Format("2006-01-02_15-04-05")
			m.batchOutputDir = filepath.Join(m.currentDir, fmt.Sprintf("batch_%s_%s", m.outputFormat, timestamp))
		}
		os.MkdirAll(m.batchOutputDir, 0755)

		m.state = stateBatchConverting
		m.converting = true
		return m, m.doBatchConvert()
	case "n", "esc":
		m.state = stateMenu
	}
	return m, nil
}

func (m Model) View() string {
	var s strings.Builder

	// Header
	header := LogoStyle.Render(LogoSmall) + "  " + SubtitleStyle.Render("Image Format Converter")
	s.WriteString(header + "\n\n")

	switch m.state {
	case stateMenu:
		s.WriteString(m.viewMenu())
	case stateSelectInput:
		s.WriteString(m.viewFileBrowser("Select Input Image"))
	case stateSelectOutput:
		s.WriteString(m.viewFileBrowser("Select Output Location"))
	case stateSelectFormat:
		s.WriteString(m.viewFormatSelect())
	case stateQuality:
		s.WriteString(m.viewQuality())
	case stateConfirm:
		s.WriteString(m.viewConfirm())
	case stateConverting:
		s.WriteString(m.viewConverting())
	case stateComplete:
		s.WriteString(m.viewComplete())
	case stateSettings:
		s.WriteString(m.viewSettings())
	case stateSelectOutputDir:
		s.WriteString(m.viewOutputDirBrowser())
	case stateBatchSelect:
		s.WriteString(m.viewBatchSelect())
	case stateBatchConfirm:
		s.WriteString(m.viewBatchConfirm())
	case stateBatchConverting:
		s.WriteString(m.viewBatchConverting())
	case stateBatchComplete:
		s.WriteString(m.viewBatchComplete())
	}

	// Footer help
	s.WriteString("\n" + m.viewHelp())

	return s.String()
}

func (m Model) viewMenu() string {
	var s strings.Builder
	s.WriteString(TitleStyle.Render("Main Menu") + "\n\n")

	for i, item := range m.menuItems {
		cursor := "  "
		style := ItemStyle
		if i == m.menuIndex {
			cursor = SelectedItemStyle.Render("▸ ")
			style = SelectedItemStyle
		}
		s.WriteString(cursor + style.Render(item) + "\n")
	}

	return BoxStyle.Render(s.String())
}

func (m Model) viewFileBrowser(title string) string {
	var s strings.Builder
	s.WriteString(TitleStyle.Render(title) + "\n")
	s.WriteString(SubtitleStyle.Render(m.currentDir) + "\n\n")

	maxVisible := m.height - 15
	if maxVisible < 5 {
		maxVisible = 5
	}

	end := m.scrollOffset + maxVisible
	if end > len(m.files) {
		end = len(m.files)
	}

	for i := m.scrollOffset; i < end; i++ {
		entry := m.files[i]
		cursor := "  "
		var style lipgloss.Style

		if i == m.fileIndex {
			cursor = SelectedItemStyle.Render("▸ ")
		}

		icon := "  "
		if entry.isDir {
			icon = "📁 "
			style = DirStyle
		} else if entry.isImg {
			icon = "🖼  "
			style = ImageFileStyle
		} else {
			icon = "📄 "
			style = FileStyle
		}

		if i == m.fileIndex {
			style = SelectedItemStyle
		}

		s.WriteString(cursor + icon + style.Render(entry.name) + "\n")
	}

	if len(m.files) > maxVisible {
		s.WriteString(fmt.Sprintf("\n%s", SubtitleStyle.Render(fmt.Sprintf("(%d/%d)", m.fileIndex+1, len(m.files)))))
	}

	return BoxStyle.Render(s.String())
}

func (m Model) viewFormatSelect() string {
	var s strings.Builder
	s.WriteString(TitleStyle.Render("Select Output Format") + "\n\n")
	s.WriteString(SubtitleStyle.Render("Input: "+filepath.Base(m.inputFile)) + "\n\n")

	for i, format := range m.formats {
		style := FormatBadge
		if i == m.formatIndex {
			style = FormatBadgeSelected
		}
		s.WriteString(style.Render(strings.ToUpper(format)) + " ")
	}
	s.WriteString("\n\n")

	info := map[string]string{
		"png":  "Lossless, supports transparency",
		"jpg":  "Lossy, best for photos",
		"gif":  "Lossless, 256 colors, animation",
		"webp": "Modern, excellent compression",
		"bmp":  "Uncompressed, large files",
		"tiff": "Lossless, professional use",
		"avif": "Modern, best compression",
	}

	selected := m.formats[m.formatIndex]
	s.WriteString(SubtitleStyle.Render(info[selected]))

	return BoxStyle.Render(s.String())
}

func (m Model) viewQuality() string {
	var s strings.Builder
	s.WriteString(TitleStyle.Render("Set Quality") + "\n\n")

	// Quality bar
	barWidth := 40
	filled := (m.quality * barWidth) / 100
	empty := barWidth - filled

	filledBar := SliderFilled.Render(strings.Repeat("=", filled))
	emptyBar := SliderTrack.Render(strings.Repeat("-", empty))
	s.WriteString("[" + filledBar + emptyBar + "]")
	s.WriteString(fmt.Sprintf("  %d%%\n\n", m.quality))

	// Quality hint
	var hint string
	switch {
	case m.quality < 30:
		hint = "Very low quality, small file size"
	case m.quality < 60:
		hint = "Low quality, reduced file size"
	case m.quality < 80:
		hint = "Good balance of quality and size"
	case m.quality < 95:
		hint = "High quality, larger file size"
	default:
		hint = "Maximum quality, largest file size"
	}
	s.WriteString(SubtitleStyle.Render(hint))

	return BoxStyle.Render(s.String())
}

func (m Model) viewConfirm() string {
	var s strings.Builder
	s.WriteString(TitleStyle.Render("Confirm Conversion") + "\n\n")

	s.WriteString("🖼  Input:   " + ImageFileStyle.Render(filepath.Base(m.inputFile)) + "\n")
	s.WriteString("📄 Output:  " + ImageFileStyle.Render(filepath.Base(m.outputFile)) + "\n")
	s.WriteString("📁 Format:  " + FormatBadge.Render(strings.ToUpper(m.outputFormat)) + "\n")
	s.WriteString(fmt.Sprintf("⚙  Quality: %d%%\n\n", m.quality))

	s.WriteString(WarningStyle.Render("Proceed with conversion? (y/n)"))

	return BoxStyle.Render(s.String())
}

func (m Model) viewConverting() string {
	var s strings.Builder
	s.WriteString(TitleStyle.Render("Converting...") + "\n\n")
	s.WriteString(m.spinner.View() + " Processing " + filepath.Base(m.inputFile))
	return BoxStyle.Render(s.String())
}

func (m Model) viewComplete() string {
	var s strings.Builder

	if m.convErr != nil {
		s.WriteString(ErrorStyle.Render("✗ Conversion Failed") + "\n\n")
		s.WriteString(m.convErr.Error())
	} else {
		s.WriteString(SuccessStyle.Render("✓ Conversion Complete") + "\n\n")
		s.WriteString("📄 Output: " + ImageFileStyle.Render(m.outputFile) + "\n")

		if info, err := os.Stat(m.outputFile); err == nil {
			size := float64(info.Size()) / 1024
			s.WriteString(fmt.Sprintf("📊 Size:   %.1f KB", size))
		}
	}

	s.WriteString("\n\nPress any key to continue...")

	return BoxStyle.Render(s.String())
}

func (m Model) viewOutputDirBrowser() string {
	var s strings.Builder
	s.WriteString(TitleStyle.Render("Select Output Directory") + "\n")
	s.WriteString(SubtitleStyle.Render(m.currentDir) + "\n\n")

	maxVisible := m.height - 15
	if maxVisible < 5 {
		maxVisible = 5
	}

	end := m.scrollOffset + maxVisible
	if end > len(m.files) {
		end = len(m.files)
	}

	for i := m.scrollOffset; i < end; i++ {
		entry := m.files[i]
		if !entry.isDir {
			continue
		}
		cursor := "  "
		style := DirStyle

		if i == m.fileIndex {
			cursor = SelectedItemStyle.Render("▸ ")
			style = SelectedItemStyle
		}

		s.WriteString(cursor + "📁 " + style.Render(entry.name) + "\n")
	}

	s.WriteString("\n" + SubtitleStyle.Render("Press 's' to select this directory"))

	return BoxStyle.Render(s.String())
}

func (m Model) viewSettings() string {
	var s strings.Builder
	s.WriteString(TitleStyle.Render("Settings") + "\n\n")

	settings := []struct {
		name  string
		value string
	}{
		{"Default Quality", fmt.Sprintf("◀ %d%% ▶", m.config.DefaultQuality)},
		{"Default Format", m.config.DefaultFormat},
		{"Output Directory", m.config.OutputDir},
		{"Show Hidden Files", boolIcon(m.config.ShowHiddenFiles)},
		{"Confirm Overwrite", boolIcon(m.config.ConfirmOverwrite)},
		{"Back", ""},
	}

	for i, setting := range settings {
		cursor := "  "
		style := ItemStyle
		if i == m.settingIndex {
			cursor = SelectedItemStyle.Render("▸ ")
			style = SelectedItemStyle
		}

		line := setting.name
		if setting.value != "" {
			line = fmt.Sprintf("%-20s %s", setting.name, setting.value)
		}
		s.WriteString(cursor + style.Render(line) + "\n")
	}

	return BoxStyle.Render(s.String())
}

func boolIcon(b bool) string {
	if b {
		return SuccessStyle.Render("●")
	}
	return lipgloss.NewStyle().Foreground(muted).Render("○")
}

func (m Model) viewBatchSelect() string {
	var s strings.Builder
	s.WriteString(TitleStyle.Render("Batch Select Images") + "\n")
	s.WriteString(SubtitleStyle.Render(m.currentDir) + "\n")
	s.WriteString(WarningStyle.Render(fmt.Sprintf("Selected: %d images", len(m.selectedFiles))) + "\n\n")

	maxVisible := m.height - 17
	if maxVisible < 5 {
		maxVisible = 5
	}

	end := m.scrollOffset + maxVisible
	if end > len(m.files) {
		end = len(m.files)
	}

	for i := m.scrollOffset; i < end; i++ {
		entry := m.files[i]
		cursor := "  "
		var style lipgloss.Style

		if i == m.fileIndex {
			cursor = SelectedItemStyle.Render("▸ ")
		}

		checkbox := "[ ] "
		if entry.selected {
			checkbox = SuccessStyle.Render("[✓] ")
		}

		icon := "   "
		if entry.isDir {
			icon = "📁 "
			style = DirStyle
			checkbox = "    "
		} else if entry.isImg {
			icon = "🖼  "
			style = ImageFileStyle
		} else {
			icon = "📄 "
			style = FileStyle
			checkbox = "    "
		}

		if i == m.fileIndex {
			style = SelectedItemStyle
		}

		s.WriteString(cursor + checkbox + icon + style.Render(entry.name) + "\n")
	}

	if len(m.files) > maxVisible {
		s.WriteString(fmt.Sprintf("\n%s", SubtitleStyle.Render(fmt.Sprintf("(%d/%d)", m.fileIndex+1, len(m.files)))))
	}

	return BoxStyle.Render(s.String())
}

func (m Model) viewBatchConfirm() string {
	var s strings.Builder
	s.WriteString(TitleStyle.Render("Confirm Batch Conversion") + "\n\n")

	s.WriteString(fmt.Sprintf("🖼  Files:   %s\n", WarningStyle.Render(fmt.Sprintf("%d images", len(m.selectedFiles)))))
	s.WriteString(fmt.Sprintf("📄 Format:  %s\n", FormatBadge.Render(strings.ToUpper(m.outputFormat))))
	s.WriteString(fmt.Sprintf("⚙  Quality: %d%%\n", m.quality))
	s.WriteString(fmt.Sprintf("📁 Output:  %s\n\n", SubtitleStyle.Render(m.config.OutputDir)))

	s.WriteString(WarningStyle.Render("Proceed with batch conversion? (y/n)"))

	return BoxStyle.Render(s.String())
}

func (m Model) viewBatchConverting() string {
	var s strings.Builder
	s.WriteString(TitleStyle.Render("Converting...") + "\n\n")
	s.WriteString(m.spinner.View() + fmt.Sprintf(" Processing %d images...", len(m.selectedFiles)))
	return BoxStyle.Render(s.String())
}

func (m Model) viewBatchComplete() string {
	var s strings.Builder

	successCount := 0
	for _, r := range m.batchResults {
		if r.success {
			successCount++
		}
	}

	if successCount == len(m.batchResults) {
		s.WriteString(SuccessStyle.Render("✓ Batch Complete") + "\n\n")
	} else {
		s.WriteString(WarningStyle.Render("⚠ Batch Complete (with errors)") + "\n\n")
	}

	s.WriteString(fmt.Sprintf("🖼  Converted: %d/%d images\n", successCount, len(m.batchResults)))
	s.WriteString(fmt.Sprintf("📁 Output:    %s\n\n", SubtitleStyle.Render(m.batchOutputDir)))

	// Show errors if any
	for _, r := range m.batchResults {
		if !r.success {
			s.WriteString(ErrorStyle.Render("✗ ") + filepath.Base(r.input) + ": " + r.err.Error() + "\n")
		}
	}

	s.WriteString("\nPress any key to continue...")

	return BoxStyle.Render(s.String())
}

func (m Model) viewHelp() string {
	var help string
	switch m.state {
	case stateMenu:
		help = "↑/↓: navigate • enter: select • q: quit"
	case stateSelectInput, stateSelectOutput:
		help = "↑/↓: navigate • enter: select • tab: toggle hidden • esc: back"
	case stateSelectFormat:
		help = "←/→: select format • enter: confirm • esc: back"
	case stateQuality:
		help = "←/→: adjust quality • enter: confirm • esc: back"
	case stateSettings:
		help = "↑/↓: navigate • ←/→: adjust • enter: toggle • esc: back"
	case stateSelectOutputDir:
		help = "↑/↓: navigate • enter: open dir • s: select current • esc: back"
	case stateBatchSelect:
		help = "↑/↓: navigate • space: select • a: all • n: none • c: continue • esc: back"
	case stateBatchConfirm:
		help = "y: confirm • n: cancel"
	default:
		help = "esc: back to menu • q: quit"
	}
	return HelpStyle.Render(help)
}

func Run() error {
	p := tea.NewProgram(NewModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
