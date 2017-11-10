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

const version = "0.2.0"

func train(lines *[]string) error {
	// Go over all the lines from the master file.

	// Set up channels listening to STDIN in the background.
	// These will persist even as STDIN is reconfigured throughout the program.
	inputCh := make(chan rune)
	errCh := make(chan error)
	go func() {
		reader := bufio.NewReader(os.Stdin)
		// TODO: handle interruption more safely.
		for {
			i, _, e := reader.ReadRune()
			if i == 3 { errCh <- errors.New("Ctrl+c") }
			if i == 27 { errCh <- errors.New("ESC") }
			if e != nil { errCh <- e }
			inputCh <- i
		}
	}()

	fmt.Println("Copy each line or use Ctrl+c or ESC to quit.")
	fmt.Println("————————————————————————————————————————————")
	for _, line := range *lines {
		err := haveUserCopyWord(inputCh, errCh, line)
		if err != nil { return err }
	}
	return nil
}

func haveUserCopyWord(inputCh chan rune, errCh chan error, line string) error {
	// Monitor the copying of a single line.
	fmt.Println(line)

	undoRaw, err := makeInputRaw()
	if err != nil { return err }

	loop:
	for _, character := range line {
		select {
		case input := <-inputCh:
			if input == character {
				fmt.Print(color.GreenString("▀"))
			} else {
				// User hit the wrong key. Show the error and wind down.
				fmt.Printf("%s %s %s",
					color.New(color.Bold).Sprint(string(input)),
					color.New(color.Bold, color.FgHiRed).Sprint("≠"),
					color.New(color.Bold).Sprint(string(character)))

				err = discardFurtherKeystrokes(inputCh, errCh)
				break loop
			}
		case err = <-errCh:
			break loop
		}
	}

	undoRaw()
	fmt.Println()
	return err
}

func makeInputRaw() (func(), error) {
	oldSettings, err := raw.MakeRaw(os.Stdin.Fd())
	return func() { raw.TcSetAttr(os.Stdin.Fd(), oldSettings) }, err
}

func discardFurtherKeystrokes(inputCh chan rune, errCh chan error) error {
	timeout := time.After(1 * time.Second)
	for {
		select {
		case <-inputCh:
		case err := <-errCh:
			return err
		case <-timeout:
			return nil
		}
	}
}

func main() {
	help := flag.Bool("h", false, "print this help message")
	inOrder := flag.Bool("o", false, "process file in order; no shuffling")
	file := flag.String("f", "/etc/dictionaries-common/words",
		"source of words or phrases etc. to type; one per line")
	flag.Parse()

	if *help {
		fmt.Printf("stricttypist v%s\n", version)
		fmt.Printf("Learn to type more accurately with immediate feedback.\n")
		flag.PrintDefaults()
		return
	}

	// Get requested master text.
	data, err := ioutil.ReadFile(*file)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 {
		fmt.Println("No lines in file.")
		os.Exit(1)
	}

	if ! *inOrder {
		// Shuffle. Fisher-Yates.
		randomizer := rand.New(rand.NewSource(time.Now().UnixNano()))
		for i := range lines {
		    j := randomizer.Intn(i + 1)
		    lines[i], lines[j] = lines[j], lines[i]
		}
	}

	err = train(&lines)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
