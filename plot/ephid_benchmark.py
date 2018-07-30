#!/usr/bin/env python3

import glob
from matplotlib import mlab
import matplotlib.pyplot as plt
import numpy as np
from typing import List

LOG_DIR = "apna_benchmark"
EXP_NAME = "ephid_benchmark*"

HOSTID_GENERATION = 0
EXPTIME_GENERATION = 1
ENCRYPT_TIME = 2
MAC_TIME = 3
CERT_TIME = 4

# Types
Ephid_Benchmark = List[List[int]]


def plot_percentile_individual(title: str, data: Ephid_Benchmark):
    fig, axes = plt.subplots()
    axes.boxplot(data, showfliers=False, patch_artist=True, vert=True)
    axes.set_title(title)
    plt.show()


def parse_data(filename: str) -> Ephid_Benchmark:
    data = [[], [], [], []]
    with open(filename, 'r') as fp:
        line = fp.readline()
        while line:
            tmp = line.split()[3:]
            for i in range(4):
                data[i].append(int(tmp[i]))
            line = fp.readline()
    for d in data:
        d.sort()
    return data


def experiments():
    files = glob.glob(LOG_DIR+'/'+EXP_NAME)
    for f in files:
        data = parse_data(f)
        plot_percentile_individual(f, data)


experiments()
