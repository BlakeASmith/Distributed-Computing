"""Parse data produced for CSC462 Lab 3"""
import json

NUMBER_OF_INPUT_SETS = 8
NUMBER_OF_WORKER_SCENARIOS = 8
NUMBER_OF_RUNS = 20

def stats():
    with open('stats.json') as f:
        for line in f:
            yield json.loads(line)

_stats = list(stats())

stats_per_input_set = [_stats[i:i+NUMBER_OF_INPUT_SETS] for i in 
                       range(0, len(_stats), NUMBER_OF_INPUT_SETS)]

print(len(stats_per_input_set))

stats_sets_per_worker_senarios = [stats_per_input_set[i:i+NUMBER_OF_WORKER_SCENARIOS]
                                  for i in range(0, len(stats_per_input_set), 
                                      NUMBER_OF_WORKER_SCENARIOS)]

bysenarios = [[] for _ in range(NUMBER_OF_WORKER_SCENARIOS)]
for workersenarios in stats_sets_per_worker_senarios:
    for i, lst in enumerate(bysenarios):
        lst.append(workersenarios[i])
    
print(bysenarios)




