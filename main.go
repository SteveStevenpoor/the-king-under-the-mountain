package main

import (
	"bufio"
	"fmt"
	"image/color"
	"math"
	"os"
	"sort"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

var p *plot.Plot = plot.New()

type House struct {
	id        int
	houseType string
	X         int
	Y         int
}

type District struct {
	id          int
	houseCount  int
	tavernCount int
	houses      []House
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: executable input.txt output.txt")
		os.Exit(1)
	}

	p.Title.Text = "The Cave map"

	inFile, err := os.Open(os.Args[1])
	handleError(err)

	outFile, err := os.Create(os.Args[2])
	handleError(err)

	in := bufio.NewReader(inFile)
	out := bufio.NewWriter(outFile)

	dstrs := parseInput(in)

	fmt.Fprint(out, greedyMinEvil(dstrs))

	err = p.Save(5*vg.Inch, 5*vg.Inch, "points.png")
	handleError(err)

	defer out.Flush()
}

func greedyMinEvil(dstrs []District) int {
	var optEvil, duration int

	sort.Slice(dstrs, func(i, j int) bool {
		return dstrs[i].GiveWeight() > dstrs[j].GiveWeight()
	})

	for r := range dstrs {
		duration += dstrs[r].houseCount
		optEvil += dstrs[r].tavernCount * duration
	}

	return optEvil
}

func (d District) GiveWeight() float64 {
	return float64(d.tavernCount) / float64(d.houseCount)
}

func parseInput(in *bufio.Reader) []District {
	var distrNum, streetNum, tavernNum int

	fmt.Fscan(in, &distrNum, &streetNum, &tavernNum)

	m := make(map[string]map[int]House)

	houseCount := parseStreets(streetNum, m, in)

	parseTaverns(tavernNum, m, in)

	deleteCrossings(m, &houseCount)

	dstrs := formDistricts(distrNum, m, houseCount)

	return dstrs
}

func deleteCrossings(m map[string]map[int]House, houseCount *int) {
	for str, v := range m {
		for otherStr, u := range m {
			if str != otherStr {
				deleteRepHouses(v, u, houseCount)
			}
		}
	}
}

func deleteRepHouses(strA, strB map[int]House, houseCount *int) {
	for keyA, v := range strA {
		for keyB, u := range strB {
			if v.X == u.X && v.Y == u.Y {
				if v.houseType == "tavern" {
					delete(strB, keyB)
				} else {
					delete(strA, keyA)
				}
				(*houseCount)--
			}
		}
	}
}

func formDistricts(distrNum int, m map[string]map[int]House, houseCount int) []District {

	distrs := make([]District, 0, houseCount)

	for _, v := range m {
		for _, u := range v {
			dstr := District{
				houseCount:  1,
				tavernCount: 0,
				houses:      make([]House, 0, houseCount),
			}
			dstr.houses = append(dstr.houses, u)
			if u.houseType == "tavern" {
				dstr.tavernCount++
			}
			distrs = append(distrs, dstr)
		}
	}

	distrs = clusterrize(distrs, distrNum)

	plotDistricts(distrs)

	return distrs
}

func plotDistricts(distrs []District) {
	for r := range distrs {
		var left, right, up, down int = 1e9, -1, -1, 1e9

		for k := range distrs[r].houses {
			if distrs[r].houses[k].X < left {
				left = distrs[r].houses[k].X
			}
			if distrs[r].houses[k].X > right {
				right = distrs[r].houses[k].X
			}
			if distrs[r].houses[k].Y < down {
				down = distrs[r].houses[k].Y
			}
			if distrs[r].houses[k].Y > up {
				up = distrs[r].houses[k].Y
			}
		}

		plotRectangle(left, right, up, down)
	}
}

func plotRectangle(left, right, up, down int) {
	eps := 0.5

	pts := make(plotter.XYs, 5)
	pts[0].X, pts[0].Y = float64(left)-eps, float64(down)-eps
	pts[1].X, pts[1].Y = float64(left)-eps, float64(up)+eps
	pts[2].X, pts[2].Y = float64(right)+eps, float64(up)+eps
	pts[3].X, pts[3].Y = float64(right)+eps, float64(down)-eps
	pts[4] = pts[0]

	l, err := plotter.NewLine(pts)
	l.Color = color.RGBA{R: 169, G: 169, B: 169, A: 255}
	p.Add(l)
	handleError(err)
}

func clusterrize(distrs []District, distrNum int) []District {
	dstrCount := len(distrs)
	currentDistrNum := dstrCount
	distMat, minDistArr, minId := createMatrices(distrs)

	for currentDistrNum != distrNum {
		idA, idB := findDistsToMerge(minDistArr, minId, dstrCount)

		updateMatrices(distMat, minDistArr, minId, idA, idB)

		mergeDistricts(distrs, idA, idB)

		currentDistrNum--
	}

	j := 0
	for i := 0; i < dstrCount; i++ {
		if distrs[i].id != -1 {
			distrs[i], distrs[j] = distrs[j], distrs[i]
			j++
		}
		if j == distrNum {
			break
		}
	}
	return distrs[:distrNum]
}

