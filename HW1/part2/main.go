package main

import (
	"encoding/csv"
	"fmt"
	"image/color"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"runtime/trace"
	"sort"
	"strconv"
	"sync"
	"time"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

var gorutineCounts = []int{1, 2, 4, 8, 16, 32, 64}
var GOMAXPROCS_vals = []int{1, 2, runtime.NumCPU()}

const (
	CPU_BOUND_WORKLOAD = "CPU-bound"
	IO_BOUND_WORKLOAD  = "IO-bound"
	TABLE_PATH         = "./results/results.csv"
	totalTasks         = 640
)

type Result struct {
	TasksPerGoroutine int
	TotalTime         float64
	Throughput        float64
	AvgLat            float64
	MaxLat            float64
	MinLat            float64
	StdLat            float64
	Goroutines        int
	WorkloadType      string
	GOMAXPROCS        int
}

func cpuwork() float64 {
	sum := 0.0
	for i := 0; i < 1e5; i++ {
		sum += math.Sqrt(float64(i+1)) * math.Log(float64(i+2))
	}
	return sum
}

func iowork() float64 {
	sum := 0.0
	for i := 0; i < 1e2; i++ {
		sum += math.Sqrt(float64(i + 1))
	}
	time.Sleep(time.Duration(3) * time.Millisecond)
	return sum
}

func runBenchmark(numGoroutines int, workloadType string,
	goMacProcs int) Result {
	runtime.GOMAXPROCS(goMacProcs)

	tasksPerGoroutine := totalTasks / numGoroutines
	latencies := make([]float64, totalTasks)
	var mu sync.Mutex
	idx := 0
	var wg sync.WaitGroup
	allStart := time.Now()

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for t := 0; t < tasksPerGoroutine; t++ {
				taskStart := time.Now()
				switch workloadType {
				case CPU_BOUND_WORKLOAD:
					cpuwork()
				case IO_BOUND_WORKLOAD:
					iowork()
				}
				lat := float64(time.Since(taskStart).Microseconds()) / 1e3
				mu.Lock()
				latencies[idx] = lat
				idx++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	totalTime := time.Since(allStart)

	sort.Float64s(latencies)
	totalTaskTime := 0.0
	for _, val := range latencies {
		totalTaskTime += val
	}
	avg := totalTaskTime / float64(totalTasks)

	variance := 0.0
	for _, val := range latencies {
		d := val - avg
		variance += d * d
	}
	variance /= float64(totalTasks)
	stddev := math.Sqrt(variance)
	throughput := float64(totalTasks) / float64(totalTime.Milliseconds())

	return Result{
		TasksPerGoroutine: tasksPerGoroutine,
		TotalTime:         float64(totalTime.Milliseconds()),
		Throughput:        throughput,
		AvgLat:            avg,
		MaxLat:            latencies[len(latencies)-1],
		MinLat:            latencies[0],
		StdLat:            stddev,
		Goroutines:        numGoroutines,
		WorkloadType:      workloadType,
		GOMAXPROCS:        goMacProcs,
	}
}

var csvHeader = []string{
	"tasks_per_goroutine", "total_time_ms", "throughput_ms", "avg_lat_ms", "max_lat_ms",
	"min_lat", "std_lat", "goroutines", "workload_type", "gomaxprocs",
}

func resultToRow(r Result) []string {
	return []string{
		strconv.Itoa(r.TasksPerGoroutine),
		fmt.Sprintf("%.2f", r.TotalTime),
		fmt.Sprintf("%.2f", r.Throughput),
		fmt.Sprintf("%.2f", r.AvgLat),
		fmt.Sprintf("%.2f", r.MaxLat),
		fmt.Sprintf("%.4f", r.MinLat),
		fmt.Sprintf("%.2f", r.StdLat),
		strconv.Itoa(r.Goroutines),
		r.WorkloadType,
		strconv.Itoa(r.GOMAXPROCS),
	}
}

func saveCSV(results []Result) error {
	f, err := os.Create(TABLE_PATH)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	_ = w.Write(csvHeader)
	for _, r := range results {
		_ = w.Write(resultToRow(r))
	}
	w.Flush()
	return w.Error()
}

func savePlots(results []Result) {
	grouped := make(map[string][]Result)
	for _, r := range results {
		grouped[r.WorkloadType] = append(grouped[r.WorkloadType], r)
	}

	type metric struct {
		name  string
		label string
		get   func(r Result) float64
	}

	metrics := []metric{
		{"throughput", "Throughput (tasks/ms)", func(r Result) float64 { return r.Throughput }},
		{"total_time", "Total Time (ms)", func(r Result) float64 { return r.TotalTime }},
		{"avg_latency", "Avg Latency (ms)", func(r Result) float64 { return r.AvgLat }},
		{"max_latency", "Max Latency (ms)", func(r Result) float64 { return r.MaxLat }},
		{"min_latency", "Min Latency (ms)", func(r Result) float64 { return r.MinLat }},
		{"std_latency", "Std Dev Latency (ms)", func(r Result) float64 { return r.StdLat }},
	}

	colors := []color.RGBA{
		{R: 255, A: 255},
		{G: 180, A: 255},
		{B: 255, A: 255},
		{R: 255, G: 165, A: 255},
	}

	for workload, group := range grouped {
		byGMP := make(map[int][]Result)
		for _, r := range group {
			byGMP[r.GOMAXPROCS] = append(byGMP[r.GOMAXPROCS], r)
		}

		var gmpKeys []int
		for k := range byGMP {
			gmpKeys = append(gmpKeys, k)
		}
		sort.Ints(gmpKeys)

		for _, m := range metrics {
			p := plot.New()
			p.Title.Text = fmt.Sprintf("%s vs Goroutines (%s)", m.label, workload)
			p.X.Label.Text = "Goroutines"
			p.Y.Label.Text = m.label

			for i, gmp := range gmpKeys {
				res := byGMP[gmp]

				sort.Slice(res, func(i, j int) bool {
					return res[i].Goroutines < res[j].Goroutines
				})

				pts := make(plotter.XYs, len(res))
				for j, r := range res {
					pts[j].X = float64(r.Goroutines)
					pts[j].Y = m.get(r)
				}

				line, _ := plotter.NewLine(pts)
				line.Color = colors[i%len(colors)]

				p.Add(line)
				p.Legend.Add(fmt.Sprintf("GMP=%d", gmp), line)
			}

			err := p.Save(6*vg.Inch, 4*vg.Inch,
				filepath.Join("results",
					fmt.Sprintf("%s_%s.png", workload, m.name)))
			if err != nil {
				fmt.Println("plot error:", err)
			}
		}
	}
}

func saveResults(results []Result) {
	if err := saveCSV(results); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to save CSV: %v", err)
	}
	savePlots(results)
}

func main() {
	f, err := os.Create("trace.out")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := trace.Start(f); err != nil {
		panic(err)
	}
	defer trace.Stop()

	var results []Result
	for _, gmp := range GOMAXPROCS_vals {
		for _, wl := range []string{CPU_BOUND_WORKLOAD, IO_BOUND_WORKLOAD} {
			for _, g := range gorutineCounts {
				r := runBenchmark(g, wl, gmp)
				results = append(results, r)
			}
		}
	}
	saveResults(results)
}
