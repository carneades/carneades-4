// Import Dung AFs represented using the Trivial Graph Format
// See <https://en.wikipedia.org/wiki/Trivial_Graph_Format>
package tgf

import (
	"bufio"
	"fmt"
	"github.com/carneades/carneades-go/internal/engine/dung"
	"io"
)

func Import(inFile io.Reader) (af dung.AF, err error) {
	reader := bufio.NewReader(inFile)
	args := make([]dung.Arg, 0, 50)
	atks := make(map[dung.Arg][]dung.Arg, 50)
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
			args = append(args, dung.Arg(token1))
		} else if !nodeList && n >= 2 { // edges list
			atks[dung.Arg(token2)] = append(atks[dung.Arg(token2)], dung.Arg(token1))
		} else {
			continue // skip empty and invalid lines
		}
	}
	return dung.NewAF(args, atks), err
}
