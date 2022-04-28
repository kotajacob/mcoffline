package main

import (
	"bufio"
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

// Version is set at build time in the Makefile.
var Version string

// User represents a user in whitelist.json
type User struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
}

func main() {
	log.SetPrefix("mcoffline " + Version + ": ")
	log.SetFlags(0)
	flag.Parse()

	arg := flag.Arg(0)
	if arg == "" {
		// Look for whitelist.json in current directory.
		arg = "whitelist.json"
	}

	in, err := os.Open(arg)
	if errors.Is(err, os.ErrNotExist) && flag.Arg(0) != "" {
		// Assume the argument was a username if the file does not exist, but an
		// argument was given.
		fmt.Println(arg, offline.NameToUUID(arg))
		return
	} else if err != nil {
		log.Fatalf("failed opening whitelist: %v\n", err)
	}
	defer in.Close()

	out, err := os.Create(arg + OfflineSuffix)
	if err != nil {
		log.Fatalf("failed creating offline-mode whitelist file: %v\n", err)
	}
	defer out.Close()

	if err := convertWhitelist(in, out); err != nil {
		log.Fatalf("failed parsing whitelist: %v\n", err)
	}

	// Reset file seek so we can read it again later to make a map of the UUIDs.
	if _, err := in.Seek(0, io.SeekStart); err != nil {
		log.Fatalf("failed reading whitelist: %v\n", err)
	}

	basePath := filepath.Dir(arg)
	worldName, err := getLevelName(filepath.Join(basePath, "server.properties"))
	if err != nil {
		log.Fatalf("failed reading server.properties to get world name: %v\n", err)
	}
	playerDataPath := filepath.Join(basePath, worldName, "playerdata")

	users, err := mapUsers(in)
	if err != nil {
		log.Fatalln("failed mapping users to UUIDs")
	}

	if err := convertPlayerdata(users, playerDataPath); err != nil {
		log.Fatalf("failed creating offline playerdata: %v\n", err)
	}
}

// convertWhitelist reads an online-mode whitelist.json file from a reader and
// writes an equivalent offline-mode whitelist.json to it's writer. The output
// is indented with 2 spaces to match the format Mojang seems to like.
func convertWhitelist(in io.Reader, out io.Writer) error {
	var users []User
	d := json.NewDecoder(in)
	err := d.Decode(&users)
	if err != nil {
		return err
	}

	for i, user := range users {
		var offlineUser User
		offlineUser.UUID = offline.NameToUUID(user.Name).String()
		offlineUser.Name = user.Name
		users[i] = offlineUser
	}

	e := json.NewEncoder(out)
	e.SetIndent("", "  ") // MC uses 2 space indented json.
	return e.Encode(&users)
}

// mapUsers reads a whitelist and creates a map of UUIDs to usernames.
func mapUsers(in io.Reader) (map[string]string, error) {
	var users []User
	d := json.NewDecoder(in)
	err := d.Decode(&users)
	if err != nil {
		return nil, err
	}

	m := make(map[string]string)
	for _, user := range users {
		m[user.UUID] = user.Name
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

// convertPlayerdata takes a map of online-mode UUIDs to player named and the
// path to a playerdata folder. It creates a playerdata.offline directory.
// Then, For each file in the input directory, a hard link is make to the
// offline directory named with the offline-mode UUID instead of it's previous
// online-mode UUID.
func convertPlayerdata(users map[string]string, path string) error {
	files, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("failed reading playerdata folder: %w", err)
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
		name, ok := users[uuid]
		if !ok {
			fmt.Printf("skipping non-whitelisted player: %v\n", onlinePath)
			continue
		}

		offlinePath := filepath.Join(
			offlineDir,
			offline.NameToUUID(name).String()+ext,
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
