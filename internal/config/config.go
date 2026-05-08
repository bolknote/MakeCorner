package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
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
	Mask       string
	OutDir     string
	SaveExif   bool
	Recursive  bool
	KeepName   bool
	Moo        bool
}

var (
	rxBackground   = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
	rxLegacyBraces = regexp.MustCompile(`\{[^}]+\}`)
)

type flagInfo struct{ long, short, desc string }

func Load(binaryName string, args []string) (Config, error) {
	fs := flag.NewFlagSet(filepath.Base(binaryName), flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

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
		fs.IntVar(field, short, v, desc)
		infos = append(infos, flagInfo{long, short, desc})
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
		fs.BoolVar(field, short, v, desc)
		infos = append(infos, flagInfo{long, short, desc})
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
		fs.StringVar(field, short, v, desc)
		infos = append(infos, flagInfo{long, short, desc})
	}

	addInt(&cfg.Quality, "quality", "q", 85, "JPEG quality")
	addInt(&cfg.Width, "width", "w", 660, "Output width")
	addInt(&cfg.Radius, "radius", "r", 10, "Corner radius")
	addString(&bgStr, "background", "b", "#ffffff", "Corner background color")
	addString(&cfg.Mask, "mask", "m", "*", "Input file mask")
	addString(&cfg.OutDir, "out-dir", "o", "out", "Output directory")
	addBool(&cfg.SaveExif, "save-exif", "e", false, "Copy EXIF metadata for JPEG outputs")
	addBool(&cfg.Recursive, "recursive", "R", false, "Recursive file discovery")
	addBool(&cfg.KeepName, "keep-name", "k", false, "Keep source file names")
	addBool(&cfg.Moo, "moo", "M", false, "Show easter egg")

	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Flags:")
		for _, fi := range infos {
			if fi.long == "moo" {
				continue
			}
			fmt.Fprintf(os.Stderr, "-%s, --%s: %s\n", fi.short, fi.long, fi.desc)
		}
	}

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	bg, err := parseBackgroundColor(bgStr)
	if err != nil {
		return Config{}, err
	}
	cfg.Background = bg
	cfg.Mask = translateLegacyMask(cfg.Mask)

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

func MooASCII() string {
	return `
                (__)
                (oo)
           /-----\/
          / |   ||
        *  /\--/\
           ~~  ~~
`
}

func parseBackgroundColor(v string) ([3]uint8, error) {
	if !rxBackground.MatchString(v) {
		return [3]uint8{}, fmt.Errorf("invalid background color %q", v)
	}
	n, _ := strconv.ParseUint(v[1:], 16, 32)
	return [3]uint8{uint8(n >> 16), uint8(n >> 8), uint8(n)}, nil
}

func translateLegacyMask(m string) string {
	return rxLegacyBraces.ReplaceAllStringFunc(m, func(part string) string {
		inner := part[1 : len(part)-1]
		return "[" + strings.ReplaceAll(inner, ",", "") + "]"
	})
}
