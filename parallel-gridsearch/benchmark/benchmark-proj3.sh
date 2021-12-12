#!/bin/bash
#
#SBATCH --mail-user=cwbryant@cs.uchicago.edu
#SBATCH --mail-type=ALL
#SBATCH --job-name=proj3_benchmark
#SBATCH --output=./slurm/out/%j.%N.stdout
#SBATCH --error=./slurm/out/%j.%N.stderr
#SBATCH --chdir=/home/cwbryant/parallel/parallel-gridsearch-colebryant/parallel-gridsearch/benchmark
#SBATCH --partition=debug
#SBATCH --nodes=1
#SBATCH --ntasks=1
#SBATCH --cpus-per-task=16
#SBATCH --mem-per-cpu=300
#SBATCH --time=4:00:00

module load golang/1.16.2
python3 speedup_proj3.py
