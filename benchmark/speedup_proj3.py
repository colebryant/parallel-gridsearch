import subprocess
import time
import matplotlib.pyplot as plt
import sys

# Constants
THREAD_NUMS = [2, 4, 6, 8, 12]
PARAM_SIZES = ['small', 'medium', 'large']
PARAM_SIZE_NUMS = {'small': 12, 'medium': 24, 'large': 36}


def get_sequential_times():
    """
    Runs the sequential version of grid search on each of the parameter sizes.
    No inputs.
    Outputs: sequential times (dictionary of keys:sizes and values:average times)
    """
    print(f'====sequential====')

    sequential_times = {}

    for size in PARAM_SIZES:
        # Find average of sequential execution times
        sequential_times_size = []
        # Repeat 3 times
        for _ in range(3):
            begin = time.time()
            subprocess.Popen(f'go run ../gridsearch/gridsearch.go {size} 2', shell=True).wait()
            end = time.time()
            diff = end - begin
            sequential_times_size.append(diff)

        # Calculate average sequential time
        avg_sequential = sum(sequential_times_size) / len(sequential_times_size)
        sequential_times[size] = avg_sequential
        print(f'Average sequential time on {size}: {avg_sequential}')

    return sequential_times


def produce_graph(sequential_times):
    """
    Runs the parallel version of grid search and produces the speedup graph.
    Inputs: sequential times (dictionary of keys:sizes and values:average times)
    Outputs: speedup graph
    """
    print(f'====parallel====')
    for size in PARAM_SIZES:
        sequential_time = sequential_times[size]
        # Loop through number of threads
        avg_threads = []
        for thread_num in THREAD_NUMS:
            thread_num_times = []
            # Repeat 3 times
            for i in range(3):
                begin = time.time()
                subprocess.Popen(f'go run ../gridsearch/gridsearch.go {size} 2 {thread_num}', shell=True).wait()
                end = time.time()
                diff = end - begin
                thread_num_times.append(diff)
            avg_thread_time = sum(thread_num_times) / len(thread_num_times)
            avg_threads.append(avg_thread_time)
            print(f'Average parallel time on {size} with {thread_num} threads: {avg_thread_time}')

        speedups = [round(sequential_time / avg_threads[i], 2) for i in range(len(avg_threads))]
        param_size_num = PARAM_SIZE_NUMS[size]
        plt.plot(THREAD_NUMS, speedups, label=f'{size}({param_size_num})')

    plt.title(f'GridSearch Speedup Graph')
    plt.xlabel('Number of Threads')
    plt.ylabel('Speedup')
    plt.legend(title='Param Combination Size')
    plt.grid()
    plt.savefig(f'graphs/parallel-speedup.png')


if __name__ == '__main__':

    sequential_times = get_sequential_times()
    produce_graph(sequential_times)
