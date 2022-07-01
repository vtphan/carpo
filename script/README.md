### Register jupyterhub users in carpo server
This script registers all the jupyterhub user in the carpo server. It also writes configuration file *(config.json)* for each users in the hub inside the Carpo directory. User can also manually register in the carpo server with the newly created config file.

The input to the script is server address, which should start with *http://*

```bash
python carpo-setup.py --server http://delphinus.cs.memphis.edu:7745
```
