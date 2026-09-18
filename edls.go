package edls

import "time"

const Windows = "windows"

// file types

const (
	fileRegular int = iota
	fileDirectory
	fileExecutable
	fileCompress
	fileImage
	fileLink
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
}

type styleFileType struct {
	icon   string
	color  string
	symbol string
}

var mapStyleByFileType = map[int]styleFileType{
	fileRegular:    {icon: "📄", color: "white", symbol: ""},
	fileDirectory:  {icon: "📁", color: "blue", symbol: "/"},
	fileExecutable: {icon: "⚙️", color: "green", symbol: "*"},
	fileCompress:   {icon: "📦", color: "yellow", symbol: ""},
	fileImage:      {icon: "🖼️", color: "magenta", symbol: ""},
	fileLink:       {icon: "🔗", color: "cyan", symbol: ""},
}
