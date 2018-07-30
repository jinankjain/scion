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
TOTAL_PARAMS = 5
OFFSET = 3

# Types
Ephid_Benchmark = List[List[int]]


def plot_percentile_individual(title: str, data: Ephid_Benchmark):
    fig, axes = plt.subplots()
    axes.boxplot(data, showfliers=False)
    axes.set_title(title)
    axes.yaxis.grid(True)
    axes.set_ylabel('Time (in nanoseconds)')
    axes.set_xlabel('Different ops involved in ephid generation')
    plt.setp(axes, xticks=[y + 1 for y in range(len(data))],
         xticklabels=['hostID', 'expTime', 'encrypt', 'mac', 'cert', 'total'])
    plt.show()


def parse_data(filename: str) -> Ephid_Benchmark:
    data = [[], [], [], [], [], []]
    with open(filename, 'r') as fp:
        line = fp.readline()
        while line:
            tmp = line.split()[OFFSET:]
            total = 0
            for i in range(TOTAL_PARAMS):
                data[i].append(int(tmp[i]))
                total += int(tmp[i])
            data[TOTAL_PARAMS].append(total)
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
