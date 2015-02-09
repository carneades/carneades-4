// Import Dung AFs represented using the Trivial Graph Format
// See <https://en.wikipedia.org/wiki/Trivial_Graph_Format>
package tgf

import (
	"../../engine/dung"
	"bufio"
	"fmt"
	"io"
)

func Import(inFile io.Reader) (af dung.AF, err error) {
	reader := bufio.NewReader(inFile)
	args := make([]string, 0, 50)
	atks := make(map[string][]string, 50)
	nodeList := true // false if reading the list of edges has begun
	var line, token1, token2 string
	var n int
	eof := false
	for !eof {
		token1, token2 = "", ""
		line, err = reader.ReadString('\n')
		if err == io.EOF {
			err = nil // io.EOF isn't really an error
			eof = true
		} else if err != nil {
			return af, err // finish immediately for real errors
		}
		n, _ = fmt.Sscan(line, &token1, &token2)
		if nodeList && n >= 1 {
			if token1 == "#" {
				nodeList = false // start of edges list
				continue
			}
			args = append(args, token1)
		} else if !nodeList && n >= 2 { // edges list
			atks[token2] = append(atks[token2], token1)
		} else {
			continue // skip empty and invalid lines
		}
	}
	return dung.NewAF(args, atks), err
}
