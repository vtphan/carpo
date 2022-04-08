from distutils.command.config import config
import json
from socket import timeout
import time

from jupyter_server.base.handlers import APIHandler
from jupyter_server.utils import url_path_join
import tornado

import requests
import os
from pathlib import Path
import uuid


def read_config_file():
    """
    reads config.json file
    :return: dict
    """
    config_file = os.path.join(os.getcwd() ,"Carpo",'config.json')
    if os.path.exists(config_file):
        f=open(config_file)
        return json.load(f)
    return {}

def create_initial_files():
    print('======================================')
    print("Check and Create Initial Files:")
    print('======================================')
    current_dir = os.getcwd()
    print(current_dir)
    if "Carpo" not in os.listdir():
        os.makedirs(os.path.join(current_dir,"Carpo"))
   
    # Create config.json file
    config_path = os.path.join(current_dir,"Carpo","config.json")
    if not os.path.isfile(config_path):
        config_data = {}
        config_data['name'] = "Dummy User"
        config_data['server'] = "http://localhost:8081"
        # Write dummy config
        with open(config_path, "w") as config_file:
            config_file.write(json.dumps(config_data))
    
    # Create blank notebook
    notebook_path = os.path.join(current_dir,"Carpo","Untitled.ipynb")
    if not os.path.isfile(notebook_path):
        content = {
                        "cells": [],
                        "metadata": {
                            "kernelspec": {
                                "display_name": "Python 3 (ipykernel)",
                                "language": "python",
                                "name": "python3"
                                },
                            "language_info": {
                            "codemirror_mode": {
                                "name": "ipython",
                                "version": 3
                                },
                            "file_extension": ".py",
                            "mimetype": "text/x-python",
                            "name": "python",
                            "nbconvert_exporter": "python",
                            "pygments_lexer": "ipython3",
                            "version": "3.8.10"
                            }
                        },
                        "nbformat": 4,
                        "nbformat_minor": 5
                    }

        with open(notebook_path, "w") as file:
            file.write(json.dumps(content, indent = 4))


class RegistrationHandler(APIHandler):
    def initialize(self,config_files):
        self.config_files = config_files
    @tornado.web.authenticated
    def get(self):

        config_data = read_config_file()
        if config_data == {}:
            create_initial_files()
            
            self.set_status(500)
            self.finish(json.dumps({'message': "Update your User Name and Server address in Carpo/config.json file and Re-register again."}))
            return
            
        if not {'name','server'}.issubset(config_data):
            self.set_status(500)
            self.finish(json.dumps({'message': "Invalid config.json file. Please check your config file."}))
            return
        
        if config_data['name'] == "Dummy User":
            self.set_status(500)
            self.finish(json.dumps({'message': "Update your User Name and Server address in Carpo/config.json file and Re-register again."}))
            return
        
        url = config_data['server'] + "/add_teacher"

        body = {}
        body['name'] = config_data['name']

        headers = {'Content-type': 'application/json', 'Accept': 'text/plain'}
       
        try:
            response = requests.post(url, data=json.dumps(body),headers=headers,timeout=5).json()
            print(response)
        except requests.exceptions.RequestException as e:
            self.set_status(500)
            self.finish(json.dumps({'message': "Carpo Server Error. {}".format(e)}))
            return

        config_data['id'] = response['id']
        # Write id to the json file.
        with open(os.path.join(os.getcwd(),"Carpo",'config.json'), "w") as config_file:
            config_file.write(json.dumps(config_data))
        print(response)
        self.finish(response)
        
