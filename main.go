package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

const (
	defaultLines = 10
)

type config struct {
	lines int
	files []string
}

func printLines(file *os.File) error {

	b := bufio.NewReader(file)

	for {
		line, err := b.ReadBytes('\n')
		if err != nil {
			fmt.Println(string(line))
			break
		}

		tmp := strings.Trim(string(line), "\r\n")
		if len(tmp) == 0 {
			continue
		}
		fmt.Println(tmp)
	}
	return nil
}

func offset(lines int, file *os.File) (int64, error) {

	info, err := file.Stat()
	if err != nil {
		return 0, err
	}

	size := info.Size() - 1
	var offset int64
	buf := make([]byte, 1)

	for {
		b := make([]byte, 1)
		offset, err = file.Seek(size, os.SEEK_SET)
		if err != nil {
			break
		}

		file.ReadAt(b, offset)
		if b[0] == '\r' || b[0] == '\n' {
			size--
		} else {
			break
		}
	}

	for lines > 0 {
		offset, err = file.Seek(size, os.SEEK_SET)
		if err != nil {
			break
		}

		file.ReadAt(buf, offset)

		if buf[0] == '\n' {
			lines--
		}

		size--
	}

	return offset, nil
}

func lines(lines int, name string, printHeaders bool) error {

	if printHeaders {
		fmt.Printf("==> %s <==\n", name)
	}

	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err = offset(lines, f); err != nil {
		return err
	}

	if err = printLines(f); err != nil {
		return err
	}

	return nil
}

func tail(c *config) error {

	var l int
	if c.lines > 0 {
		l = c.lines
	} else {
		l = defaultLines
	}

	var printHeaders bool
	if len(c.files) > 1 {
		printHeaders = true
	}

	for _, f := range c.files {
		if err := lines(l, f, printHeaders); err != nil {
			return err
		}
	}

	return nil
}

func parseArgs(args []string) (*config, error) {

	var config config

	for _, v := range args {
		if strings.HasPrefix(v, "-n") {
			arg := strings.Split(v, "=")
			n, err := strconv.Atoi(arg[1])
			if err != nil {
				return nil, err
			}
			config.lines = n
		} else {
			config.files = append(config.files, v)
		}
	}

	return &config, nil

}

func main() {

	c, err := parseArgs(os.Args[1:])
	if err != nil {
		log.Fatal(err)
		os.Exit(-1)
	}

	err = tail(c)
	if err != nil {
		log.Fatal(err)
		os.Exit(-1)
	}

}