func mergeDistricts(distrs []District, idA, idB int) {
	distrs[idB].id = -1
	distrs[idA] = District{
		id:          idA,
		houseCount:  distrs[idA].houseCount + distrs[idB].houseCount,
		tavernCount: distrs[idA].tavernCount + distrs[idB].tavernCount,
		houses:      append(distrs[idA].houses, distrs[idB].houses...),
	}
}

func findDistsToMerge(minDistArr []float64, minId []int, len int) (int, int) {
	var min float64 = 1e9
	var idA, idB int

	for i := 0; i < len; i++ {
		if minDistArr[i] < min {
			min = minDistArr[i]
			idA = i
			idB = minId[i]
		}
	}

	return idA, idB
}

func createMatrices(dstrs []District) ([][]float64, []float64, []int) {
	dstCount := len(dstrs)

	distMat := make([][]float64, dstCount)
	minArr := make([]float64, dstCount)
	minId := make([]int, dstCount)

	for i := 0; i < dstCount; i++ {
		min := 1e9
		dstrs[i].id = i
		distMat[i] = make([]float64, dstCount)
		for j := 0; j < dstCount; j++ {
			if j == i {
				distMat[i][j] = 1e9
				continue
			}
			v := dstrs[i].houses[0]
			u := dstrs[j].houses[0]
			distMat[i][j] = math.Sqrt(math.Pow(float64(v.X-u.X), 2) + math.Pow(float64(v.Y-u.Y), 2))
			if distMat[i][j] < min {
				min = distMat[i][j]
				minArr[i] = distMat[i][j]
				minId[i] = j
			}
		}
	}

	return distMat, minArr, minId
}

func updateMatrices(m [][]float64, minArr []float64, minId []int, idA, idB int) {
	minArr[idA] = 1e9
	minArr[idB] = 1e9

	for i := range m {
		if i != idA && i != idB {
			m[idA][i] = math.Min(m[idA][i], m[idB][i])
			m[i][idA] = m[idA][i]
			if m[idA][i] < minArr[idA] {
				minArr[idA] = m[idA][i]
				minId[idA] = i
			}
		}
		m[i][idB] = 1e9
		m[idB][i] = 1e9
	}

	for i := range minId {
		if minId[i] == idB {
			minId[i] = idA
		}
	}
}

func parseTaverns(tavernNum int, m map[string]map[int]House, in *bufio.Reader) {
	pts := make(plotter.XYs, tavernNum)

	var tavernStreet [2]string
	var tavernHouse int

	for i := 0; i < tavernNum; i++ {
		fmt.Fscan(in, &tavernStreet[0], &tavernStreet[1], &tavernHouse)
		address := tavernStreet[0] + tavernStreet[1]
		address = address[:len(address)-1]
		if entry, ok := m[address][tavernHouse]; ok {
			entry.houseType = "tavern"
			pts[i].X = float64(entry.X)
			pts[i].Y = float64(entry.Y)
			m[address][tavernHouse] = entry
		}
	}

	plotTavern(pts)
}

func parseStreets(streetNum int, m map[string]map[int]House, in *bufio.Reader) int {
	id := 0

	for i := 0; i < streetNum; i++ {
		var prefix, streetType string
		var startX, startY, houseNum, distOnStreet int

		fmt.Fscan(in, &prefix, &streetType, &startX, &startY, &houseNum, &distOnStreet)

		m[prefix+streetType] = make(map[int]House)
		pts := make(plotter.XYs, houseNum)

		for j := 0; j < houseNum; j++ {
			var isRoad, isAvenue int
			if streetType == "Road" {
				isRoad = 1
			} else {
				isAvenue = 1
			}

			h := House{
				id:        id,
				houseType: "regular",
				X:         startX + distOnStreet*j*isAvenue,
				Y:         startY + distOnStreet*j*isRoad,
			}

			id++
			m[prefix+streetType][j] = h
			pts[j].X = float64(h.X)
			pts[j].Y = float64(h.Y)
		}

		plotStreet(pts)
	}

	return id
}

func plotTavern(pts plotter.XYs) {
	s, err := plotter.NewScatter(pts)
	s.Shape = draw.CircleGlyph{}
	s.Radius = 3
	s.Color = color.RGBA{G: 255, A: 255}
	p.Add(s)
	handleError(err)
}

func plotStreet(pts plotter.XYs) {
	l, s, err := plotter.NewLinePoints(pts)
	s.Shape = draw.PyramidGlyph{}
	s.Color = color.RGBA{B: 255, A: 255}
	l.Color = color.RGBA{R: 255, A: 255}
	p.Add(s, l)
	handleError(err)
}

func handleError(e error) {
	if e != nil {
		panic(e)
	}
}
