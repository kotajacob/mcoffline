package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Tnze/go-mc/offline"
)

// OfflineSuffix is the suffix added to whitelist.json and playerdata.
const OfflineSuffix = ".offline"

// User represents a user in whitelist.json or ops.json
type User struct {
	Name              string `json:"name"`
	UUID              string `json:"uuid"`
	Level             int    `json:"level,omitempty"`
	BypassPlayerLimit bool   `json:"bypassPlayerLimit,omitempty"`
}

// Version is set at build time in the Makefile.
var Version string

func main() {
	log.SetPrefix("mcoffline " + Version + ": ")
	log.SetFlags(0)
	flag.Parse()

	arg := flag.Arg(0)
	if arg == "" {
		// Look for whitelist.json in current directory.
		arg = "whitelist.json"
	}

	users, err := loadWhitelist(arg)
	if errors.Is(err, os.ErrNotExist) && flag.Arg(0) != "" {
		// Assume the argument was a username if the file does not exist, but an
		// argument was given.
		fmt.Println(arg, offline.NameToUUID(arg))
		return
	} else if err != nil {
		log.Fatalf("failed opening whitelist: %v\n", err)
	}
	basePath := filepath.Dir(arg)

	opsPath := filepath.Join(basePath, "ops.json")
	in, err := os.Open(opsPath)
	if err != nil {
		log.Fatalf("failed opening ops.json: %v\n", err)
	}
	defer in.Close()

	src, err := io.ReadAll(in)
	if err != nil {
		log.Fatalf("failed reading ops.json: %v\n", err)
	}

	if err = createOfflineJSON(src, opsPath+OfflineSuffix); err != nil {
		log.Fatalf("failed to create ops.json.offline: %v\n", err)
	}

	worldName, err := getLevelName(filepath.Join(basePath, "server.properties"))
	if err != nil {
		log.Fatalf("failed reading server.properties to get world name: %v\n", err)
	}

	advancementsPath := filepath.Join(basePath, worldName, "advancements")
	if err := convertDirectory(users, advancementsPath); err != nil {
		log.Fatalf("failed creating offline advancements: %v\n", err)
	}

	playerdataPath := filepath.Join(basePath, worldName, "playerdata")
	if err := convertDirectory(users, playerdataPath); err != nil {
		log.Fatalf("failed creating offline playerdata: %v\n", err)
	}

	statsPath := filepath.Join(basePath, worldName, "stats")
	if err := convertDirectory(users, statsPath); err != nil {
		log.Fatalf("failed creating offline stats: %v\n", err)
	}
}

// loadWhitelist reads an online-mode whitelist file by path, creates an
// offline mode whitelist in the same directory, and returns a map of
// online-mode UUIDs to usernames.
func loadWhitelist(path string) (map[string]User, error) {
	in, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer in.Close()

	src, err := io.ReadAll(in)
	if err != nil {
		return nil, err
	}

	if err = createOfflineJSON(src, path+OfflineSuffix); err != nil {
		return nil, err
	}

	// Reset file seek so we can read it again to make a map of the UUIDs.
	if _, err := in.Seek(0, io.SeekStart); err != nil {
		log.Fatalf("failed reading whitelist: %v\n", err)
	}
	return mapUsers(in)
}

// createOfflineJSON reads online-mode .json data (whitelist or ops) and writes
// an equivalent offline-mode .json file to the path specified.
func createOfflineJSON(src []byte, path string) error {
	in := bytes.NewReader(src)
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	if err := convertJSON(in, out); err != nil {
		return fmt.Errorf("failed converting json file: %w\n", err)
	}
	return nil
}

// convertJSON reads online-mode .json (whitelist or ops) from a reader and
// writes an equivalent offline-mode .json to it's writer. The output is
// indented with 2 spaces to match the format Mojang seems to like.
func convertJSON(in io.Reader, out io.Writer) error {
	var users []User
	dec := json.NewDecoder(in)
	err := dec.Decode(&users)
	if err != nil {
		return err
	}

	offlineUsers := make([]User, len(users))
	for i, user := range users {
		// Do some type checking as a treat (and to avoid crashing if the json
		// is corrupted).
		offlineUsers[i].UUID = offline.NameToUUID(user.Name).String()
		offlineUsers[i].Name = user.Name
		offlineUsers[i].Level = user.Level
		offlineUsers[i].BypassPlayerLimit = user.BypassPlayerLimit
	}

	e := json.NewEncoder(out)
	e.SetIndent("", "  ") // MC uses 2 space indented json.
	return e.Encode(&offlineUsers)
}

// mapUsers reads a whitelist and creates a map of online-mode UUIDs to Users.
func mapUsers(in io.Reader) (map[string]User, error) {
	var users []User
	d := json.NewDecoder(in)
	err := d.Decode(&users)
	if err != nil {
		return nil, err
	}

	m := make(map[string]User, len(users))
	for _, user := range users {
		m[user.UUID] = user
	}
	return m, nil
}

// getWorld takes the path to a server.properties file, reads it, and returns
// the string value for the "level-name" key.
func getLevelName(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		i := strings.Index(line, "=")
		if i < 1 {
			continue
		}

		key := strings.TrimSpace(line[:i])
		if key == "level-name" {
			var value string
			if len(line) > i {
				value = strings.TrimSpace(line[i+1:])
			}
			return value, nil
		}
	}
	return "", fmt.Errorf("level-name not found in %v", path)
}

// convertDirectory takes a map of online-mode UUIDs to player names and the
// path to a folder with player data (stats, advancements, or playerdata). It
// creates a new directory with a .offline suffix. Then, For each file in the
// input directory, a hard link is made to the offline directory named with the
// offline-mode UUID instead of it's previous online-mode UUID.
func convertDirectory(users map[string]User, path string) error {
	files, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("failed reading folder: %w", err)
	}

	offlineDir := filepath.Clean(path) + OfflineSuffix
	err = os.Mkdir(offlineDir, 0755)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return fmt.Errorf("failed creating %v: %w", offlineDir, err)
	}

	for _, file := range files {
		// Skip non-files... there really shouldn't be any, but who knows.
		if file.IsDir() {
			continue
		}

		onlinePath := filepath.Join(path, file.Name())

		ext := filepath.Ext(file.Name())
		uuid := strings.TrimSuffix(file.Name(), ext)
		user, ok := users[uuid]
		if !ok {
			fmt.Printf("skipping non-whitelisted player: %v\n", onlinePath)
			continue
		}

		offlinePath := filepath.Join(
			offlineDir,
			offline.NameToUUID(user.Name).String()+ext,
		)

		// Using a hardlink to avoid pointless disk writing. Dunno if this works
		// on windows lol. I guess you can always send me an angry email if this
		// doesn't work for you :P
		if err := os.Link(
			onlinePath,
			offlinePath,
		); err != nil && !errors.Is(err, os.ErrExist) {
			return fmt.Errorf("failed creating offline version of %v named %v: %w",
				onlinePath,
				offlinePath,
				err,
			)
		}
	}
	return nil
}
