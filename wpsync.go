// wpsync - command-line tool to sync wordpress
// https://github.com/mkaz/wpsync
//

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

// Config is the structure of the jwt-auth response and
// settings, it is used to unmarshal the data
type Config struct {
	SiteURL string `json:"site-url"`
	Token   string `json:"token"`
}

type Post struct {
	Id        int    `json:"id"`
	Title     string `json:"-"`
	Date      string `json:"-"`
	URL       string `json:"link"`
	Content   string `json:"-"`
	Category  string `json:"-"`
	Status    string `json:"status"`
	Tags      string `json:"-"`
	LocalFile string
	ModDate   time.Time `json:"-"`
	SyncDate  time.Time
}

type Page struct {
	Id        int    `json:"id"`
	Title     string `json:"-"`
	URL       string `json:"link"`
	Content   string `json:"-"`
	Status    string `json:"status"`
	ParentId  int    `json:"-"`
	Template  string `json:"-"`
	Order     string `json:"-"`
	LocalFile string
	ModDate   time.Time `json:"-"`
	SyncDate  time.Time
}

type Media struct {
	Id        int    `json:"id"`
	URL       string `json:"source_url"`
	Link      string `json:"link"`
	LocalFile string
}

var conf Config
var log Logger
var setup bool
var dryrun bool
var confirm bool

// read config and parse args
func myInit() {

	var helpFlag = flag.Bool("help", false, "Display help and quit")
	var versionFlag = flag.Bool("version", false, "Display version and quit")
	var testFlag = flag.Bool("test", false, "Test config and authentication")
	flag.BoolVar(&log.Quiet, "quiet", false, "Do not display info messages")
	flag.BoolVar(&log.DebugLevel, "debug", false, "Display debug messages")
	flag.BoolVar(&dryrun, "dryrun", false, "Test run, shows what will happen")
	flag.BoolVar(&setup, "init", false, "Create settings for blog and auth")
	flag.BoolVar(&confirm, "confirm", false, "Confirm prompt before upload")
	flag.Parse()

	if *helpFlag {
		usage()
	}

	if *versionFlag {
		fmt.Println("wpsync v0.2.0")
		os.Exit(0)
	}

	CrDirIfNtExst("./pages")
	CrDirIfNtExst("./posts")
	CrDirIfNtExst("./media")

	file, err := ioutil.ReadFile("wpsync.json")
	if err != nil {
		log.Debug("wpsync.json file not found, running setup", err)
		setup = true
	} else {
		if err := json.Unmarshal(file, &conf); err != nil {
			log.Fatal("Error parsing wpsync.json", err)
		}
	}

	if *testFlag {
		if testSetup() {
			fmt.Println("Test setup passed. 👍")
			os.Exit(0)
		} else {
			fmt.Println("Test setup fail. 👎")
			os.Exit(1)
		}
	}

	if setup {
		runSetup()
	}

	// test setup
	if !testSetup() {
		// setup not working
		// check if runSetup() ran with setup
		// if not run it now otherwise bail
		log.Fatal("Error validating.", conf)
	}

}

// route command and args
func main() {

	// custom init, this allows testing override
	// go test will always run init()
	myInit()

	// posts
	localPosts := getLocalPosts()
	if len(localPosts) > 0 {
		remotePosts := getRemotePosts()
		newPosts, updatedPosts := comparePosts(localPosts, remotePosts)
		if !dryrun {
			newPosts = loadPostsFromFiles(newPosts)
			newPosts = createPosts(newPosts)

			updatedPosts = loadPostsFromFiles(updatedPosts)
			updatedPosts = updatePosts(updatedPosts)

			writeRemotePosts(newPosts, updatedPosts)
		}
	}

	// pages
	localPages := getLocalPages()
	if len(localPages) > 0 {
		remotePages := getRemotePages()
		newPages, updatedPages := comparePages(localPages, remotePages)
		if !dryrun {
			newPages = loadPagesFromFiles(newPages)
			newPages = createPages(newPages)

			updatedPages = loadPagesFromFiles(updatedPages)
			updatedPages = updatePages(updatedPages)

			writeRemotePages(newPages, updatedPages)
		}
	}

	// media
	localMedia := getLocalMedia()
	if len(localMedia) > 0 {
		remoteMedia := getRemoteMedia()
		newMedia := compareMedia(localMedia, remoteMedia)

		if !dryrun {
			uploadedMedia := uploadMediaItems(newMedia)
			writeRemoteMedia(uploadedMedia)
		}
	}
}

func confirmPrompt(prompt string) bool {

	// confirmation not required
	if !confirm {
		return true
	}

	var ans string
	fmt.Print(prompt)
	ans = readLine()
	if ans == "y" || ans == "Y" {
		return true
	} else {
		return false
	}
}

// Display Usage
func usage() {
	fmt.Println("usage: wpsync [args]")
	fmt.Println("Arguments:")
	flag.PrintDefaults()
	fmt.Println("")
	os.Exit(0)
}

// If no idea what dirs to create
func CrDirIfNtExst(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			log.Fatal("Unable to create dir: ", err)
		}
	}
}
