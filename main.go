package main

import (
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"golang.org/x/exp/constraints"
)

func main() {
	flagPattern := flag.String("p", "", "filter by pattern")
	flagAll := flag.Bool("a", false, "files including hide files")
	flagNumberRecords := flag.Int("n", 0, "number of records to show")

	hasOrderByTime := flag.Bool("t", false, "sort by time, oldest first")
	hasOrderBySize := flag.Bool("s", false, "sort by size, smallest first")
	hasOrderReverse := flag.Bool("r", false, "reverse order while sorting")

	flag.Parse()

	path := flag.Arg(0)

	if path == "" {
		path = "."
	}

	dirs, err := os.ReadDir(path)

	if err != nil {
		panic(err)
	}

	fs := []file{}

	for _, dir := range dirs {
		isHidden := isHidden(dir.Name(), path)

		if isHidden && !*flagAll {
			continue // hace que no se muestren los archivos ocultos si no se pasa el flag -a
		}

		if *flagPattern != "" {
			isMatched, err := regexp.MatchString("(?i)"+*flagPattern, dir.Name())
			if err != nil {
				panic(err)
			}

			if !isMatched {
				continue
			}
		}

		f, err := getFile(dir, isHidden, path)
		if err != nil {
			panic(err)
		}

		fs = append(fs, f)
	}

	if !*hasOrderBySize || !*hasOrderByTime {
		orderByName(fs, *hasOrderReverse)
	}

	if *flagNumberRecords == 0 || *flagNumberRecords > len(fs) {
		*flagNumberRecords = len(fs)
	}

	if *hasOrderBySize && !*hasOrderByTime {
		orderBySize(fs, *hasOrderReverse)
	}

	if *hasOrderByTime {
		orderByTime(fs, *hasOrderReverse)
	}

	printList(fs, *flagNumberRecords, path)
}

func mySort[T constraints.Ordered](i, j T, isReverse bool) bool {
	if isReverse {
		return i > j
	}
	return i < j
}

func orderByName(files []file, isReverse bool) {
	sort.SliceStable(files, func(i, j int) bool {
		return mySort(
			strings.ToLower(files[i].name),
			strings.ToLower(files[j].name),
			isReverse)
	})
}

func orderBySize(files []file, isReverse bool) {
	sort.SliceStable(files, func(i, j int) bool {
		return mySort(
			files[i].size,
			files[j].size,
			isReverse)
	})
}

func orderByTime(files []file, isReverse bool) {
	sort.SliceStable(files, func(i, j int) bool {
		return mySort(
			files[i].modificationTime.Unix(),
			files[j].modificationTime.Unix(),
			isReverse)
	})
}

func printList(fs []file, n int, basePath string) {
	absPath, _ := filepath.Abs(basePath)
	fmt.Printf("\n📂 %s %s\n\n", gray("Directorio:"), blue(absPath))

	var maxUser, maxGroup, maxSize, maxName int
	var totalSize int64
	var filesCount, dirsCount, hiddenCount int

	// First pass to calculate maximum widths and stats
	for _, f := range fs[:n] {
		if f.isHidden {
			hiddenCount++
		}
		if f.isDir {
			dirsCount++
		} else {
			filesCount++
			totalSize += f.size
		}

		if len(f.userName) > maxUser {
			maxUser = len(f.userName)
		}
		if len(f.groupName) > maxGroup {
			maxGroup = len(f.groupName)
		}
		sizeStr := humanize.Bytes(uint64(f.size))
		if len(sizeStr) > maxSize {
			maxSize = len(sizeStr)
		}

		nameLen := len(f.name)
		if f.symlinkTarget != "" {
			nameLen += 4 + len(f.symlinkTarget) // " -> target"
		}
		if nameLen > maxName {
			maxName = nameLen
		}
	}

	// Second pass to print perfectly aligned
	for _, f := range fs[:n] {
		style := mapStyleByFileType[f.fileType]
		sizeStr := humanize.Bytes(uint64(f.size))

		paddedUser := f.userName
		if maxUser > 0 {
			paddedUser += strings.Repeat(" ", maxUser-len(f.userName))
		}
		paddedGroup := f.groupName
		if maxGroup > 0 {
			paddedGroup += strings.Repeat(" ", maxGroup-len(f.groupName))
		}
		paddedSize := strings.Repeat(" ", maxSize-len(sizeStr)) + sizeStr

		userStr := ""
		if maxUser > 0 {
			userStr = gray(paddedUser) + "  "
		}
		groupStr := ""
		if maxGroup > 0 {
			groupStr = gray(paddedGroup) + "  "
		}

		nameOutput := setColor(f.name, style.color)

		// calculate physical length for padding
		nameLen := len(f.name)
		if f.symlinkTarget != "" {
			nameOutput += gray(" -> ") + cyan(f.symlinkTarget)
			nameLen += 4 + len(f.symlinkTarget)
		}

		padSpaces := ""
		if maxName > nameLen {
			padSpaces = strings.Repeat(" ", maxName-nameLen)
		}

		typeStr := ""
		if f.contentType != "" {
			typeStr = padSpaces + "  " + gray(f.contentType)
		} else {
			typeStr = padSpaces
		}

		fmt.Printf("%s  %s%s%s  %s  %s %s%s %s %s\n",
			colorizeMode(f.mode),
			userStr,
			groupStr,
			colorizeSize(f.size, paddedSize),
			colorizeDate(f.modificationTime),
			style.icon,
			nameOutput,
			style.symbol,
			markHidden(f.isHidden),
			typeStr)
	}

	fmt.Printf("\n📊 %s %s %s %s %s | %s %s | %s %s\n\n",
		gray("Total:"), cyan(fmt.Sprintf("%d", filesCount)), gray("archivos y"), cyan(fmt.Sprintf("%d", dirsCount)), gray("carpetas"),
		gray("Peso total:"), colorizeSize(totalSize, humanize.Bytes(uint64(totalSize))),
		gray("Ocultos:"), yellow(fmt.Sprintf("%d", hiddenCount)),
	)
}

