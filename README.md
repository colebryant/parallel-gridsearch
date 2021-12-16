# README
* Please ensure that Pandas, NumPy, and Scikit-Learn are installed in the Python environment. The program will not run without these libraries installed
* I am using a heavily downsampled (balanced) version of the mnist dataset, with 3600 entries. This is for the sake of speed, especially since we are performing cross-validation on the model. The dataset is called mnist_resampled.csv and is stored in the gridsearch/data directory
* To run:
  * `cd` into the `gridsearch` directory, and run `go run gridsearch.go [paramsize] [kfolds] [threads]`
  * Usage:
    * paramsize = the size of the parameter json file to perform gridsearch on. This should be 'small', 'medium', or 'large'
    * kfolds = the number of kfolds in the model cross validation (must be > 1). Recommend 2 or 3 for speedier performance
    * threads = the number of threads (optional, will run parallel version)
  * To run the program sequentially, simply run without the `[threads]` parameter
  * When completed, the program will print out the best parameters and associated training accuracy

<img src="benchmark/graphs/parallel-speedup.png" width="600"/>
