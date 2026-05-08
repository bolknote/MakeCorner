package config

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type Config struct {
	Quality    int
	Width      int
	Radius     int
	Background [3]uint8
	Masks      []string
	OutDir     string
	SaveExif   bool
	Recursive  bool
	KeepName   bool
	Moo        bool
}

// Sentinel errors returned by Load. The caller is expected to use errors.Is
// to differentiate them from validation errors:
//   - ErrUsageRequested: the user asked for --help; flag has already printed
//     the usage to the configured writer. The CLI should exit successfully.
//   - ErrFlagParse: argv was malformed; flag has already printed both the
//     diagnostic and the usage. The CLI should exit non-zero without
//     printing the same error a second time.
var (
	ErrUsageRequested = errors.New("usage requested")
	ErrFlagParse      = errors.New("flag parse failed")
)

var rxBackground = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

type flagInfo struct {
	long, short, desc string
	def               string
}

// Load parses arguments using a flag set whose diagnostics go to stderr. The
// binaryName is used to derive both the program name in usage and the
// optional INI file (binaryName.ini, falling back to makecorner.ini).
func Load(binaryName string, args []string, stderr io.Writer) (Config, error) {
	if stderr == nil {
		stderr = io.Discard
	}
	fs := flag.NewFlagSet(filepath.Base(binaryName), flag.ContinueOnError)
	fs.SetOutput(stderr)

	ini := readIniOptions(binaryName)
	var cfg Config
	var bgStr string
	var infos []flagInfo

	addInt := func(field *int, long, short string, def int, desc string) {
		v := def
		for _, k := range []string{long, short} {
			if s, ok := ini[k]; ok {
				if n, err := strconv.Atoi(s); err == nil {
					v = n
					break
				}
			}
		}
		fs.IntVar(field, long, v, desc)
		fs.IntVar(field, short, v, "alias for --"+long)
		infos = append(infos, flagInfo{long, short, desc, strconv.Itoa(def)})
	}
	addBool := func(field *bool, long, short string, def bool, desc string) {
		v := def
		for _, k := range []string{long, short} {
			if s, ok := ini[k]; ok {
				v = s == "1" || strings.EqualFold(s, "true")
				break
			}
		}
		fs.BoolVar(field, long, v, desc)
		fs.BoolVar(field, short, v, "alias for --"+long)
		infos = append(infos, flagInfo{long, short, desc, strconv.FormatBool(def)})
	}
	addString := func(field *string, long, short, def, desc string) {
		v := def
		for _, k := range []string{long, short} {
			if s, ok := ini[k]; ok {
				v = s
				break
			}
		}
		fs.StringVar(field, long, v, desc)
		fs.StringVar(field, short, v, "alias for --"+long)
		infos = append(infos, flagInfo{long, short, desc, def})
	}

	var maskStr string
	addInt(&cfg.Quality, "quality", "q", 85, "JPEG/WebP/HEIF/AVIF quality (1..100)")
	addInt(&cfg.Width, "width", "w", 660, "Output width; 0 keeps source width")
	addInt(&cfg.Radius, "radius", "r", 10, "Corner radius in pixels")
	addString(&bgStr, "background", "b", "#ffffff", "Corner background color (#RRGGBB)")
	addString(&maskStr, "mask", "m", "*", "Input file mask (glob)")
	addString(&cfg.OutDir, "out-dir", "o", "out", "Output directory")
	addBool(&cfg.SaveExif, "save-exif", "e", false, "Copy EXIF metadata for JPEG outputs")
	addBool(&cfg.Recursive, "recursive", "R", false, "Recursive file discovery")
	addBool(&cfg.KeepName, "keep-name", "k", false, "Keep source file names")
	addBool(&cfg.Moo, "moo", "M", false, "Show easter egg")

	fs.Usage = func() { writeUsage(stderr, fs.Name(), infos) }

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return Config{}, ErrUsageRequested
		}
		return Config{}, ErrFlagParse
	}

	bg, err := parseBackgroundColor(bgStr)
	if err != nil {
		return Config{}, err
	}
	cfg.Background = bg
	cfg.Masks = expandMask(maskStr)

	if cfg.Quality < 1 || cfg.Quality > 100 {
		return Config{}, errors.New("quality must be between 1 and 100")
	}
	if cfg.Radius < 0 {
		return Config{}, errors.New("radius must be >= 0")
	}
	if cfg.Width < 0 {
		return Config{}, errors.New("width must be >= 0")
	}

	return cfg, nil
}

func writeUsage(w io.Writer, name string, infos []flagInfo) {
	_, _ = fmt.Fprintf(w, "Usage: %s [flags]\n\nFlags:\n", name)
	for _, fi := range infos {
		if fi.long == "moo" {
			continue
		}
		_, _ = fmt.Fprintf(w, "  -%s, --%-12s %s (default: %s)\n", fi.short, fi.long, fi.desc, fi.def)
	}
}

func parseBackgroundColor(v string) ([3]uint8, error) {
	if !rxBackground.MatchString(v) {
		return [3]uint8{}, fmt.Errorf("invalid background color %q", v)
	}
	n, err := strconv.ParseUint(v[1:], 16, 32)
	if err != nil {
		// Should be unreachable given rxBackground, but guard anyway so we
		// never return a silently-zeroed color on future regex changes.
		return [3]uint8{}, fmt.Errorf("invalid background color %q: %w", v, err)
	}
	return [3]uint8{uint8(n >> 16), uint8(n >> 8), uint8(n)}, nil
}

// expandMask expands a single mask pattern into one or more filepath.Match
// patterns. Brace groups are expanded left-to-right (no nesting):
//   - Single-char alternatives {j,J} become a character class [jJ], which
//     filepath.Match understands natively.
//   - Multi-char alternatives {jpg,png} expand into separate patterns
//     ["*.jpg", "*.png"], because filepath.Match has no alternation syntax.
func expandMask(m string) []string {
	open := strings.IndexByte(m, '{')
	if open == -1 {
		return []string{m}
	}
	closeOff := strings.IndexByte(m[open:], '}')
	if closeOff == -1 {
		return []string{m}
	}
	close := open + closeOff

	prefix := m[:open]
	suffix := m[close+1:]
	alternatives := strings.Split(m[open+1:close], ",")

	allSingle := true
	for _, a := range alternatives {
		if len(a) != 1 {
			allSingle = false
			break
		}
	}

	if allSingle {
		// {j,J} → [jJ]: a character class that filepath.Match supports.
		return expandMask(prefix + "[" + strings.Join(alternatives, "") + "]" + suffix)
	}

	// Multi-char alternatives: expand into one pattern per alternative.
	var result []string
	for _, a := range alternatives {
		result = append(result, expandMask(prefix+a+suffix)...)
	}
	return result
}
