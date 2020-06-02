import os
import json
import matplotlib.pyplot as plt
from pathlib import Path

def run_sequential():
    textfiles = list(Path(__file__).parent.glob('pg-*'))
    with open('sequential_stats.json', 'w') as of:

        results = []
        for i, _ in enumerate(textfiles[:-1]):
            running = textfiles[:i+1]
            print(f"running mrsequential.go with {len(running)} files")
            cmd = "go run ./mrsequential.go wc.so " + " ".join([str(p) for p in running])
            stream = os.popen(cmd)
            output = stream.read()
            _json = json.loads(output)
            _json['label'] = f'seq_{len(running)}'
            _json['nbytes'] = sum((len(it.read_text()) for it in running))
            results.append(_json)
        json.dump(results, of,sort_keys=True)

def graph_sequential():
    stats = []
    with open('sequential_stats.json', 'r') as stat:
        stats = json.loads(stat.read())

    times = [it['TotalTime'] / 1000000 for it in stats]
    numbytes = [it['nbytes'] for it in stats]

    plt.plot(numbytes, times, linestyle='--', marker='o', label="TotalTime")

    plt.plot(numbytes, [it['MapSeqenceTime'] / 1000000 for it in stats], linestyle='--', 
            marker='o', label="Time in Mapping Phase")

    plt.plot(numbytes, [it['ReduceSequenceTime'] / 1000000 for it in stats], linestyle='--', 
            marker='o', label="Time in Reducing Phase")

    plt.plot(numbytes, [it['SortTime'] / 1000000 for it in stats], linestyle='--', 
            marker='o', label="Time spend Sorting Intermedate Keys")

    plt.ylabel('Runtime in ms')
    plt.xlabel('Number of Bytes In Input')
    for i, label in enumerate([it['label'] for it in stats]):
        plt.annotate(label, (times[i], numbytes[i]))

    plt.legend()
    plt.savefig('seq_time_perf.png')


# run_sequential()
graph_sequential()





