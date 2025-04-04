package main

import (
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"strings"
	"sync"
	"time"
)

type Tile struct {
	isBomb      bool
	isUncovered bool
	isFlagged   bool
	nearbyBombs int
	x           int
	y           int
	z           int
}

var RandomGenerator = rand.New(rand.NewSource(1))

func main() {
	go func() {
		println("pprof en écoute sur http://localhost:6060")
		http.ListenAndServe("localhost:6060", nil)
	}()

	size := 25
	bombCount := 600
	grid := generateGrid(size, bombCount)
	//displayGrid(grid, true)
	grid = solve(grid, bombCount)
	//displayGrid(grid, false)

	time.Sleep(30 * time.Second)
}

func generateGrid(size int, bombCount int) [][][]Tile {
	if size*size*size < bombCount {
		return nil
	}

	grid := make([][][]Tile, size)

	for x := range size {
		grid[x] = make([][]Tile, size)
		for y := range size {
			grid[x][y] = make([]Tile, size)
			for z := range size {
				tile := Tile{}
				tile.x = x
				tile.y = y
				tile.z = z
				grid[x][y][z] = tile
			}
		}
	}

	bombsPlaced := 0
	for bombsPlaced < bombCount {
		x := RandomGenerator.Intn(size)
		y := RandomGenerator.Intn(size)
		z := RandomGenerator.Intn(size)
		if !grid[x][y][z].isBomb {
			grid[x][y][z].isBomb = true
			forEachNeighbour(grid, x, y, z, func(tile Tile) {
				grid[tile.x][tile.y][tile.z].nearbyBombs++
			})
			bombsPlaced++
		}
	}

	return grid
}

func solve(grid [][][]Tile, bombCount int) [][][]Tile {
	gridSize := len(grid)
	tilesCount := gridSize * gridSize * gridSize
	emptyTilesCount := tilesCount - bombCount

	hasFailed := false
	x := RandomGenerator.Intn(gridSize - 1)
	y := RandomGenerator.Intn(gridSize - 1)
	z := RandomGenerator.Intn(gridSize - 1)

	for {
		//println("Played :", x, y)
		if grid[x][y][z].isBomb {
			hasFailed = true
			break
		}

		grid = uncoverTile(grid, x, y, z)
		grid = flagTiles(grid)
		//displayGrid(grid, false)
		var uncoveredTilesCount = 0
		for _, slice := range grid {
			for _, row := range slice {
				for _, tile := range row {
					if tile.isUncovered {
						uncoveredTilesCount++
					}
				}
			}
		}
		if uncoveredTilesCount == emptyTilesCount {
			break
		}

		nextTile := getFirstSafeTile(grid)
		if nextTile == nil {
			hasFailed = true
			break
		}
		x = nextTile.x
		y = nextTile.y
		z = nextTile.z
	}

	if hasFailed {
		println("*BOOM*")
	} else {
		println("SUCCESS")
	}

	return grid
}

func displayGrid(grid [][][]Tile, showAll bool) {
	size := len(grid)

	println("\n----------------------------------------------------------\n")
	for _, slice := range grid {
		for _, row := range slice {
			print(strings.Repeat("-", size*4+1))
			print("\n|")
			for _, tile := range row {
				if !showAll && tile.isFlagged {
					print(" ⚑ ")
				} else if !showAll && !tile.isUncovered {
					print("███")
				} else if tile.isBomb {
					print(" X ")
				} else if tile.nearbyBombs != 0 {
					print(" ")
					print(tile.nearbyBombs)
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
	print("\n")
}

func forEachNeighbour(grid [][][]Tile, x, y, z int, action func(tile Tile)) {
	size := len(grid)

	directions := [26][3]int{
		{-1, -1, -1}, {-1, -1, 0}, {-1, -1, 1},
		{-1, 0, -1}, {-1, 0, 0}, {-1, 0, 1},
		{-1, 1, -1}, {-1, 1, 0}, {-1, 1, 1},

		{0, -1, -1}, {0, -1, 0}, {0, -1, 1},
		{0, 0, -1} /*, {0, 0, 0}*/, {0, 0, 1},
		{0, 1, -1}, {0, 1, 0}, {0, 1, 1},

		{1, -1, -1}, {1, -1, 0}, {1, -1, 1},
		{1, 0, -1}, {1, 0, 0}, {1, 0, 1},
		{1, 1, -1}, {1, 1, 0}, {1, 1, 1},
	}

	for _, dir := range directions {
		nx, ny, nz := x+dir[0], y+dir[1], z+dir[2]
		if nx >= 0 && nx < size && ny >= 0 && ny < size && nz >= 0 && nz < size {
			action(grid[nx][ny][nz])
		}
	}
}

func getNeighboursLeft(grid [][][]Tile, tile Tile) []Tile {
	var neighboursLeft []Tile
	forEachNeighbour(grid, tile.x, tile.y, tile.z, func(tile Tile) {
		if !tile.isUncovered {
			neighboursLeft = append(neighboursLeft, tile)
		}
	})
	return neighboursLeft
}

func getNearbyFlaggedBombsCount(grid [][][]Tile, tile Tile) int {
	flaggedBombsCount := 0
	forEachNeighbour(grid, tile.x, tile.y, tile.z, func(tile Tile) {
		if tile.isFlagged {
			flaggedBombsCount++
		}
	})
	return flaggedBombsCount
}

func flagTiles(grid [][][]Tile) [][][]Tile {
	var wg sync.WaitGroup
	for _, slice := range grid {
		for _, row := range slice {
			for _, tile := range row {
				wg.Add(1)
				go func() {
					defer wg.Done()
					if !tile.isUncovered || tile.nearbyBombs == 0 {
						return
					}

					neighboursLeft := getNeighboursLeft(grid, tile)

					if len(neighboursLeft) == tile.nearbyBombs {
						for _, neighbour := range neighboursLeft {
							if !neighbour.isFlagged {
								grid[neighbour.x][neighbour.y][neighbour.z].isFlagged = true
							}
						}
					}
				}()
			}
		}
	}
	wg.Wait()
	return grid
}

func getFirstSafeTile(grid [][][]Tile) *Tile {
	for _, slice := range grid {
		for _, row := range slice {
			for _, tile := range row {
				if !tile.isUncovered || tile.nearbyBombs == 0 {
					continue
				}

				neighboursLeft := getNeighboursLeft(grid, tile)
				nearbyFlaggedBombsCount := getNearbyFlaggedBombsCount(grid, tile)
				if tile.nearbyBombs == nearbyFlaggedBombsCount &&
					len(neighboursLeft) > nearbyFlaggedBombsCount {
					for _, neighbour := range neighboursLeft {
						if !neighbour.isFlagged {
							return &grid[neighbour.x][neighbour.y][neighbour.z]
						}
					}
				}
			}
		}
	}
	return nil
}

func uncoverTile(grid [][][]Tile, x, y, z int) [][][]Tile {
	grid[x][y][z].isUncovered = true
	tile := grid[x][y][z]

	if tile.nearbyBombs > 0 {
		return grid
	}

	forEachNeighbour(grid, x, y, z, func(tile Tile) {
		if !tile.isUncovered {
			grid = uncoverTile(grid, tile.x, tile.y, tile.z)
		}
	})

	return grid
}
