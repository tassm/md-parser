package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"regexp"
)

type TokenType int

const (
	Text TokenType = iota
	Header1
	Header2
	Header3
	Header4
	Header5
	Header6
	Italic
	Bold
	Strike
	BulletList
	NumberedList
	CodeBlock
	CodeInline
	BlockQuote
)

type Token struct {
	Type  TokenType
	Value string
}

type Parser struct {
	tokens  []Token
	current int
}

func NewParser(markdown []byte) *Parser {
	p := &Parser{}
	p.tokenize(markdown)
	return p
}

func (p *Parser) tokenize(markdown []byte) {
	// Tokenize the markdown string into tokens
	// Here, you would implement your tokenization logic
	// For simplicity, we'll just split by space and new lines
	// and identify headers, italics, bold, lists, and code blocks
	// You might want to expand this tokenizer for more complex Markdown
	// features.
	// This is a very basic tokenizer and may not handle all edge cases.
	// It's recommended to use a proper lexer/parser library for real-world scenarios.

	// Split the markdown into lines
	lines := bytes.Split(markdown, []byte("\n"))

	// compile regexes
	lir := regexp.MustCompile(`^[0-9]+\.\ `)
	blockScan := false
	blockScanIndent := false

	var tokenType TokenType
	var value string

	// Iterate over lines to tokenize
	for i, line := range lines {
		// handle multiline block scanning with indent
		if bytes.HasPrefix(line, []byte("    ")) {
			fmt.Printf("adding line %d to code block %s\n", i, string(line))
			blockScanIndent = true
			tokenType = CodeBlock
			value = value + string(line) + "\n"
		} else if bytes.HasPrefix(line, []byte("```")) { // block scan with fences
			blockScan = !blockScan
			tokenType = CodeBlock
			fmt.Printf("code block found on line: %d\n", i)
			if !blockScan {
				fmt.Printf("block scan ended full value is %s", value)
				p.tokens = append(p.tokens, Token{Type: tokenType, Value: value})
				continue
			}
		} else if blockScan {
			fmt.Printf("adding line %d to code block %s\n", i, string(line))
			value = value + string(line) + "\n"
		} else if blockScanIndent && !bytes.HasPrefix(line, []byte("    ")) {
			p.tokens = append(p.tokens, Token{Type: tokenType, Value: value})
			blockScanIndent = false
		}
		if !blockScan && !blockScanIndent {
			tokenType = -1
			value = ""
			if bytes.HasPrefix(line, []byte("######")) { // headings 6-1
				tokenType = Header6
				value = string(bytes.TrimLeft(line, "###### "))
			} else if bytes.HasPrefix(line, []byte("#####")) {
				tokenType = Header5
				value = string(bytes.TrimLeft(line, "##### "))
			} else if bytes.HasPrefix(line, []byte("#### ")) {
				tokenType = Header4
				value = string(bytes.TrimLeft(line, "#### "))
			} else if bytes.HasPrefix(line, []byte("### ")) {
				tokenType = Header3
				value = string(bytes.TrimLeft(line, "### "))
			} else if bytes.HasPrefix(line, []byte("## ")) {
				tokenType = Header2
				value = string(bytes.TrimLeft(line, "## "))
			} else if bytes.HasPrefix(line, []byte("# ")) {
				tokenType = Header1
				value = string(bytes.TrimLeft(line, "# "))
			} else if bytes.HasPrefix(line, []byte("- ")) { // list types
				tokenType = BulletList
				value = string(bytes.TrimLeft(line, "- "))
			} else if lir.Match(line) {
				tokenType = NumberedList
				value = string(lir.ReplaceAll(line, nil))
			} else if bytes.HasPrefix(line, []byte("> ")) {
				tokenType = BlockQuote
				value = string(bytes.TrimLeft(line, "> "))
			} else {
				// Handle bold and italics and anything where it can appear inline...
				for {
					if bytes.Contains(line, []byte("**")) || bytes.Contains(line, []byte("__")) { //bold
						startIdx := -1
						if idx := bytes.Index(line, []byte("**")); idx != -1 {
							startIdx = idx
						}
						if idx := bytes.Index(line, []byte("__")); idx != -1 && (startIdx == -1 || idx < startIdx) {
							startIdx = idx
						}

						endIdx := bytes.Index(line[startIdx+2:], line[startIdx:startIdx+2]) // same delimiter
						if endIdx == -1 {
							break
						}
						endIdx += startIdx + 2
						p.tokens = append(p.tokens, Token{Bold, string(line[startIdx+2 : endIdx])})
						line = append(line[:startIdx], line[endIdx+2:]...)
					} else if bytes.Contains(line, []byte("*")) || bytes.Contains(line, []byte("_")) { // italic
						startIdx := -1
						if idx := bytes.Index(line, []byte("*")); idx != -1 {
							startIdx = idx
						}
						if idx := bytes.Index(line, []byte("_")); idx != -1 && (startIdx == -1 || idx < startIdx) {
							startIdx = idx
						}

						endIdx := bytes.Index(line[startIdx+1:], line[startIdx:startIdx+1]) // same delimiter
						if endIdx == -1 {
							break
						}
						endIdx += startIdx + 1
						p.tokens = append(p.tokens, Token{Italic, string(line[startIdx+1 : endIdx])})
						line = append(line[:startIdx], line[endIdx+1:]...)
					} else if bytes.Contains(line, []byte("~~")) { // strike
						startIdx := -1
						if idx := bytes.Index(line, []byte("~~")); idx != -1 && (startIdx == -1 || idx < startIdx) {
							startIdx = idx
						}

						endIdx := bytes.Index(line[startIdx+2:], line[startIdx:startIdx+2]) // same delimiter
						if endIdx == -1 {
							break
						}
						endIdx += startIdx + 2
						p.tokens = append(p.tokens, Token{Strike, string(line[startIdx+2 : endIdx])})
						line = append(line[:startIdx], line[endIdx+2:]...)
					} else if bytes.Contains(line, []byte("`")) { // inline code
						startIdx := -1
						if idx := bytes.Index(line, []byte("`")); idx != -1 {
							startIdx = idx
						}

						endIdx := bytes.Index(line[startIdx+1:], line[startIdx:startIdx+1]) // same delimiter
						if endIdx == -1 {
							break
						}
						endIdx += startIdx + 1
						p.tokens = append(p.tokens, Token{CodeInline, string(line[startIdx+1 : endIdx])})
						line = append(line[:startIdx], line[endIdx+1:]...)
					} else {
						break
					}
				}
				if len(line) > 0 {
					p.tokens = append(p.tokens, Token{Text, string(line)})
				}
			}
			// Append token to the list of tokens
			p.tokens = append(p.tokens, Token{Type: tokenType, Value: value})
		}
	}
}

