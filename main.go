package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func main() {
	token, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	discord, err := discordgo.New(strings.TrimSpace(string(token)))
	if err != nil {
		panic(err)
	}

	discord.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		dir := path.Join(
			"storage",
			m.ChannelID,
			m.Author.ID,
			m.ID,
		)

		log.Print("Saving message ", dir)

		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Printf("Could not create message directory %s: %s", dir, err)
			return
		}

		raw, err := json.Marshal(*m.Message)
		if err != nil {
			log.Print("Could not convert message to JSON: ", err)
		} else {
			if err := ioutil.WriteFile(path.Join(dir, "msg.json"), raw, 0644); err != nil {
				log.Print("Could not write JSON: ", err)
				fmt.Println(string(raw))
			}
		}

		now := m.Timestamp.String() + "\n"
		if err := ioutil.WriteFile(path.Join(dir, "timestamp"), []byte(now), 0644); err != nil {
			log.Printf("Could not write timestamp %s: %s", now, err)
		}

		if err := ioutil.WriteFile(path.Join(dir, "content"), []byte(m.Content + "\n"), 0644); err != nil {
			log.Printf("Could not write content \"%s\": %s", m.Content, err)
		}

		for i, a := range m.Attachments {
			resp, err := http.Get(a.URL)
			if err != nil {
				log.Printf("Could not download attachment %s: %s", a.URL, err)
				continue
			}
			defer resp.Body.Close()

			name := strconv.Itoa(i)
			url, err := url.Parse(a.URL)
			if err != nil {
				log.Printf("Could not parse URL %s: err", a.URL)
			} else {
				name += "_" + path.Base(url.Path)
			}

			file, err := os.Create(path.Join(dir, name))
			if err != nil {
				log.Printf("Could not open file %s: %s", name, err)
				continue
			}
			defer file.Close()

			if _, err := io.Copy(file, resp.Body); err != nil {
				log.Printf("Could not save attachment to file %s: %s", name, err)
			}
		}
	})
	discord.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentDirectMessages
	if err := discord.Open(); err != nil {
		panic(err)
	}

	log.Print("Press enter to stop.")
	fmt.Scanln()
	log.Print("Shutting down...")
	discord.Close()
}
