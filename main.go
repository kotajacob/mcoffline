package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/Tnze/go-mc/offline"
)

var Version string

type User struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
}

func main() {
	log.SetPrefix("mcoffline " + Version + ": ")
	log.SetFlags(0)
	flag.Parse()
	if len(flag.Args()) == 0 {
		// Use STDIN and STDOUT.
		if err := convertWhitelist(os.Stdin, os.Stdout); err != nil {
			log.Fatalf("failed parsing whitelist: %v", err)
		}
		return
	}

	arg := flag.Arg(0)
	in, err := os.Open(arg)
	if errors.Is(err, os.ErrNotExist) {
		// Assume the argument was a username if the file does not exist
		fmt.Println(arg, offline.NameToUUID(arg))
		return
	} else if err != nil {
		log.Fatalf("failed opening whitelist: %v", err)
	}
	defer in.Close()

	// Use second arg as output file if given
	out := os.Stdout
	if flag.Arg(1) != "" {
		f, err := os.OpenFile(flag.Arg(1), os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			log.Fatalf("failed opening output: %v", err)
		}
		defer f.Close()
		out = f
	}

	if err := convertWhitelist(in, out); err != nil {
		log.Fatalf("failed parsing whitelist: %v", err)
	}
}

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
