import json

from jupyter_server.base.handlers import APIHandler
from jupyter_server.utils import url_path_join
import tornado

import requests
import os
from pathlib import Path
import uuid


class RegistrationHandler(APIHandler):
    @tornado.web.authenticated
    def get(self):

        f=open(os.getcwd()+'/teacher_config.json')
        config_data = json.load(f)
        
        url = config_data['server'] + "/add_teacher"

        body = {}
        body['name'] = config_data['name']

        headers = {'Content-type': 'application/json', 'Accept': 'text/plain'}
        response = requests.post(url, data=json.dumps(body),headers=headers)

        data = {
            "go-server": response.json()
        }
        self.finish(json.dumps(data))
class RouteHandler(APIHandler):
    # The following decorator should be present on all verb methods (head, get, post,
    # patch, put, delete, options) to ensure only authorized user can request the
    # Jupyter server
    @tornado.web.authenticated
    def get(self):
        # Check if submission file exist or not 
        self.submission_file()

        response = requests.get("http://localhost:8081/teachers/submissions")
        self.finish(response.json())
    
    def submission_file(self):
        path = 'FeedbackData'
        file = 'all_submissions.ipynb'

        if not os.path.exists(path):
            os.makedirs(path)

        if not os.path.exists(os.path.join(path, file)):
            print("Creating File in path {}/{}".format(path,file))
            submission_file = os.path.join(path, file)
        
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
    

            # Serializing json 
            json_object = json.dumps(content, indent = 4)
            
            # Writing to sample.json
            with open(submission_file, "w") as outfile:
                outfile.write(json_object)
            outfile.close()

class ProblemHandler(APIHandler):

    @tornado.web.authenticated
    def post(self):
        input_data = self.get_json_body()

        # Read Json file and add infos.
        f=open(os.getcwd()+'/teacher_config.json')
        config_data = json.load(f)
        input_data['teacher_id'] = config_data['id']
        url = config_data['server'] + "/problem"

        print("Input Data: ", input_data)
        headers = {'Content-type': 'application/json', 'Accept': 'text/plain'}
        response = requests.post(url, data=json.dumps(input_data),headers=headers)

        data = {
            "go-server": response.json()
        }

        self.finish(json.dumps(data))
    
class GradeHandler(APIHandler):

    @tornado.web.authenticated
    def post(self):
        input_data = self.get_json_body()

        # Read Json file and add infos.
        f=open(os.getcwd()+'/teacher_config.json')
        config_data = json.load(f)

        input_data['teacher_id'] = 1
        url = config_data['server'] + "/submissions/grade"

        print("Input Data: ", input_data)
        headers = {'Content-type': 'application/json', 'Accept': 'text/plain'}
        response = requests.post(url, data=json.dumps(input_data),headers=headers)

        data = {
            "go-server": response.json()
        }
        self.finish(json.dumps(data))
class FeedbackHandler(APIHandler):

    @tornado.web.authenticated
    def post(self):
        input_data = self.get_json_body()

        # Read Json file and add infos.
        f=open(os.getcwd()+'/teacher_config.json')
        config_data = json.load(f)
        input_data['teacher_id'] = 1
        url = config_data['server'] + "/teachers/feedbacks" #TODO create this endpoint

        print("Input Data: ", input_data)
        headers = {'Content-type': 'application/json', 'Accept': 'text/plain'}
        response = requests.post(url, data=json.dumps(input_data),headers=headers)

        data = {
            "go-server": response.json()
        }

        self.feedback_file()

        self.finish(json.dumps(data))
    
    def feedback_file(self):
        path = 'FeedbackData'
        # Get student name from id and use the name in the filename.
        question_id = 100
        url = "http://localhost:8081/students/get_submission_feedbacks?question_id="+str(question_id)
        response = requests.get(url)
        resp = response.json()

        if len(resp['data']) == 0:
            return
            
        studnet_name = resp['data'][0]['name']    
        file = 'feedback_'+ studnet_name + '.ipynb'
        submission_file = os.path.join(path, file)
        sub_file_exist = os.path.exists(os.path.join(path, file))

        if not sub_file_exist:
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
        

            # Serializing json 
            json_object = json.dumps(content, indent = 4)
            
            # Writing to sample.json
            with open(submission_file, "w") as outfile:
                outfile.write(json_object)
            outfile.close()

        # Read the same file and append the new blocks:
        with open(submission_file, "r") as file:
            data = json.loads(file.read())
            data['cells'] = []
            
            for feedback in resp['data']:
                data["cells"].append({
                        "cell_type": "markdown",
                        "id": str(uuid.uuid4()),
                        "metadata": {},
                        "source": [ x+"\n" for x in feedback['comment'].split("\n") ]
                        })
                data["cells"].append({
                        "cell_type": "code",
                        "id": str(uuid.uuid4()),
                        "metadata": {},
                        "source": [ x+"\n" for x in feedback['code_feedback'].split("\n") ],
                        "outputs": []
                        })

        with open(submission_file, "w") as file:
                json.dump(data, file)
    

def setup_handlers(web_app):
    host_pattern = ".*$"

    base_url = web_app.settings["base_url"]
    route_pattern_code = url_path_join(base_url, "teacher-ext", "code")
    handlers = [(route_pattern_code, RouteHandler)]
    web_app.add_handlers(host_pattern, handlers)

    route_pattern_problem =  url_path_join(web_app.settings['base_url'], "teacher-ext", "problem")
    web_app.add_handlers(host_pattern, [(route_pattern_problem, ProblemHandler)])

    route_pattern_grade =  url_path_join(web_app.settings['base_url'], "teacher-ext", "submissions/grade")
    web_app.add_handlers(host_pattern, [(route_pattern_grade, GradeHandler)])

    route_pattern_grade =  url_path_join(web_app.settings['base_url'], "teacher-ext", "submissions/feedbacks")
    web_app.add_handlers(host_pattern, [(route_pattern_grade, FeedbackHandler)])


    route_pattern_grade =  url_path_join(web_app.settings['base_url'], "teacher-ext", "register")
    web_app.add_handlers(host_pattern, [(route_pattern_grade, RegistrationHandler)])



