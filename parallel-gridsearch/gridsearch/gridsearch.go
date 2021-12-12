package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"proj3/queue"
	"proj3/worker"
	"runtime"
	"strconv"
	"sync"
)

// Parameters struct holds the hyperparameters gridsearch is performed on
type Parameters struct {
	Kernel []string `json:"kernel"`
	C      []float64    `json:"C"`
	Gamma []float64     `json:"gamma"`
}


// sequential performs the gridsearch operation in sequential fashion
func sequential(params Parameters, kFolds int) {
	var trainAccuracies []float64
	var paramSchemes [][]string
	kFoldsString := fmt.Sprintf("%d", kFolds)

	// Loop through all combinations of parameters
	for _, kernel := range params.Kernel {
		for _, c := range params.C {
			for _, gamma := range params.Gamma {
				cString := fmt.Sprintf("%f", c)
				gammaString := fmt.Sprintf("%f", gamma)

				// Execute python modeling script and get output
				accuracy := runModel(kFoldsString, kernel, cString, gammaString)

				// Add accuracy output and parameter scheme to list
				trainAccuracies = append(trainAccuracies, accuracy)
				currentParams := []string{kernel, cString, gammaString}
				paramSchemes = append(paramSchemes, currentParams)
			}
		}
	}
	determineOutput(trainAccuracies, paramSchemes)
}


// parallelWorkBalancing performs the gridsearch operation in parallel fashion (with multiple worker queues and work-sharing)
func parallelWorkBalancing(params Parameters, kFolds int, threads int) {

	totalWork := len(params.Kernel) * len(params.C) * len(params.Gamma)
	kFoldsString := fmt.Sprintf("%d", kFolds)

	ctx := &worker.SharedContext{Mutex: &sync.Mutex{}, Group: &sync.WaitGroup{}, TrainAccuracies: make([]float64, totalWork), ParamSchemes: make([][]string, totalWork)}

	// Allocate and instantiate queues and workers
	queues := make([]queue.Queue, threads)
	workers := make([]*worker.Worker, threads)
	for i := 0; i < threads; i++ {
		queues[i] = queue.NewCoarseGrainedQueue()
	}
	for i := 0; i < threads; i++ {
		workers[i] = worker.NewWorker(i, ctx, queues)
	}

	// Determine amount of work per worker
	workAmount := totalWork / threads
	workCount := 0

	// Loop through all combinations of parameters
	for _, kernel := range params.Kernel {
		for _, c := range params.C {
			for _, gamma := range params.Gamma {
				// Get current queue index
				currentQueueIndex := workCount / workAmount
				// Allocate the extra work for the last queue
				if currentQueueIndex >= threads {
					currentQueueIndex = threads - 1
				}
				// Get current queue
				currentQueue := queues[currentQueueIndex]

				currentKernel := kernel
				cString := fmt.Sprintf("%f", c)
				gammaString := fmt.Sprintf("%f", gamma)
				currentCount := workCount

				// Create Runnable task
				trainModel := func(arg interface{}) {

					// Use closure to allocate kernel, c, gamma, workIndex
					myKernel := currentKernel
					myCString := cString
					myGammaString := gammaString
					myWorkIndex := currentCount

					argCtx := arg.(*worker.SharedContext)

					// Execute python modeling script and get output
					accuracy := runModel(kFoldsString, myKernel, myCString, myGammaString)

					// Add accuracy output and parameter scheme to list
					argCtx.TrainAccuracies[myWorkIndex] = accuracy
					currentParams := []string{myKernel, cString, gammaString}
					argCtx.ParamSchemes[myWorkIndex] = currentParams
				}

				// Push runnable task into queue
				currentQueue.Enqueue(trainModel)

				// Increment work count when finished
				workCount++
			}
		}
	}

	// Call Run() on each of the workers
	for _, w := range workers {
		w.Run()
	}

	// Call Exit() on each of the workers
	for _, w := range workers {
		w.Exit()
	}

	ctx.Group.Wait()
	determineOutput(ctx.TrainAccuracies, ctx.ParamSchemes)
}


// runModel runs the SVM model via the python script with given parameters and returns the training accuracy
func runModel(kFolds string, kernel string, c string, gamma string) float64 {
	_, b, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(b)
	cmd := exec.Command("python3", "svm_mnist.py", kFolds, kernel, c, gamma)
	cmd.Dir = basePath
	out, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}
	accuracy, _ := strconv.ParseFloat(string(out), 64)
	if err != nil {
		panic(err)
	}
	return accuracy
}


// determineOutput prints out the best model training accuracy and associated hyper parameters
func determineOutput(trainAccuracies []float64, paramSchemes[][]string) {
	var max float64
	var maxIndex int
	for i := 0; i < len(trainAccuracies); i++ {
		if trainAccuracies[i] > max {
			max = trainAccuracies[i]
			maxIndex = i
		}
	}
	bestParams := paramSchemes[maxIndex]
	output := fmt.Sprintf("Best SVM classifier training accuracy: %f\nUsing parameters: \n" +
		"    kernel: %v\n    C: %s\n    gamma: %s\n", max, bestParams[0], bestParams[1], bestParams[2])
	fmt.Print(output)
}


// main function invoked with command line
func main() {

	const usage = "Usage: gridsearch paramsize kfolds [threads]\n" +
		"    paramsize = the size of the parameter json file to perform gridsearch on. This should be 'small', 'medium', or 'large'\n" +
		"    kfolds = the number of kfolds in the model cross validation (must be > 1). Recommend 2 or 3 for speedier performance\n" +
		"    threads = the number of threads (optional, will run parallel version)\n"

	// Print usage statement for incorrect commands
	if len(os.Args) != 3 && len(os.Args) != 4 {
		fmt.Print(usage)
		os.Exit(2)
	}

	paramSize := os.Args[1]

	// Read in hyperparameters to perform gridsearch on
	_, b, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(b)
	file, err := ioutil.ReadFile(basePath+"/inputs/params_" + paramSize + ".json")
	if err != nil {
		panic(err)
	}

	params := Parameters{}
	_ = json.Unmarshal(file, &params)

	if len(os.Args) == 3 { 	// Run sequential version
		kFolds, _ := strconv.Atoi(os.Args[2])
		sequential(params, kFolds)

	} else { // Run parallel version
		kFolds, _ := strconv.Atoi(os.Args[2])
		threads, _ := strconv.Atoi(os.Args[3])
		parallelWorkBalancing(params, kFolds, threads)
	}
}