class RouteHandler(APIHandler):
    # The following decorator should be present on all verb methods (head, get, post,
    # patch, put, delete, options) to ensure only authorized user can request the
    # Jupyter server
    @tornado.web.authenticated
    def get(self):

        config_data = read_config_file()

        if not {'id', 'server'}.issubset(config_data):
            self.set_status(500)
            self.finish(json.dumps({'message': "User is not registered. Please Register User."}))
            return
        
        url = config_data['server'] + "/teachers/submissions"
        
        try:
            response = requests.get(url,timeout=5).json()
        except requests.exceptions.RequestException as e:
            self.set_status(500)
            self.finish(json.dumps({'message': "Carpo Server Error. {}".format(e)}))
            return

        # Write response to individual Notebook
        self.submission_file(response['data'])

        self.finish(response)

    def post(self):
        input_data = self.get_json_body()
        headers = {'Content-type': 'application/json', 'Accept': 'text/plain'}
        
        config_data = read_config_file()
        if not {'id','server'}.issubset(config_data):
            self.set_status(500)
            self.finish(json.dumps({'message': "User is not registered. Please Register User."}))
            return
        
        url = config_data['server'] + "/teachers/submissions"

        try:
            response = requests.post(url, data=json.dumps(input_data),headers=headers,timeout=5).json()
        except requests.exceptions.RequestException as e:
            self.set_status(500)
            self.finish(json.dumps({'message': "Carpo Server Error. {}".format(e)}))
            return

        # Delete the local submission notebook
        notebook_path = os.path.join("Carpo", "Submissions", "problem_{}".format(input_data['problem_id']), "{:03d}".format(input_data['submission_id']) + ".ipynb" )
        if os.path.exists(notebook_path):
            os.remove(notebook_path)

        self.finish(response)

    
    def submission_file(self, data):

        for res in data:
            dir_path = os.path.join("Carpo", "Submissions", "problem_{}".format(res['problem_id']))
            file_path = "submission_{:03d}".format(res['id']) + ".ipynb"
            if not os.path.exists(dir_path):
                os.makedirs(dir_path)

            submission_file = os.path.join(dir_path, file_path)
            if not os.path.exists(submission_file):
                content = {
                        "cells": [],
                        "metadata": {
                            "kernelspec": {
                                "display_name": "Python 3 (ipykernel)",
                                "language": "python",
                                "name": "python3"
                                },
                            "language_info": {
                            "codemirror_mode": {
                                "name": "ipython",
                                "version": 3
                                },
                            "file_extension": ".py",
                            "mimetype": "text/x-python",
                            "name": "python",
                            "nbconvert_exporter": "python",
                            "pygments_lexer": "ipython3",
                            "version": "3.8.10"
                            }
                        },
                        "nbformat": 4,
                        "nbformat_minor": 5
                    }
                info_block = ["---\n"]
                content["cells"].append({
                        "cell_type": "markdown",
                        "id": str(uuid.uuid4()),
                        "metadata": {},
                        "source": info_block + [ x+"\n" for x in res['info'].split("\n") ]
                        })

                msg_prefix = ["Student: {} at {} says:\n ".format(res['student_name'], res['time'])]
                content["cells"].append({
                        "cell_type": "markdown",
                        "id": str(uuid.uuid4()),
                        "metadata": {},
                        "source": msg_prefix + [ x+"\n" for x in res['message'].split("\n") ]
                        })

                code_block = ["#{} {} {}\n".format(res['student_id'], res['problem_id'], res['id'])]
                content["cells"].append({
                        "cell_type": "code",
                        "execution_count": 0,
                        "id": str(uuid.uuid4()),
                        "metadata": {},
                        "source": code_block + [ x+"\n" for x in res['code'].split("\n") ],
                        "outputs": []
                        })

                content["cells"].append({
                        "cell_type": "markdown",
                        "id": str(uuid.uuid4()),
                        "metadata": {},
                        "source": "Instructor Feedback for " + res['student_name'] + " :\n\n"
                        })

                # Serializing json 
                json_object = json.dumps(content, indent = 4)

                with open(submission_file, "w") as file:
                    file.write(json_object)
            
