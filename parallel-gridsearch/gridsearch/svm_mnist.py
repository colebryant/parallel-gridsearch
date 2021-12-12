#!/usr/local/bin/python
import sys
import pandas as pd
import numpy as np
from sklearn.svm import SVC
from sklearn.preprocessing import StandardScaler
from sklearn.pipeline import make_pipeline
from sklearn.model_selection import cross_validate

kfolds = int(sys.argv[1])
kernel = sys.argv[2]
c_value = float(sys.argv[3])
gamma = float(sys.argv[4])

mnist_data = pd.read_csv("data/mnist_resampled.csv")

X = mnist_data.iloc[:, 1:]
y = mnist_data.iloc[:, 0]

clf = make_pipeline(StandardScaler(), SVC(kernel=kernel, C=c_value, gamma=gamma))

cv_results = cross_validate(clf, X, y, cv=kfolds, scoring='accuracy')

print(np.average(cv_results['test_score']), end='')