func (p *Parser) parse() string {
	var result string
	var ulFlag bool
	var olFlag bool
	for p.current < len(p.tokens) {
		token := p.tokens[p.current]
		// end lists
		if ulFlag && token.Type != BulletList {
			ulFlag = false
			result += "</ul>"
		} else if olFlag && token.Type != NumberedList {
			olFlag = false
			result += "</ol>"
		}
		switch token.Type {
		case Text:
			result += token.Value + "\n"
		case Header6:
			result += "<h6>" + token.Value + "</h6>\n"
		case Header5:
			result += "<h5>" + token.Value + "</h5>\n"
		case Header4:
			result += "<h4>" + token.Value + "</h4>\n"
		case Header3:
			result += "<h3>" + token.Value + "</h3>\n"
		case Header2:
			result += "<h2>" + token.Value + "</h2>\n"
		case Header1:
			result += "<h1>" + token.Value + "</h1>\n"
		case Italic:
			result += "<i>" + token.Value + "</i>\n" // TODO: find a way to only newline at the end?
		case Bold:
			result += "<b>" + token.Value + "</b>\n"
		case Strike:
			result += "<s>" + token.Value + "</s>\n"
		case CodeInline:
			result += "<code>" + token.Value + "</code>\n"
		case BulletList:
			if !ulFlag {
				ulFlag = true
				result += "<ul>"
			}
			result += "<li>" + token.Value + "</li>\n"
		case NumberedList:
			if !olFlag {
				olFlag = true
				result += "<ol>"
			}
			result += "<li>" + token.Value + "</li>\n"
		case CodeBlock:
			result += "<pre><code>" + token.Value + "</code></pre>\n"
		case BlockQuote:
			result += "<blockquote>" + token.Value + "</blockquote>\n"
		}
		p.current++
	}
	return result
}

func main() {
	bytes, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		parser := NewParser(bytes)
		html := parser.parse()
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, "%s", html)
	})
	fmt.Println("Server is listening on port 8080...")
	http.ListenAndServe(":8080", nil)
}
