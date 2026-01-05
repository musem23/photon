package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	primary   = lipgloss.Color("#7C3AED")
	secondary = lipgloss.Color("#A78BFA")
	accent    = lipgloss.Color("#F59E0B")
	success   = lipgloss.Color("#10B981")
	danger    = lipgloss.Color("#EF4444")
	muted     = lipgloss.Color("#6B7280")
	bg        = lipgloss.Color("#1F2937")
	bgLight   = lipgloss.Color("#374151")
	fg        = lipgloss.Color("#F9FAFB")
	fgMuted   = lipgloss.Color("#9CA3AF")

	// Logo style
	LogoStyle = lipgloss.NewStyle().
			Foreground(primary).
			Bold(true).
			MarginBottom(1)

	// Title styles
	TitleStyle = lipgloss.NewStyle().
			Foreground(fg).
			Bold(true).
			Padding(0, 1).
			MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(fgMuted).
			Italic(true)

	// Box styles
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primary).
			Padding(1, 2)

	ActiveBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accent).
			Padding(1, 2)

	// List item styles
	ItemStyle = lipgloss.NewStyle().
			Foreground(fg).
			PaddingLeft(2)

	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(accent).
				Bold(true).
				PaddingLeft(1)

	// Format badge styles
	FormatBadge = lipgloss.NewStyle().
			Foreground(fg).
			Background(primary).
			Padding(0, 1).
			MarginRight(1)

	FormatBadgeSelected = lipgloss.NewStyle().
				Foreground(bg).
				Background(accent).
				Bold(true).
				Padding(0, 1).
				MarginRight(1)

	// Status styles
	SuccessStyle = lipgloss.NewStyle().
			Foreground(success).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(danger).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(accent)

	// Progress bar
	ProgressStyle = lipgloss.NewStyle().
			Foreground(primary)

	ProgressCompleteStyle = lipgloss.NewStyle().
				Foreground(success)

	// Help text
	HelpStyle = lipgloss.NewStyle().
			Foreground(muted).
			MarginTop(1)

	// Quality slider
	SliderTrack = lipgloss.NewStyle().
			Foreground(bgLight)

	SliderFilled = lipgloss.NewStyle().
			Foreground(primary)

	SliderThumb = lipgloss.NewStyle().
			Foreground(accent).
			Bold(true)

	// File item styles
	DirStyle = lipgloss.NewStyle().
			Foreground(secondary).
			Bold(true)

	FileStyle = lipgloss.NewStyle().
			Foreground(fg)

	ImageFileStyle = lipgloss.NewStyle().
			Foreground(success)

	// Tab styles
	TabStyle = lipgloss.NewStyle().
			Foreground(fgMuted).
			Padding(0, 2)

	ActiveTabStyle = lipgloss.NewStyle().
			Foreground(accent).
			Bold(true).
			Padding(0, 2).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(accent)

	// Button styles
	ButtonStyle = lipgloss.NewStyle().
			Foreground(fg).
			Background(bgLight).
			Padding(0, 3).
			MarginRight(1)

	ButtonActiveStyle = lipgloss.NewStyle().
				Foreground(bg).
				Background(accent).
				Bold(true).
				Padding(0, 3).
				MarginRight(1)
)

const Logo = `
 ██████╗ ██╗  ██╗ ██████╗ ████████╗ ██████╗ ███╗   ██╗
 ██╔══██╗██║  ██║██╔═══██╗╚══██╔══╝██╔═══██╗████╗  ██║
 ██████╔╝███████║██║   ██║   ██║   ██║   ██║██╔██╗ ██║
 ██╔═══╝ ██╔══██║██║   ██║   ██║   ██║   ██║██║╚██╗██║
 ██║     ██║  ██║╚██████╔╝   ██║   ╚██████╔╝██║ ╚████║
 ╚═╝     ╚═╝  ╚═╝ ╚═════╝    ╚═╝    ╚═════╝ ╚═╝  ╚═══╝`

const LogoSmall = " ⚛ PHOTON"
