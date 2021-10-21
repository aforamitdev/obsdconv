package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

const (
	RuneDollar = 0x24 // $
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("引数の個数が不正です")
	}

	filename := os.Args[1]

	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0o666)
	if err != nil {
		log.Fatalf("os.Create failed %v", err)
	}
	defer file.Close()
	fileinfo, err := file.Stat()
	if err != nil {
		log.Fatalf("file.Stat failed %v", err)
	}

	content := make([]byte, fileinfo.Size())
	_, err = file.Read(content)
	if err != nil {
		log.Fatalf("file.Write failed: %v", err)
	}

	frontMatter, body := splitMarkdown([]rune(string(content)))
	newContent := replace(body)
	title := getH1(newContent)
	fmt.Printf("Title: %v\n", title)
	fmt.Printf("Front Matter: <<\n%v>>\n", string(frontMatter))

	newFile, err := os.Create("new." + filename)
	if err != nil {
		log.Fatalf("os.Create failed: %v", err)
	}
	defer newFile.Close()
	newFile.Write([]byte(string(newContent)))
}

func replace(content []rune) []rune {
	scanner := bufio.NewScanner(strings.NewReader(string(content)))

	newContent := make([]rune, 0)
	lineLenCodeBlock := 0
	inCodeBlock := false
	inMathBlock := false
	for scanner.Scan() {
		newLine := make([]rune, 0)
		line := []rune(scanner.Text())

		// コードブロック内
		if inCodeBlock {
			if advance := scanRepeat(line, "`"); advance >= lineLenCodeBlock {
				inCodeBlock = false
			}
			newContent = append(newContent, line...)
			newContent = append(newContent, '\n')
			continue
		}

		// コードブロックに入る
		if advance := scanRepeat(line, "`"); advance >= 3 {
			inCodeBlock = true
			lineLenCodeBlock = advance
			newContent = append(newContent, line...)
			newContent = append(newContent, '\n')
			continue
		}

		// math ブロック内
		if inMathBlock {
			if advance := scanRepeat(line, "$"); advance == 2 {
				inMathBlock = false
			}
			newContent = append(newContent, line...)
			newContent = append(newContent, '\n')
			continue
		}

		// math ブロックに入る
		if advance := scanRepeat(line, "$"); advance == 2 {
			inMathBlock = true
			newContent = append(newContent, line...)
			newContent = append(newContent, '\n')
			continue
		}

		id := 0
		for id < len(line) {

			// inline ブロック
			if advance := scanInlineBlock(line[id:]); advance > 0 {
				newLine = append(newLine, line[id:id+advance]...)
				id += advance
				continue
			}

			// inline math
			if advance := scanInlineMath(line[id:]); advance > 0 {
				newLine = append(newLine, line[id:id+advance]...)
				id += advance
				continue
			}

			// エスケープ
			if advance, escaped := scanEscaped(line[id:]); advance > 0 {
				newLine = append(newLine, escaped...)
				id += advance
				continue
			}

			if advance, _, _ := scanExternalLink(line[id:]); advance > 0 {
				newLine = append(newLine, line[id:id+advance]...)
				id += advance
				continue
			}

			if advance := scanRepeat(line[id:], "#"); advance > 1 {
				newLine = append(newLine, line[id:id+advance]...)
				id += advance
				continue
			}

			if advance, _ := scanTag(line[id:]); advance > 0 {
				id += advance
				continue
			}

			// internl link [[]]
			if advance, content := scanInternalLink(line[id:]); advance > 0 {
				if content == "" { // [[ ]] はスキップ
					id += advance
					continue
				}
				link := genHugoLink(content)
				newLine = append(newLine, []rune(link)...)
				id += advance
				continue
			}

			// 普通の文字
			newLine = append(newLine, line[id])
			id++
		}
		newContent = append(newContent, newLine...)
		newContent = append(newContent, '\n')
	}
	return newContent
}
