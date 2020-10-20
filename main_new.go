package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"syscall"

	"github.com/gin-gonic/gin"
)

type pageData struct {
	CurTab    string
	FreeSpace string
	Tabs      []string
	Folders   []string
	Files     []string
	Renamed   bool
	OldName   string
}

var pd pageData

const (
	WD = "C:/Personal/Test/" // working directory
	//	WD = "/home/juggerrus/torrents_sorted/"
	//	TORRENTSDIR = "/home/juggerrus/torrents/"
	BASETEMPL = "./templates/layout.html"
	DEBUG     = true
)

type DiskStatus struct { //
	All  uint64 `json:"all"`
	Used uint64 `json:"used"`
	Free uint64 `json:"free"`
}

// disk usage of path/disk
func DiskUsage(path string) (disk DiskStatus) {
	fs := syscall.Statfs_t{}
	err := syscall.Statfs(path, &fs)
	if err != nil {
		return
	}
	disk.All = fs.Blocks * uint64(fs.Bsize)
	disk.Free = fs.Bfree * uint64(fs.Bsize)
	disk.Used = disk.All - disk.Free
	return
}

const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
)

func getDirectoryContents(dir string) (output []string) {

	Files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal("Error reading directory contents: " + err.Error())
	}

	for _, file := range Files {
		output = append(output, file.Name())
	}

	return
}

func getFileExtension(name string) (s string) {
	pos := strings.LastIndex(name, ".")
	if pos != -1 {
		s = name[pos:]
	}
	return
}

func IsDirectory(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		log.Fatal("Не могу проверить папка или нет: " + err.Error())
	}
	return fileInfo.IsDir()
}

func indexHandler(c *gin.Context) {

	c.Redirect(http.StatusFound, "/folder/"+pd.Tabs[0])

}

func folderHandler(c *gin.Context) {

	pd.Files = make([]string, 0)
	pd.Folders = make([]string, 0)
	pd.CurTab = c.Param("folder")
	path := WD + "/" + pd.CurTab + "/"
	contents := getDirectoryContents(path)

	for _, item := range contents {
		if IsDirectory(path + item) {
			pd.Folders = append(pd.Folders, item)
		} else if getFileExtension(item) != "xml" {
			pd.Files = append(pd.Files, item)
		}
	}

	sort.Strings(pd.Files)
	sort.Strings(pd.Folders)

	disk := DiskUsage(WD)

	pd.FreeSpace = fmt.Sprintf("All: %.2f GB\n", float64(disk.Free)/float64(GB))

	c.HTML(http.StatusOK, "layout", pd)

	pd.Renamed = false
}

func renameHandler(c *gin.Context) {

	old_name := c.PostForm("old_name")
	new_name := c.PostForm("new_name")
	if len(new_name) == 0 {
		new_name = old_name
	} else {
		if !IsDirectory(WD + "/" + pd.CurTab + "/" + old_name) {
			new_name += getFileExtension(old_name)
		}
	}
	new_folder := c.PostForm("new_folder")

	if DEBUG {
		log.Println("Получен POST запрос")
		log.Println("Old_Name: " + old_name)
		log.Println("New_Name: " + new_name)
		log.Println("Folder: " + new_folder)
	}
	os.Rename(WD+"/"+pd.CurTab+"/"+old_name, WD+"/"+new_folder+"/"+new_name)
	pd.Renamed = true
	pd.OldName = old_name

	folderHandler(c)
}

func main() {
	log.Println("Starting...")

	pd.Tabs = getDirectoryContents(WD)
	pd.Renamed = false

	if !DEBUG {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	router.LoadHTMLGlob("templates/*")
	router.SetHTMLTemplate(template.Must(template.ParseFiles(BASETEMPL, "templates/index.html")))
	router.Static("/assets", "./assets")
	router.StaticFile("/favicon.ico", "./assets/favicon.ico")
	router.GET("/", indexHandler)
	router.GET("/folder/:folder", folderHandler)
	router.POST("/folder/:folder", renameHandler)
	router.Run(":3000")

}
