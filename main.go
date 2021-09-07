package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"github.com/gonejack/get"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/tumblr/tumblr.go"
	"github.com/tumblr/tumblrclient.go"
)

var (
	config   string
	outdir   string
	template bool
	verbose  bool

	cmd = &cobra.Command{
		Use:   "tumblr-likes",
		Short: "Save tumblr liked images",
		Run: func(c *cobra.Command, args []string) {
			err := run(c, args)
			if err != nil {
				log.Fatal(err)
			}
		},
	}
)

func init() {
	log.SetOutput(os.Stdout)

	fs := cmd.Flags()
	fs.SortFlags = false
	fs.StringVarP(&config, "config", "c", "config.json", "config file")
	fs.StringVarP(&outdir, "outdir", "o", "likes", "output directory")
	fs.BoolVarP(&template, "template", "t", false, "print config template")
	fs.BoolVarP(&verbose, "verbose", "v", false, "verbose")
}
func run(command *cobra.Command, args []string) (err error) {
	var c struct {
		ConsumerKey    string `json:"consumer_key"`
		ConsumerSecret string `json:"consumer_secret"`
		Token          string `json:"token"`
		TokenSecret    string `json:"token_secret"`
	}
	if template {
		bytes, _ := json.MarshalIndent(c, "", "    ")
		fmt.Printf("%s", bytes)
		return
	}

	// parse config
	bytes, err := ioutil.ReadFile(config)
	{
		if err == nil {
			err = json.Unmarshal(bytes, &c)
		}
		if err != nil {
			return fmt.Errorf("read config file failed: %s", err)
		}
	}

	client := tumblrclient.NewClientWithToken(c.ConsumerKey, c.ConsumerSecret, c.Token, c.TokenSecret)

	posts, err := fetch(client)
	if err != nil {
		return
	}
	reverse(posts)
	err = download(posts)
	if err != nil {
		return
	}

	return
}
func fetch(client *tumblrclient.Client) (posts []tumblr.PostInterface, err error) {
	const limit = 50

	param := make(url.Values)
	param.Set("limit", cast.ToString(limit))

	max := 1000
	offset := 0
	for len(posts) < max {
		param.Set("offset", cast.ToString(offset))

		if verbose {
			log.Printf("fetch post %d-%d/%d", len(posts), len(posts)+limit, max)
		}

		likes, err := tumblr.GetLikes(client, param)
		if err != nil {
			return nil, err
		}
		parsed, err := likes.Full()
		if err != nil {
			return nil, err
		}
		posts = append(posts, parsed...)

		offset += len(likes.Posts)
		if int(likes.TotalLikes) < max {
			max = int(likes.TotalLikes)
		}
	}

	return
}
func download(posts []tumblr.PostInterface) (err error) {
	if verbose {
		log.Printf("process %d posts", len(posts))
	}

	tasks := get.NewDownloadTasks()
	for _, i := range posts {
		switch post := i.(type) {
		case *tumblr.PhotoPost:
			for _, photo := range post.Photos {
				tasks.Add(photo.OriginalSize.Url, filepath.Join(outdir, filepath.Base(photo.OriginalSize.Url)))
			}
		}
	}

	g := get.Default()
	{
		g.OnEachStart = func(t *get.DownloadTask) {
			if verbose {
				log.Printf("save %s as %s", t.Link, t.Path)
			}
		}
	}
	g.Batch(tasks, 5, time.Minute)

	tasks.ForEach(func(t *get.DownloadTask) {
		if t.Err != nil {
			log.Printf("save %s as %s failed: %s", t.Link, t.Path, t.Err)
		}
	})

	return
}
func reverse(s interface{}) {
	size := reflect.ValueOf(s).Len()
	swap := reflect.Swapper(s)
	for i, j := 0, size-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
}
func main() {
	_ = cmd.Execute()
}
