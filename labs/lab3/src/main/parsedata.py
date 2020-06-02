"""Parse data produced for CSC462 Lab 3"""
import json
import pandas
from plotly import graph_objects as go

NUMBER_OF_INPUT_SETS = 8
NUMBER_OF_WORKER_SCENARIOS = 7
NUMBER_OF_RUNS = 20

def stats():
    with open('stats.json') as f:
        for line in f:
            yield json.loads(line)

statgen = stats()


def stats_by_input_sets():

    _istats = [[]] * NUMBER_OF_INPUT_SETS

    i = 0
    while True:
        stat = next(statgen, None)
        if not stat:
            break
        _istats[i].append(stat)
        i += 1
        if i == NUMBER_OF_INPUT_SETS:
            i = 0

    yield from _istats



def stats_by_num_workers():
    i = 0
    stats_sets_per_worker_senarios = [[]] * NUMBER_OF_RUNS
    istats = stats_by_input_sets()
    while True:

        statset = next(istats, None)
        if not statset:
            break

        stats_sets_per_worker_senarios[i].append(statset)
        i+=1

        if i == NUMBER_OF_WORKER_SCENARIOS:
            i = 0

    return stats_sets_per_worker_senarios

df = pandas.DataFrame.from_records(stats_by_num_workers())



fig = go.Figure(data=[go.Table(
    header=dict(values=list(df.columns),
                fill_color='paleturquoise',
                align='left'),
    cells=dict(values=[df.TotalTime],
               fill_color='lavender',
               align='left'))
])

fig.show()



