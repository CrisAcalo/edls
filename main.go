package edls

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"

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
		isHidden := isHidden(dir.Name())

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

		f, err := getFile(dir, isHidden)
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

	printList(fs, *flagNumberRecords)
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

func printList(fs []file, n int) {
	for _, file := range fs[:n] {

		style := mapStyleByFileType[file.fileType]

		fmt.Printf("%s %s %s %10d %s %s %s %s\n", file.mode, file.userName, file.groupName, file.size, file.modificationTime.Format(time.DateTime), style.icon, file.name, style.symbol)
	}
}

func getFile(dir fs.DirEntry, isHidden bool) (file, error) {
	info, err := dir.Info()

	if err != nil {
		return file{}, fmt.Errorf("dir.Info(): %v", err)
	}

	f := file{
		name:             dir.Name(),
		isDir:            dir.IsDir(),
		isHidden:         isHidden,
		userName:         "Cris",
		groupName:        "General",
		size:             info.Size(),
		modificationTime: info.ModTime(),
		mode:             info.Mode().String(),
	}
	setFile(&f)

	return f, nil
}

func setFile(f *file) {
	switch {
	case isLink(*f):
		f.fileType = fileLink
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

func isHidden(fileName string) bool {
	return strings.HasPrefix(fileName, ".")
}
