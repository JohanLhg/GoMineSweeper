package main

import (
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"reflect"
	"strings"
	"time"
)

type Tile struct {
	isBomb bool
}

var RandomGenerator = rand.New(rand.NewSource(3))

func main() {

	go func() {
		println("pprof en écoute sur http://localhost:6060")
		http.ListenAndServe("localhost:6060", nil)
	}()
	
	size := 25
	bombCount := 75
	grid := generateGrid(size, bombCount)
	//displayGrid(grid, [][]int{}, [][]int{}, true)
	solve(grid, bombCount)

	time.Sleep(30 * time.Second)
}

func generateGrid(size int, bombCount int) [][]Tile {
	if size*size < bombCount {
		return nil
	}

	matrix := make([][]Tile, size)
	for i := range matrix {
		matrix[i] = make([]Tile, size)
	}

	bombsPlaced := 0
	for bombsPlaced < bombCount {
		i := RandomGenerator.Intn(size)
		j := RandomGenerator.Intn(size)
		if !matrix[i][j].isBomb {
			matrix[i][j].isBomb = true
			bombsPlaced++
		}
	}

	return matrix
}

func forEachNeighbour(grid [][]Tile, x int, y int, action func(x int, y int)) {
	size := len(grid)

	directions := [8][2]int{
		{-1, -1}, {-1, 0}, {-1, 1},
		{0, -1}, {0, 1},
		{1, -1}, {1, 0}, {1, 1},
	}

	for _, dir := range directions {
		nx, ny := x+dir[0], y+dir[1]
		if nx >= 0 && nx < size && ny >= 0 && ny < size {
			action(nx, ny)
		}
	}
}

func contains(slice [][]int, target []int) bool {
	for _, item := range slice {
		if reflect.DeepEqual(item, target) {
			return true
		}
	}
	return false
}

func isUncovered(uncoveredTiles [][]int, x int, y int) bool {
	return contains(uncoveredTiles, []int{x, y})
}

func isFlagged(flaggedTiles [][]int, x int, y int) bool {
	return contains(flaggedTiles, []int{x, y})
}

func getNeighboursLeft(grid [][]Tile, uncoveredTiles [][]int, x int, y int) [][]int {
	var neighboursLeft [][]int
	forEachNeighbour(grid, x, y, func(nx int, ny int) {
		if !isUncovered(uncoveredTiles, nx, ny) {
			neighboursLeft = append(neighboursLeft, []int{nx, ny})
		}
	})
	return neighboursLeft
}

func getNearbyFlaggedBombs(grid [][]Tile, flaggedTiles [][]int, x int, y int) [][]int {
	var flaggedBombs [][]int
	forEachNeighbour(grid, x, y, func(x, y int) {
		if isFlagged(flaggedTiles, x, y) {
			flaggedBombs = append(flaggedBombs, []int{x, y})
		}
	})
	return flaggedBombs
}

func displayGrid(grid [][]Tile, uncoveredTiles [][]int, flaggedTiles [][]int, showAll bool) {
	size := len(grid)

	for x, column := range grid {
		print(strings.Repeat("-", size*4+1))
		print("\n|")
		for y, tile := range column {
			nearbyBombs := countNearbyBombs(grid, x, y)
			if !showAll && isFlagged(flaggedTiles, x, y) {
				print(" ⚑ ")
			} else if !showAll && !isUncovered(uncoveredTiles, x, y) {
				print("███")
			} else if tile.isBomb {
				print(" X ")
			} else if nearbyBombs != 0 {
				print(" ")
				print(nearbyBombs)
				print(" ")
			} else {
				print("   ")
			}
			print("|")
		}
		print("\n")
	}
	print(strings.Repeat("-", size*4+1))
	print("\n")
}

func solve(grid [][]Tile, bombCount int) {
	gridSize := len(grid)
	var uncoveredTiles [][]int
	var flaggedTiles [][]int

	//displayGrid(grid, uncoveredTiles, flaggedTiles, true)
	hasFailed := false
	x := RandomGenerator.Intn(gridSize - 1)
	y := RandomGenerator.Intn(gridSize - 1)

	for {
		//fmt.Println("Played :", x, y)
		if grid[x][y].isBomb {
			hasFailed = true
			break
		}

		uncoveredTiles = uncoverTile(grid, uncoveredTiles, x, y)
		flaggedTiles = flagTiles(grid, uncoveredTiles, flaggedTiles)
		//displayGrid(grid, uncoveredTiles, flaggedTiles, false)
		if bombCount == len(flaggedTiles) && len(uncoveredTiles) == gridSize*gridSize-bombCount {
			break
		}
		nextTile := getFirstSafeTile(grid, uncoveredTiles, flaggedTiles)
		if len(nextTile) < 2 {
			hasFailed = true
			break
		}
		x = nextTile[0]
		y = nextTile[1]
	}

	//displayGrid(grid, uncoveredTiles, flaggedTiles, false)

	if hasFailed {
		println("*BOOM*")
	} else {
		println("SUCCESS")
	}
}

func flagTiles(grid [][]Tile, uncoveredTiles [][]int, flaggedTiles [][]int) [][]int {
	newFlag := true
	for newFlag {
		newFlag = false
		for _, tile := range uncoveredTiles {
			x := tile[0]
			y := tile[1]

			if countNearbyBombs(grid, x, y) == 0 {
				continue
			}

			if len(getNeighboursLeft(grid, uncoveredTiles, x, y)) == countNearbyBombs(grid, x, y) {
				forEachNeighbour(grid, x, y, func(x, y int) {
					if !isUncovered(uncoveredTiles, x, y) && !isFlagged(flaggedTiles, x, y) {
						newFlag = true
						flaggedTiles = append(flaggedTiles, []int{x, y})
					}
				})
			}
		}
	}
	return flaggedTiles
}

func getFirstSafeTile(grid [][]Tile, uncoveredTiles [][]int, flaggedTiles [][]int) []int {
	for _, tile := range uncoveredTiles {
		x := tile[0]
		y := tile[1]

		if countNearbyBombs(grid, x, y) == 0 {
			continue
		}

		if countNearbyBombs(grid, x, y) == len(getNearbyFlaggedBombs(grid, flaggedTiles, x, y)) &&
			len(getNeighboursLeft(grid, uncoveredTiles, x, y)) > len(getNearbyFlaggedBombs(grid, flaggedTiles, x, y)) {
			neighbours := getNeighboursLeft(grid, uncoveredTiles, x, y)
			for _, tile := range neighbours {
				nx := tile[0]
				ny := tile[1]

				if !isFlagged(flaggedTiles, nx, ny) {
					return []int{nx, ny}
				}
			}
		}
	}
	return []int{}
}

func uncoverTile(grid [][]Tile, uncoveredTiles [][]int, x int, y int) [][]int {
	uncoveredTiles = append(uncoveredTiles, []int{x, y})

	if countNearbyBombs(grid, x, y) > 0 {
		return uncoveredTiles
	}

	forEachNeighbour(grid, x, y, func(nx int, ny int) {
		if !isUncovered(uncoveredTiles, nx, ny) && !grid[nx][ny].isBomb {
			uncoveredTiles = uncoverTile(grid, uncoveredTiles, nx, ny)
		}
	})

	return uncoveredTiles
}

func countNearbyBombs(grid [][]Tile, x int, y int) int {
	nearbyBombs := 0

	forEachNeighbour(grid, x, y, func(nx int, ny int) {
		if grid[nx][ny].isBomb {
			nearbyBombs++
		}
	})

	return nearbyBombs
}
