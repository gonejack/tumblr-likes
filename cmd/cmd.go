package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"path/filepath"
	"reflect"
	"time"

	"github.com/alecthomas/kong"
	"github.com/gonejack/get"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/spf13/cast"
	"github.com/tumblr/tumblr.go"
	"github.com/tumblr/tumblrclient.go"
	"github.com/uniplaces/carbon"
)

type options struct {
	Config   string `short:"c" default:"config.json" help:"Config file."`
	Output   string `short:"o" default:"likes" help:"Output directory."`
	MaxFetch int    `short:"m" default:"200" help:"How many likes to be fetched."`
	Template bool   `short:"t" help:"Print config template."`
	Verbose  bool   `short:"v" help:"Verbose printing."`
	About    bool   `help:"Show about."`
}
type credentials struct {
	ConsumerKey    string `json:"consumer_key"`
	ConsumerSecret string `json:"consumer_secret"`
	Token          string `json:"token"`
	TokenSecret    string `json:"token_secret"`
}
type record struct {
	gorm.Model
	URL string
}

type command struct {
	options
	credentials

	client *tumblrclient.Client
	db     *gorm.DB
}

func (c *command) Run() (err error) {
	kong.Parse(&c.options,
		kong.Name("tumblr-likes"),
		kong.Description("Save tumblr liked images."),
		kong.UsageOnError(),
	)
	return c.run()
}
func (c *command) run() (err error) {
	switch {
	case c.Template:
		bytes, _ := json.MarshalIndent(&c.credentials, "", "    ")
		fmt.Printf("%s", bytes)
		return
	case c.About:
		fmt.Println("Visit https://github.com/gonejack/tumblr-likes")
		return
	}

	dbname := "record.db"
	c.db, err = gorm.Open("sqlite3", dbname)
	if err != nil {
		return fmt.Errorf("open db file %s failed: %s", dbname, err)
	}
	c.db.AutoMigrate(new(record))
	c.db.Unscoped().Delete(new(record), "updated_at < ?", carbon.Now().SubYear().String())
	defer c.db.Close()

	// parse config
	bytes, err := ioutil.ReadFile(c.Config)
	if err == nil {
		err = json.Unmarshal(bytes, &c.credentials)
	}
	if err != nil {
		return fmt.Errorf("read config file failed: %s", err)
	}
	c.client = tumblrclient.NewClientWithToken(c.ConsumerKey, c.ConsumerSecret, c.Token, c.TokenSecret)

	posts, err := c.fetch()
	if err != nil {
		return
	}

	err = c.download(posts)
	if err != nil {
		return
	}

	return
}
func (c *command) fetch() (posts []tumblr.PostInterface, err error) {
	const limit = 50

	param := make(url.Values)
	param.Set("limit", cast.ToString(limit))

	want := c.MaxFetch
	offset := 0
	for len(posts) < want {
		param.Set("offset", cast.ToString(offset))

		if c.Verbose {
			log.Printf("fetch likes %03d-%03d/%d", len(posts), len(posts)+limit, want)
		}

		likes, err := tumblr.GetLikes(c.client, param)
		if err != nil {
			return nil, err
		}
		full, err := likes.Full()
		if err != nil {
			return nil, err
		}
		posts = append(posts, full...)

		offset += len(likes.Posts)
		if int(likes.TotalLikes) < want {
			want = int(likes.TotalLikes)
		}
	}

	if len(posts) > 0 {
		reverse(posts)
	}

	return
}
func (c *command) download(posts []tumblr.PostInterface) (err error) {
	if c.Verbose {
		log.Printf("process %d posts", len(posts))
	}

	tasks := get.NewDownloadTasks()
	for _, p := range posts {
		switch post := p.(type) {
		case *tumblr.PhotoPost:
			for _, photo := range post.Photos {
				r := new(record)
				c.db.First(r, "url == ?", photo.OriginalSize.Url)
				if r.URL == "" {
					tasks.Add(photo.OriginalSize.Url, filepath.Join(c.Output, filepath.Base(photo.OriginalSize.Url)))
				}
			}
		}
	}
	geter := get.Default()
	geter.OnEachStart = func(t *get.DownloadTask) {
		if c.Verbose {
			log.Printf("save %s", t.Link)
		}
	}
	geter.Batch(tasks, 5, time.Minute)
	tasks.ForEach(func(t *get.DownloadTask) {
		if t.Err == nil {
			c.db.Save(&record{URL: t.Link})
		} else {
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

func New() *command {
	return new(command)
}
