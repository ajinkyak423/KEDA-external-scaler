import requests
import json

with open('output.json', 'r') as file:
    data = json.load(file)

minikube_ip = '10.0.2.15'  
node_port = 31868  

url = f'http://{minikube_ip}:{node_port}/predict'

headers = {"Prediction-Window": "10m"}

response = requests.post(url, json=data, headers=headers)

print(response.json())
