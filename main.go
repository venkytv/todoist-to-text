package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/go-resty/resty/v2"
	"github.com/keybase/go-keychain"
	"github.com/venkytv/go-config"
)

const (
	INSTAPAPER_USERNAME = "venky@duh-uh.com"
	TODOIST_API         = "https://api.todoist.com/rest/v1"
)

type Task struct {
	Id          int64  `json:"id"`
	Content     string `json:"content"`
	Description string `json:"description"`
	Name        string
	Url         string
	Note        string
}

func (t *Task) Parse(cfg *config.Config) {
	var tags []string
	for _, tag := range strings.Split(cfg.GetString("tags"), ",") {
		tags = append(tags, "#"+strings.TrimSpace(tag))
	}
	if len(tags) > 0 {
		t.Content += " " + strings.Join(tags, " ")
	}

	re := regexp.MustCompile(`\[([^\]]*)\]\(([^\)]*)\)\s*(.*)`)
	m := re.FindStringSubmatch(t.Content)
	if len(m) < 1 {
		log.Print("Regexp match fail: ", t.Content)
		t.Name = t.Content
	} else {
		t.Name = m[1]
		t.Url = m[2]
		t.Note = m[3]
	}
}

func (t Task) Out() string {
	var out string
	if len(t.Url) > 0 {
		out = fmt.Sprintf("[%s](%s)", t.Name, t.Url)
		if len(t.Note) > 0 {
			out += " " + t.Note
		}
	} else {
		out = t.Name
	}
	return out
}

type Bookmark struct {
	Id int64 `json:"bookmark_id"`
}

func getTodoistApiToken(cfg *config.Config) string {
	token := cfg.GetString("api-token")
	if len(token) < 1 {
		item, err := keychain.GetGenericPassword("todoist", "api-token", "", "")
		if err != nil {
			log.Fatal(err)
		}
		token = string(item)
	}
	return token
}

func getInstapaperCredentials(cfg *config.Config) (string, string) {
	log.Print(cfg.AllKeys())
	username := cfg.GetString("instapaper-username")

	password := cfg.GetString("instapaper-password")
	if len(password) < 1 {
		item, err := keychain.GetGenericPassword("https://instapaper.com",
			username, "", "")
		if err != nil {
			log.Fatal(err)
		}
		password = string(item)
	}

	log.Print(username)
	return username, password
}

func postToInstapaper(cfg *config.Config, client *resty.Client, url string) (string, error) {
	username, password := getInstapaperCredentials(cfg)
	log.Print(username, password)

	log.Print("Posting to Instapaper: ", url)
	request := client.R().SetQueryParams(map[string]string{
		"username": username,
		"password": password,
		"url":      url,
	})
	resp, err := request.
		Get("https://www.instapaper.com/api/add")
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode() != 201 {
		log.Fatal(resp.Status())
	}

	var bookmark Bookmark
	err = json.Unmarshal(resp.Body(), &bookmark)
	if err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("https://www.instapaper.com/read/%d", bookmark.Id), nil
}

func getTask(cfg *config.Config, client *resty.Client, task_id string) Task {
	token := getTodoistApiToken(cfg)
	request := client.R().SetHeader("Accept", "application/json").SetAuthToken(token)

	log.Print("Fetching tasks for ID: ", task_id)
	var task Task
	req_url := fmt.Sprintf("%s/tasks/%s", TODOIST_API, task_id)
	resp, err := request.SetResult(&task).Get(req_url)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode() != 200 {
		log.Fatal(resp.Status())
	}

	task.Parse(cfg)
	if len(task.Url) > 0 {
		url, err := postToInstapaper(cfg, client, task.Url)
		if err == nil {
			log.Print("Instapaper Url: ", url)
			task.Url = url
		}
	}

	log.Printf("Closing task %s: %s", task_id, task.Out())
	resp, err = request.Post(req_url + "/close")
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode() != 204 {
		log.Fatal(resp.Status())
	}

	return task
}

func loadConfig() *config.Config {
	f := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	f.String("api-token", "", "todoist API token")
	f.String("instapaper-username", INSTAPAPER_USERNAME, "Instapaper username")
	f.String("instapaper-password", "", "Instapaper password")
	f.String("tags", "", "comma-separated list of tags to add to the output")

	return config.Load(f, "TTT")
}

func main() {
	cfg := loadConfig()
	args := cfg.Args()
	log.Printf("%v | %v\n", cfg, args)
	if len(args) < 1 {
		log.Fatal("Task ID argument mandatory")
	}

	client := resty.New()
	task_id := args[0]

	t := getTask(cfg, client, task_id).Out()
	clipboard.WriteAll(t)
	fmt.Println(t)
}
