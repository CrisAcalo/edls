package main

import (
	"strings"
	"time"

	"github.com/fatih/color"
)

const Windows = "windows"

// file types

const (
	fileRegular int = iota
	fileDirectory
	fileExecutable
	fileCompress
	fileImage
	fileLink
	fileHidden
)

// file extensions
const (
	exe = ".exe"
	deb = ".deb"
	zip = ".zip"
	gz  = ".gz"
	tar = ".tar"
	rar = ".rar"
	png = ".png"
	jpg = ".jpg"
	gif = ".gif"
)

type file struct {
	name             string
	fileType         int
	isDir            bool
	isHidden         bool
	userName         string
	groupName        string
	size             int64
	modificationTime time.Time
	mode             string
	symlinkTarget    string
	contentType      string
}

type styleFileType struct {
	icon   string
	color  color.Attribute
	symbol string
}

var mapStyleByFileType = map[int]styleFileType{
	fileRegular:    {icon: "📄", color: color.FgWhite, symbol: ""},
	fileDirectory:  {icon: "📁", color: color.FgBlue, symbol: "/"},
	fileExecutable: {icon: "⚙️", color: color.FgGreen, symbol: "*"},
	fileCompress:   {icon: "📦", color: color.FgYellow, symbol: ""},
	fileImage:      {icon: "🖼️", color: color.FgMagenta, symbol: ""},
	fileLink:       {icon: "🔗", color: color.FgCyan, symbol: ""},
	fileHidden:     {icon: "🙈", color: color.FgHiBlack, symbol: ""},
}

var (
	blue    = color.New(color.FgBlue).Add(color.Bold).SprintFunc()
	green   = color.New(color.FgGreen).Add(color.Bold).SprintFunc()
	red     = color.New(color.FgRed).Add(color.Bold).SprintFunc()
	magneta = color.New(color.FgMagenta).Add(color.Bold).SprintFunc()
	cyan    = color.New(color.FgCyan).Add(color.Bold).SprintFunc()
	yellow  = color.New(color.FgYellow).Add(color.Bold).SprintFunc()
	black   = color.New(color.FgHiBlack).Add(color.Bold).SprintFunc()
	gray    = color.New(color.FgHiBlack).SprintFunc()
)

const (
	_KB = 1024
	_MB = 1024 * _KB
	_GB = 1024 * _MB
)

func colorizeMode(mode string) string {
	var sb strings.Builder
	for _, c := range mode {
		str := string(c)
		switch str {
		case "d":
			sb.WriteString(blue(str))
		case "r":
			sb.WriteString(yellow(str))
		case "w":
			sb.WriteString(red(str))
		case "x":
			sb.WriteString(green(str))
		case "-":
			sb.WriteString(gray(str))
		default:
			sb.WriteString(str)
		}
	}
	return sb.String()
}

func colorizeSize(size int64, sizeStr string) string {
	if size < _KB {
		return cyan(sizeStr)
	} else if size < 10*_MB {
		return green(sizeStr)
	} else if size < 500*_MB {
		return yellow(sizeStr)
	}
	return red(sizeStr)
}

func colorizeDate(t time.Time) string {
	timeStr := t.Format(time.DateTime)
	diff := time.Since(t)

	if diff < 24*time.Hour {
		return green(timeStr)
	} else if diff < 30*24*time.Hour {
		return cyan(timeStr)
	} else if diff < 365*24*time.Hour {
		return yellow(timeStr)
	}
	return gray(timeStr)
}
