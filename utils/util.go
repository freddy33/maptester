package utils

import (
	"bufio"
	"github.com/google/logger"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

var Verbose = false
var Log = logger.Init("util", true, false, ioutil.Discard)

func FileExists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	if err != nil {
		Log.Errorf("testing if file %s exists unknown error %v", path, err)
		return false
	}
	return true
}

func DeleteFile(path string) bool {
	if FileExists(path) {
		ExitOnError(os.Remove(path))
		return true
	}
	return false
}

func DirExists(dir string, subPath string) (bool, string) {
	p := filepath.Join(dir, subPath)
	fi, err := os.Stat(p)
	if os.IsNotExist(err) {
		return false, p
	}
	if err != nil {
		Log.Errorf("searching for %s folder in %s returned unknown error %v", subPath, dir, err)
		return false, p
	}
	return fi != nil && fi.IsDir(), p
}

func AbsPath(dir string) string {
	absPath, err := filepath.Abs(dir)
	if err != nil {
		Log.Fatalf("could not extract absolute path returned unknown error %v", err)
		return ""
	}
	return absPath
}

func GetGitRootDir() string {
	absPath := AbsPath(".")
	p := absPath
	// Check first if we are below the checkout dir
	if b, p := DirExists(p, "qsm-go"); b {
		if b, _ = DirExists(p, ".git"); b {
			return p
		} else {
			Log.Fatalf("found qsm-go sub folder at %s which not a git checkout", p)
			return ""
		}
	}
	for {
		if p == "." || p == "/" {
			Log.Fatalf("did not find path with git under %s", absPath)
			return ""
		}
		if b, _ := DirExists(p, ".git"); b {
			return p
		}
		p = filepath.Dir(p)
	}
}

func getOrCreateBuildSubDir(subPath string) string {
	buildDir := GetBuildDir()
	b, p := DirExists(buildDir, subPath)
	if !b {
		err := os.MkdirAll(p, 0755)
		if err != nil {
			Log.Fatalf("could not create sub build dir %s due to error %v", p, err)
			return ""
		}
	}
	return p
}

func GetBuildDir() string {
	gitRootDir := GetGitRootDir()
	b, p := DirExists(gitRootDir, "build")
	if !b {
		err := os.MkdirAll(p, 0755)
		if err != nil {
			Log.Fatalf("could not create build dir %s due to error %v", p, err)
			return ""
		}
	}
	return p
}

func GetConfDir() string {
	b, p := DirExists(GetGitRootDir(), "conf")
	if !b {
		Log.Fatalf("conf dir %s does not exists!", p)
		return ""
	}
	return p
}

func CreateFile(dir, fileName string) *os.File {
	p := filepath.Join(dir, fileName)
	f, err := os.Create(p)
	if err != nil {
		Log.Fatalf("could not create file %s due to %v", p, err)
		return nil
	}
	return f
}

func GetGenDataDir() string {
	return getOrCreateBuildSubDir("gendata")
}

func GetOutPerfDir() string {
	return getOrCreateBuildSubDir("perf")
}

func ExitOnError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

var lengthByte = make([]byte, 1)

func WriteDataBlock(file *os.File, bytes []byte) int {
	n, err := file.Write(bytes)
	ExitOnError(err)
	return n
}

func WriteDataBlockPrefixSize(file *os.File, bytes []byte) byte {
	l := len(bytes)
	if l > 255 {
		logger.Errorf("Cannot write block bigger than 255. It is %d", l)
		return 0
	}
	lengthByte[0] = byte(l)
	_, err := file.Write(lengthByte)
	ExitOnError(err)
	_, err = file.Write(bytes)
	ExitOnError(err)
	return lengthByte[0]
}

func ReadDataBlockPrefixSize(r *bufio.Reader) []byte {
	length, err := r.ReadByte()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		ExitOnError(err)
	}
	result := make([]byte, length)
	_, err = io.ReadFull(r, result)
	if err != nil {
		ExitOnError(err)
	}
	return result
}

func CloseFile(file *os.File) {
	ExitOnError(file.Close())
}

func WriteNextString(file *os.File, text string) {
	_, err := file.WriteString(text)
	ExitOnError(err)
}

const DefaultEpsilon = 1e-4

func Float32Equal(a, b float32) bool {
	return Float32EqualEpsilon(a, b, DefaultEpsilon)
}

func Float32EqualEpsilon(a, b float32, epsilon float32) bool {
	delta := a - b
	return (delta >= 0.0 && delta < epsilon) || (delta < 0.0 && delta > -epsilon)
}
