import os
import json
from pathlib import Path

textfiles = list(Path(__file__).parent.glob('pg-*'))
of = open('sequential_stats.json', 'a')

for i, txt in enumerate(textfiles[:-1]):
    running = textfiles[:i+1]
    print(f"running mrsequential.go with {len(running)} files")
    cmd = "go run ./mrsequential.go wc.so " + " ".join([str(p) for p in running])
    stream = os.popen(cmd)
    output = stream.read()
    _json = json.loads(output)
    _json['label'] = f'seq_{len(running)}'
    jsonstr = json.dumps(_json, indent=4, sort_keys=True)
    of.write(jsonstr + '\n')
