Paper link: https://tonymuu.github.io/assets/GoLFS-paper.pdf

# Getting Started

Follow these steps to set up your first GoLFS cluster:
1. From a Linux system, download the `build.zip` from https://github.com/tonymuu/GLFS/releases/tag/v0.0.1-pre-alpha. The compressed file contains both server and application executables, as well as a few convenient scripts used for setting up clusters, running tests, and doing evaluations.
2. Extract all files into an conveniently accessible folder, navigate to that folder from a commandline.
3. The quickest way to start playing around with the system is to use convient scripts: run `/bin/bash setup_cluster.sh 7` which will set up a GoLFS cluster with a single master and 7 chunk servers in the background. Then you can run `./build/app -mode i` to start an interactive test application. You can issue simple commands like create, write, read, or delete from the test application. To shutdown the cluster, run `/bin/bash terminate.sh`.
4. To run evaluation mode,
5. To generate evaluation report, simply run `/bin/bash run_report.sh`. The report output will be in the file `./eval/eval.txt`.

# Debug logs
All output logs are located under the `./logs/` folder. Specifically, evaluation output logs are located under `./eval_output.txt`.
