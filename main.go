package main

import (
	"bufio"
	"errors"
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

// 出力内容を出力する
func printLines(file *os.File) error {

	b := bufio.NewReader(file)

	for {
		line, err := b.ReadString('\n')
		if err != nil {
			fmt.Print(line)
			break
		}

		fmt.Print(line)

	}
	return nil
}

// 出力行数に応じた対象ファイルの出力開始地点を決定する
// ファイル末尾から1byteずつ読みこみ、改行コードが存在すれば出力行数をデクリメントし、
// 出力行数がゼロになった時点で出力開始地点とみなす

// 現状、1byteずつ読み込んでいるが、バッファリングなどの改善をしていきたい
func startPoint(lines int, file *os.File) error {

	info, err := file.Stat()
	if err != nil {
		return err
	}

	size := info.Size() - 1
	var offset int64
	buf := make([]byte, 1)

	// 対象ファイルの末尾に改行コードのみ存在した場合
	// 既存のtailコマンドではその行を無視して、出力行としてカウントしていない挙動をしている。
	// 改行コードが存在すれば出力行数をデクリメントする実装とした場合、既存のtailコマンドと比較すると
	// 出力される行数が1行減ってしまうため、この for ループではファイル末尾に改行コードのみ存在した場合
	// 出力行数に含めないようにしている
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

	// ファイルの出力開始位置を決定する
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

	// 出力行数より多い行数を持つファイルの場合、出力する際に余計な改行コードが入ってしまうため
	// 意図的に出力開始位置をインクリメントしている。
	// [TODO] 修正したいが、現状どうすればよいか理解できていない。
	if offset != 0 {
		size++
		size++
		file.Seek(size, os.SEEK_SET)
	}

	return nil
}

func tailFiles(lines int, name string, printHeaders bool) error {

	if printHeaders {
		fmt.Printf("==> %s <==\n", name)
	}

	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()

	if err = startPoint(lines, f); err != nil {
		return err
	}

	if err = printLines(f); err != nil {
		return err
	}

	return nil
}

// 設定内容に応じて動作を変更し、実際のtail処理をtailFiles関数に委譲する
func tail(c *config) error {

	var l int

	// 表示する行数の設定
	// 0以下の値が引数として設定されている場合には、デフォルトの行数（10）を表示する
	if c.lines > 0 {
		l = c.lines
	} else {
		l = defaultLines
	}

	// 複数ファイルが指定されている場合に、ファイル名をヘッダ情報として出力する
	var printHeaders bool
	if len(c.files) > 1 {
		printHeaders = true
	}

	for i, f := range c.files {
		// 複数行ファイルを対象とした際に、tailコマンドの表示に合わせるため改行を出力する
		if i >= 1 {
			println()
		}
		if err := tailFiles(l, f, printHeaders); err != nil {
			return err
		}
	}

	return nil
}

// コマンドライン引数を解析する関数
func parseArgs(args []string) (*config, error) {

	var config config

	for _, v := range args {

		// コマンドライン引数の解析
		// -nオプションのみに対応し、現状それ以外の引数はtail対象のファイルとして扱う
		if strings.HasPrefix(v, "-n") {
			arg := strings.Split(v, "=")
			if len(arg) < 2 {
				return nil, errors.New("-n=出力行数 の形式で指定してください")
			}
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

	if err = tail(c); err != nil {
		log.Fatal(err)
		os.Exit(-1)
	}

}