class ProblemHandler(APIHandler):

    @tornado.web.authenticated
    def post(self):
        input_data = self.get_json_body()

        config_data = read_config_file()
        if not {'id','server'}.issubset(config_data):
            self.set_status(500)
            self.finish(json.dumps({'message': "User is not registered. Please Register User."}))
            return
        
        input_data['teacher_id'] = config_data['id']
        
        url = config_data['server'] + "/problem"

        headers = {'Content-type': 'application/json', 'Accept': 'text/plain'}
        try:
            response = requests.post(url, data=json.dumps(input_data),headers=headers, timeout=5).json()
        except requests.exceptions.RequestException as e:
            self.set_status(500)
            self.finish(json.dumps({'message': "Carpo Server Error. {}".format(e)}))
            return

        self.finish(response)

    @tornado.web.authenticated
    def delete(self):
        input_data = self.get_json_body()

        config_data = read_config_file()

        if not {'id','server'}.issubset(config_data):
            self.set_status(500)
            self.finish(json.dumps({'message': "User is not registered. Please Register User."}))
            return

        input_data['teacher_id'] = config_data['id']
        url = config_data['server'] + "/problem"

        print("Input Data: ", input_data)
        headers = {'Content-type': 'application/json', 'Accept': 'text/plain'}
        try:
            response = requests.delete(url, data=json.dumps(input_data),headers=headers, timeout=5)
        except requests.exceptions.RequestException as e:
            self.set_status(500)
            self.finish(json.dumps({'message': "Carpo Server Error. {}".format(e)}))
            return
        self.finish(json.dumps(response.text))
    
class GradeHandler(APIHandler):

    @tornado.web.authenticated
    def post(self):
        input_data = self.get_json_body()

        config_data = read_config_file()
        if not {'id','server'}.issubset(config_data):
            self.set_status(500)
            self.finish(json.dumps({'message': "User is not registered. Please Register User."}))
            return

        input_data['teacher_id'] = config_data['id']
        url = config_data['server'] + "/submissions/grade"

        print("Input Data: ", input_data)
        headers = {'Content-type': 'application/json', 'Accept': 'text/plain'}
        response = requests.post(url, data=json.dumps(input_data),headers=headers,timeout=5)

        data = {
            "go-server": response.json()
        }
        self.finish(data)
class FeedbackHandler(APIHandler):

    @tornado.web.authenticated
    def post(self):
        input_data = self.get_json_body()

        config_data = read_config_file()
        if not {'id','server'}.issubset(config_data):
            self.set_status(500)
            self.finish(json.dumps({'message': "User is not registered. Please Register User."}))
            return

        input_data['teacher_id'] = config_data['id']
        url = config_data['server'] + "/teachers/feedbacks" 

        headers = {'Content-type': 'application/json', 'Accept': 'text/plain'}
        try:
            response = requests.post(url, data=json.dumps(input_data),headers=headers,timeout=5).json()
        except requests.exceptions.RequestException as e:
            self.set_status(500)
            self.finish(json.dumps({'message': "Carpo Server Error. {}".format(e)}))
            return


        self.finish(response)
    
def setup_handlers(web_app):
    host_pattern = ".*$"

    base_url = web_app.settings["base_url"]
    route_pattern_code = url_path_join(base_url, "carpo-teacher", "submissions")
    handlers = [(route_pattern_code, RouteHandler)]
    web_app.add_handlers(host_pattern, handlers)

    route_pattern_problem =  url_path_join(web_app.settings['base_url'], "carpo-teacher", "problem")
    web_app.add_handlers(host_pattern, [(route_pattern_problem, ProblemHandler)])

    route_pattern_grade =  url_path_join(web_app.settings['base_url'], "carpo-teacher", "submissions/grade")
    web_app.add_handlers(host_pattern, [(route_pattern_grade, GradeHandler)])

    route_pattern_grade =  url_path_join(web_app.settings['base_url'], "carpo-teacher", "submissions/feedbacks")
    web_app.add_handlers(host_pattern, [(route_pattern_grade, FeedbackHandler)])


    route_pattern_grade =  url_path_join(web_app.settings['base_url'], "carpo-teacher", "register")
    web_app.add_handlers(host_pattern, [(route_pattern_grade, RegistrationHandler, dict(config_files = create_initial_files()))])


