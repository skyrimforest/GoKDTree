/*
@File    :   Lab5.go
@Time    :   2023/11/06 16:25:13
@Author  :   Skyrim
@Version :   1.0
@Site    :   https://github.com/skyrimforest
@Desc    :   None
*/

package main

import (
	"KDTree"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

var train_data [][]float64
var test_data [][]float64
var test_label []int

const (
	TRAIN_DATA_RATIO float64 = 0.7
	LABEL_NORMAL     int     = 0
	LABEL_SMURF      int     = 1
)

var discrete_dims map[int]int = map[int]int{
	1: 1,
	2: 1,
	3: 1,
	// 41: 1,
}
var discrete_value_map map[int]map[string]int = map[int]map[string]int{
	1: map[string]int{},
	2: map[string]int{},
	3: map[string]int{},
	// 41: map[string]int{},
}

func normalize(data []float64, maxes []float64, mins []float64) {
	for i := 0; i < len(data); i++ {
		// if discrete_dims[i] != 0 {
		// 	continue
		// }
		if maxes[i] != mins[i] {
			data[i] = (data[i] - mins[i]) / (maxes[i] - mins[i])
		}
	}
}

func showResult() {
	normal_points := plotter.XYs{}
	smurf_points := plotter.XYs{}
	for idx, it := range thresholds {
		normal_points = append(normal_points, plotter.XY{
			X: it,
			Y: normal_accus[idx],
		})
		smurf_points = append(smurf_points, plotter.XY{
			X: it,
			Y: smurf_accus[idx],
		})
	}
	p := plot.New()
	p.Title.Text = "How Normal/Smurf's accuracy fluctuate with threshold"
	p.X.Label.Text = "Threshold"
	p.Y.Label.Text = "Accuracy"
	p.Y.Min, p.Y.Max = 50, 100
	p.X.Min, p.X.Max = 0.4, 3.2

	if err := plotutil.AddLines(p,
		"Normal Accuracy", normal_points,
		"Smurf Accuracy", smurf_points,
	); err != nil {
		log.Fatal(err)
	}
	if err := p.Save(5*vg.Inch, 5*vg.Inch, "target.png"); err != nil {
		panic(err)
	}
}

var normals = [][]float64{}
var smurfs = [][]float64{}

// 读取数据,每个line都是一个str数组
func readData() {
	csvData, err := os.Open("data.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer csvData.Close()
	csvReader := csv.NewReader(csvData)

	maxes := make([]float64, 42)
	for i := 0; i < len(maxes); i++ {
		maxes[i] = -float64(1e7)
	}
	mins := make([]float64, 42)
	for i := 0; i < len(mins); i++ {
		mins[i] = float64(1e7)
	}

	for {
		line, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		if line[41] == "normal." || line[41] == "smurf." {
			//从[]string转换为[]float64
			//不要标签了!
			converted := make([]float64, len(line)-1)
			for i := 0; i < len(line)-1; i++ {
				if discrete_dims[i] != 0 {
					if discrete_value_map[i][line[i]] == 0 {
						discrete_value_map[i][line[i]] = len(discrete_value_map[i]) + 1
					}
					converted[i] = float64(discrete_value_map[i][line[i]])
				} else {
					converted[i], err = strconv.ParseFloat(line[i], 64)
					if err != nil {
						log.Fatal(err)
					}
				}
				maxes[i] = math.Max(maxes[i], converted[i])
				mins[i] = math.Min(mins[i], converted[i])
			}
			if line[41] == "normal." {
				// fmt.Println("this is normal:", converted)
				normals = append(normals, converted)
			} else {
				// fmt.Println("this is smurf:", converted)
				smurfs = append(smurfs, converted)
			}
		}
	}

	for _, it := range normals {
		normalize(it, maxes, mins)
		// fmt.Println(it)
		if float64(len(train_data)) < TRAIN_DATA_RATIO*float64(len(normals)) && float64(rand.Intn(100))/100 < TRAIN_DATA_RATIO {
			train_data = append(train_data, it)
		} else {
			test_data = append(test_data, it)
			test_label = append(test_label, LABEL_NORMAL)
		}
	}
	for _, it := range smurfs {
		normalize(it, maxes, mins)
		// fmt.Println(it)
		test_data = append(test_data, it)
		test_label = append(test_label, LABEL_SMURF)
	}
}

var normal_accus []float64
var normal_errors []float64
var smurf_accus []float64
var smurf_errors []float64
var thresholds []float64

func initArray() {
	for i := 0.5; i <= 3; i += 0.1 {
		thresholds = append(thresholds, i)
	}
}

func main() {
	kdtree, err := KDTree.NewKDTree(41, discrete_dims)
	if err != nil {
		log.Fatal(err)
	}
	readData()
	initArray()
	for _, it := range train_data {
		kdtree.Insert(it)
	}

	for _, threshold := range thresholds {
		normal_counts := []int{0.0, 0.0}
		smurf_counts := []int{0.0, 0.0}
		for i, data := range test_data {
			res := 0
			cur_max, _ := kdtree.Get_nearest(data)
			if cur_max > threshold {
				res = LABEL_SMURF
			} else {
				res = LABEL_NORMAL
			}
			if test_label[i] == LABEL_NORMAL {
				normal_counts[0] += 1
				if res == LABEL_NORMAL {
					normal_counts[1] += 1
				}
			} else {
				smurf_counts[0] += 1
				if res == LABEL_SMURF {
					smurf_counts[1] += 1
				}
			}
		}

		temp1 := 0
		temp2 := 0
		for _, it := range test_label {
			if it == LABEL_NORMAL {
				temp1++
			} else {
				temp2++
			}
		}
		normal_accu := 100 * float64(normal_counts[1]) / float64(normal_counts[0])
		normal_error := 100 * float64(normal_counts[0]-normal_counts[1]) / float64(normal_counts[0])
		smurf_accu := 100 * float64(smurf_counts[1]) / float64(smurf_counts[0])
		smurf_error := 100 * float64(smurf_counts[0]-smurf_counts[1]) / float64(smurf_counts[0])

		normal_accus = append(normal_accus, normal_accu)
		normal_errors = append(normal_errors, normal_error)
		smurf_accus = append(smurf_accus, smurf_accu)
		smurf_errors = append(smurf_errors, smurf_error)

		fmt.Printf("Distance threshold=%.2f\n", threshold)
		fmt.Printf("Train data: %v NORMAL data\n", len(train_data))
		fmt.Printf("Test results\n")

		fmt.Printf("NORMAL correct/total=%v/%v, accu=%.2f, error=%.2f\n", normal_counts[1], normal_counts[0], normal_accu, normal_error)

		fmt.Printf("SMURF correct/total=%v/%v, accu=%.2f, error=%.2f\n", smurf_counts[1], smurf_counts[0], smurf_accu, smurf_error)

		fmt.Println()
	}
	showResult()
}
