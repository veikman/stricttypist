package main

// This file constitutes the source code of Stricttypist.
//
// Stricttypist is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Stricttypist is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Stricttypist.  If not, see <http://www.gnu.org/licenses/>.
//
// Copyright 2017 Viktor Eikman

import (
	"errors"
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"math/rand"
	"strings"
	"time"

    "github.com/creack/termios/raw"
	"github.com/fatih/color"
)

const version = "0.1.0"

func train(filepath string) error {
	data, err := ioutil.ReadFile(filepath)
	if err != nil { return err }

	reader := bufio.NewReader(os.Stdin)
	randomizer := rand.New(rand.NewSource(time.Now().UnixNano()))
	lines := strings.Split(string(data), "\n")
	n := len(lines)
	if n == 0 { return errors.New("No words in file.") }

	fmt.Printf("Picking randomly from %d words.\n", n)
	fmt.Println("Copy each word, or use Ctrl+C or ESC to quit.")
	fmt.Println("—————————————————————————————————————————————")
	for {
		err := haveUserCopyWord(reader, lines[randomizer.Intn(n)])
		if err != nil { return err }
	}
	return nil
}

func haveUserCopyWord(reader *bufio.Reader, word string) error {
	fmt.Println(word)
	for _, character := range word {
		input, err := readRune(reader)
		if err != nil { return err }

		if input == character {
			fmt.Print(color.GreenString("▀"))
		} else {
			fmt.Printf("%s not %s!\n",
				color.New(color.Bold, color.FgWhite).Sprint(string(input)),
				color.New(color.Bold, color.FgWhite).Sprint(string(character)))

			continuation := time.Now().Add(time.Second)
			for time.Now().Before(continuation) {
				_, err = readRune(reader)
				if err != nil { return err }
			}
			return err
		}
	}
	fmt.Println()
	return nil
}

func readRune(reader *bufio.Reader) (rune, error) {
	oldSettings, err := raw.MakeRaw(os.Stdin.Fd())
	if err != nil { return 0, err }
	defer raw.TcSetAttr(os.Stdin.Fd(), oldSettings)
	input, _, err := reader.ReadRune()

	// TODO: handle interruption more safely.
	if input == 3 { return 0, errors.New("Ctrl+C") }
	if input == 27 { return 0, errors.New("ESC") }

	return input, err
}

func main() {
	help := flag.Bool("h", false, "print this help message")
	file := flag.String("f", "/etc/dictionaries-common/words",
		"source of words or phrases etc. to type; one per line")
	flag.Parse()

	if *help {
		fmt.Printf("stricttypist v%s\n", version)
		fmt.Printf("Learn to type more accurately with immediate feedback.\n")
		flag.PrintDefaults()
		return
	}

	err := train(*file)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