func getContentType(fullPath string, isDir bool, isLink bool) string {
	if isDir || isLink {
		return ""
	}
	f, err := os.Open(fullPath)
	if err != nil {
		return ""
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil || n == 0 {
		return ""
	}
	contentType := http.DetectContentType(buf[:n])
	if idx := strings.Index(contentType, ";"); idx != -1 {
		contentType = contentType[:idx]
	}
	return contentType
}

func getFile(dir fs.DirEntry, isHidden bool, basePath string) (file, error) {
	info, err := dir.Info()

	if err != nil {
		return file{}, fmt.Errorf("dir.Info(): %v", err)
	}

	fullPath := filepath.Join(basePath, dir.Name())
	userName, groupName := getUserAndGroup(info.Sys(), fullPath)

	isLink := info.Mode()&os.ModeSymlink != 0
	symlinkTarget := ""
	if isLink {
		symlinkTarget, _ = os.Readlink(fullPath)
	}

	contentType := getContentType(fullPath, info.IsDir(), isLink)

	f := file{
		name:             dir.Name(),
		isDir:            dir.IsDir(),
		isHidden:         isHidden,
		userName:         userName,
		groupName:        groupName,
		size:             info.Size(),
		modificationTime: info.ModTime(),
		mode:             info.Mode().String(),
		symlinkTarget:    symlinkTarget,
		contentType:      contentType,
	}
	setFile(&f)

	return f, nil
}

func setFile(f *file) {
	switch {
	case isLink(*f):
		f.fileType = fileLink
	case f.isHidden:
		f.fileType = fileHidden
	case f.isDir:
		f.fileType = fileDirectory
	case isExec(*f):
		f.fileType = fileExecutable
	case isCompress(*f):
		f.fileType = fileCompress
	case isImage(*f):
		f.fileType = fileImage
	default:
		f.fileType = fileRegular
	}
}

func setColor(nameFile string, styleColor color.Attribute) string {
	switch styleColor {
	case color.FgBlue:
		return blue(nameFile)
	case color.FgGreen:
		return green(nameFile)
	case color.FgRed:
		return red(nameFile)
	case color.FgMagenta:
		return magneta(nameFile)
	case color.FgCyan:
		return cyan(nameFile)
	case color.FgYellow:
		return yellow(nameFile)
	case color.FgHiBlack:
		return black(nameFile)
	default:
		return nameFile
	}
}

func isLink(f file) bool {
	return strings.HasPrefix(strings.ToUpper(f.mode), "L")
}

func isExec(f file) bool {
	if runtime.GOOS == Windows {
		return strings.HasSuffix(f.name, exe)
	}

	return strings.Contains(f.mode, "x")
}

func isCompress(f file) bool {
	return strings.HasSuffix(f.name, zip) ||
		strings.HasSuffix(f.name, gz) ||
		strings.HasSuffix(f.name, tar) ||
		strings.HasSuffix(f.name, rar) ||
		strings.HasSuffix(f.name, deb)
}

func isImage(f file) bool {
	return strings.HasSuffix(f.name, png) ||
		strings.HasSuffix(f.name, jpg) ||
		strings.HasSuffix(f.name, gif)
}

func isHidden(fileName string, basePath string) bool {
	if strings.HasPrefix(fileName, ".") {
		return true
	}

	filePath := fileName

	if runtime.GOOS == Windows {
		filePath = filepath.Join(basePath, fileName)
	}

	return isHiddenFile(filePath)
}

func markHidden(isHidden bool) string {
	if !isHidden {
		return ""
	}
	return yellow("ø")
}
