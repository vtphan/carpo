# SIGCSE_2022  

This file includes the version of the system components and their installation steps.

## Component Versions

- central server: 0.0.1
- carpo-student: 0.0.7
- carpo-teacher: 0.0.6

## Installation

Installation guide for the individual components:

### A. Central Server

1. Clone the repo:
```bash  
    cd carpo/carpo_server
    go build
    ./carpo -c config.json
```

### B. Carpo Student Extension 

Make sure you have JupyterLab server is running such that you can access it via web browser at http://localhost:8888/lab 

1. Open Terminal from JupyterLab Launcher.

2. Install student extension.

```bash  
    pip install carpo-student
```

3. Restart your JupyterLab server.
*There should be `Carpo` directory in the current working directory.*

4. Update the `Carpo/config.json` with your name and server url from A.

### C. Carpo Teacher Extension
Make sure you have JupyterLab server is running such that you can access it via web browser at http://localhost:8889/lab

1. Open Terminal from JupyterLab Launcher.

2. Install teacher extension.

```bash  
    pip install carpo-teacher
```

3. Restart your JupyterLab server.
*There should be `Carpo` directory in the current working directory.*

4. Update the `Carpo/config.json` with your name and server url from A.