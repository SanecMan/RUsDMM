package dmmap

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"

	"github.com/SpaiR/strongdmm/internal/app/byond/dmvars"
)

/**
A slightly modified algorithm made on rust:
https://raw.githubusercontent.com/SpaceManiac/SpacemanDMM/5e51421/src/tools/dmm/read.rs
Unlike the original one, doesn't care about storing keys as base52 number and uses a simple string for that.
*/
func parseData(file *os.File) (*Data, error) {
	r := bufio.NewReader(file)

	var (
		dmmData = Data{
			Filepath:   file.Name(),
			Dictionary: make(map[Key][]Prefab),
			Grid:       make(map[Coord]Key),
		}

		firstRune = true

		inCommentLine  bool
		commentTrigger bool
		inQuoteBlock   bool
		inKeyBlock     bool
		inDataBlock    bool
		inVarEditBlock bool
		afterDataBlock bool
		escaping       bool
		skipWhitespace bool

		currData   []Prefab
		currPrefab = Prefab{Vars: &dmvars.Variables{}}
		currVar    = make([]rune, 0)
		currDatum  = make([]rune, 0)

		currKey       []rune
		currKeyLength = 0
	)

	for {
		if c, _, err := r.ReadRune(); err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		} else {
			if c == '\n' || c == '\r' {
				inCommentLine = false
				commentTrigger = false
				continue
			} else if inCommentLine {
				continue
			} else if c == '\t' {
				continue
			}

			if c == '/' && !inQuoteBlock {
				if commentTrigger {
					inCommentLine = true
					// If the first rune and it's a comment, then we make an assumption that it's a TGM format.
					if firstRune {
						dmmData.IsTgm = true
						firstRune = false
					}
					continue
				} else {
					commentTrigger = true
				}
			} else {
				commentTrigger = false
			}

			if inDataBlock {
				if inVarEditBlock {
					if inQuoteBlock {
						if c == '\\' {
							currDatum = append(currDatum, c)
							escaping = true
						} else if escaping {
							currDatum = append(currDatum, c)
							escaping = false
						} else if c == '"' {
							currDatum = append(currDatum, c)
							inQuoteBlock = false
						} else {
							currDatum = append(currDatum, c)
						}
					} else {
						if skipWhitespace && c == ' ' {
							skipWhitespace = false
							continue
						}
						skipWhitespace = false

						if c == '"' {
							currDatum = append(currDatum, c)
							inQuoteBlock = true
						} else if c == '=' && len(currVar) == 0 {
							currVar = make([]rune, len(currDatum))
							copy(currVar, currDatum)
							currDatum = currDatum[:0]
							length := len(currVar)
							for length > 0 && currVar[length-1] == ' ' {
								length -= 1
							}
							currVar = currVar[:length]
							skipWhitespace = true
						} else if c == ';' {
							value := string(currDatum)
							currPrefab.Vars.Put(string(currVar), &value)
							currVar = currVar[:0]
							currDatum = currDatum[:0]
							skipWhitespace = true
						} else if c == '}' {
							if len(currVar) > 0 {
								value := string(currDatum)
								currPrefab.Vars.Put(string(currVar), &value)
								currVar = currVar[:0]
								currDatum = currDatum[:0]
							}
							inVarEditBlock = false
						} else {
							currDatum = append(currDatum, c)
						}
					}
				} else if c == '{' {
					currPrefab.Path = string(currDatum)
					currDatum = currDatum[:0]
					inVarEditBlock = true
				} else if c == ',' {
					if len(currPrefab.Path) == 0 && len(currDatum) > 0 {
						currPrefab.Path = string(currDatum)
						currDatum = currDatum[:0]
					}
					currData = append(currData, currPrefab)
					currPrefab = Prefab{Vars: &dmvars.Variables{}}
				} else if c == ')' {
					if len(currPrefab.Path) == 0 && len(currDatum) > 0 {
						currPrefab.Path = string(currDatum)
						currDatum = currDatum[:0]
					}
					currData = append(currData, currPrefab)
					currPrefab = Prefab{Vars: &dmvars.Variables{}}
					key := Key(currKey)
					currKey = currKey[:0]
					data := make([]Prefab, len(currData))
					copy(data, currData)
					currData = currData[:0]
					currKeyLength = 0
					dmmData.Dictionary[key] = data
					inDataBlock = false
					afterDataBlock = true
				} else {
					currDatum = append(currDatum, c)
				}
			} else if inKeyBlock {
				if c == '"' {
					inKeyBlock = false
					dmmData.KeyLength = currKeyLength
				} else {
					currKeyLength += 1
					currKey = append(currKey, c)
				}
			} else if c == '"' {
				inKeyBlock = true
				afterDataBlock = false
			} else if c == '(' {
				if afterDataBlock {
					currKey = currKey[:0]
					currKeyLength = 0
					break
				} else {
					inDataBlock = true
					afterDataBlock = false
				}
			}
		}
	}

	type axis int

	const (
		X axis = iota
		Y
		Z
	)

	var (
		readingAxis = X

		currX, currY, currZ uint16 = 0, 0, 0
		currNum             uint16 = 0
		baseX               uint16 = 0

		inCoordBlock = true
		inMapString  = false
		adjustY      = true
	)

	for {
		if c, _, err := r.ReadRune(); err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		} else {
			if inCoordBlock {
				if c == ',' {
					if readingAxis == X {
						currX = currNum
						currNum = 0
						dmmData.MaxX = int(math.Max(float64(dmmData.MaxX), float64(currX)))
						baseX = currX
						readingAxis = Y
					} else if readingAxis == Y {
						currY = currNum
						currNum = 0
						dmmData.MaxY = int(math.Max(float64(dmmData.MaxY), float64(currY)))
						readingAxis = Z
					} else {
						return nil, fmt.Errorf("incorrect number of axis [%d]", readingAxis)
					}
				} else if c == ')' {
					if readingAxis != Z {
						return nil, fmt.Errorf("incorrect reading axis [%d] (expected %d)", readingAxis, Z)
					}
					currZ = currNum
					currNum = 0
					dmmData.MaxZ = int(math.Max(float64(dmmData.MaxZ), float64(currZ)))
					inCoordBlock = false
					readingAxis = X
				} else {
					x, _ := strconv.ParseInt(string(c), 10, 16)
					currNum = 10*currNum + uint16(x)
				}
			} else if inMapString {
				if c == '"' {
					inMapString = false
					adjustY = true
					currY -= 1
				} else if c == '\r' {
					dmmData.IsWinLineBreak = true // Windows line break for sure.
				} else if c == '\n' {
					if adjustY {
						adjustY = false
					} else {
						currY += 1
					}
					currX = baseX
				} else {
					currKeyLength += 1
					currKey = append(currKey, c)
					if currKeyLength == dmmData.KeyLength {
						currKeyLength = 0
						dmmData.Grid[Coord{X: currX, Y: currY, Z: currZ}] = Key(currKey)
						currKey = currKey[:0]
						dmmData.MaxX = int(math.Max(float64(dmmData.MaxX), float64(currX)))
						currX += 1
					}
				}
			} else if c == '(' {
				inCoordBlock = true
			} else if c == '"' {
				inMapString = true
			}
		}
	}

	dmmData.MaxY = int(math.Max(float64(dmmData.MaxY), float64(currY)))

	// Make Y axis to go from bottom to top
	reversedGrid := make(map[Coord]Key, len(dmmData.Grid))
	for z := 1; z <= dmmData.MaxZ; z++ {
		for y := 1; y <= dmmData.MaxY; y++ {
			for x := 1; x <= dmmData.MaxX; x++ {
				reversedGrid[Coord{
					X: uint16(x),
					Y: uint16(dmmData.MaxY + 1 - y),
					Z: uint16(z),
				}] = dmmData.Grid[Coord{
					X: uint16(x),
					Y: uint16(y),
					Z: uint16(z),
				}]
			}
		}
	}
	dmmData.Grid = reversedGrid

	return &dmmData, nil
}
