import json
import requests
import argparse
import os, pwd, grp, sys, subprocess

skip_list = ["kritish", "jupyter-admin", "lost+found"]

config_file = "Carpo/config.json"
carpo_version = "0.0.3"

parser = argparse.ArgumentParser(description='input to the script')

parser.add_argument('--server', type=str, help='server address is a required argument')
args = parser.parse_args()


def get_users_in_system():
    user_directory = os.listdir("/home")
    return [ user for user in user_directory if user not in skip_list ]

def register_student(name: str, server: str):
    endpoint = server + "/add_student"

    # request body
    data = {
        'name': name
    }

    # headers
    headers = {'Content-type': 'application/json', 'Accept': 'text/plain'}
    print(f"Registering student {name}...")
    resp = requests.post(url=endpoint, data=json.dumps(data), headers=headers, timeout=5)
    if resp.status_code >= 200 or resp.status_code <= 299:
        return resp.json()
    else:
        print(f"Failed to register user {name}...")
        return {}

def write_config(resp: dict, name: str, server: str):
    config = {}

    if 'id' in resp:
        config['id'] = resp['id']
    else:
        print(f"Manual registration required for user {name}...")
    
    config['name'] = name
    config['server'] = server
    config['carp_version'] = carpo_version

    config_path = os.path.join("/home/jupyter-" + name, config_file)

    if os.path.exists(config_path):
        os.remove(config_path)

    user_dir = "/home/jupyter-" + name
    if "Carpo" not in os.listdir(os.chdir(user_dir)):
        os.makedirs(os.path.join(user_dir, "Carpo"))

    print(f"Writing config for student {name} with id {resp['id']}...")
    with open(config_path, "w") as file:
        file.write(json.dumps(config, indent=4))

    # Change ownership of the file
    uid = pwd.getpwnam("jupyter-" + name).pw_uid
    gid = grp.getgrnam("jupyter-" + name).gr_gid
    os.chown(config_path, uid, gid)


def main():
    # List user
    # for each user
    # cd to user's home directory
    # check config_path
    # register user to carpo
    # write config to file
    server = args.server
    if not server.startswith("http://"):
        print(f"{server} - server address should start with http://")
        return 

    user_list = get_users_in_system()
    if not user_list:
        print("Failed to get users in the system.")
        return

    for user in user_list:
        # TLJH user: jupyter-amanda
        name = user.replace("jupyter-","")
        student_config = register_student(name, server)
        write_config(student_config, name, server )

if __name__ == '__main__':
    if os.geteuid() == 0:
        main()
    else:
        subprocess.check_call(['sudo', sys.executable] + sys.argv)