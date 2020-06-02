import subprocess
import time
from pathlib import Path

plugins = {
        "wc.so":"../mrapps/wc.go",
        "crash.so":"../mrapps/crash.go",
}

def run(cmd):
    print(cmd)
    return subprocess.Popen(cmd.split())

def gocompile(plugin):
    """compile a go plugin"""
    cmd = f"go build -buildmode=plugin {plugin}"
    run(cmd).wait()
    
def inputs():
    """generator to produce input sets"""
    inputfiles = list(Path(__file__).parent.glob('pg-*'))
    for i, _ in enumerate(inputfiles[:-1]):
        yield [str(p) for p in inputfiles[:i+1]]

def startworkers(n, plugsrc):
    processes = []
    for _ in range(n):
        cmd = f"go run ./mrworker.go {plugsrc}"
        processes.append(run(cmd))
    return processes

def runmaster(plugins, inputset):
    """runs the master on all of the plugins and input sets"""
    cmd_args = " ".join(inputset)
    cmd = f"go run ./mrmaster.go {cmd_args}"
    return run(cmd)


procs = []
try:
    for plug, src in plugins.items():
        for _ in range(20):
            for i in reversed(range(1, 8)):
                    gocompile(src)
                    for testset in inputs():
                        print('starting master')
                        procs.append(runmaster(plug, testset))
                        time.sleep(1)
                        print('starting workers')
                        procs.extend(startworkers(i, plug))
                        exit_codes = [p.wait() for p in procs]
                        print(exit_codes)
except KeyboardInterrupt:
    exit()








 
