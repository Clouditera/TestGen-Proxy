import requests

url = 'http://127.0.0.1:8080/'
username = 'cloud' 
password = 'cloud'

data = {'func_signature': 'value1', 'repository_url': 'value2'}

headers = {'Content-type': 'application/json'}

response = requests.post(url, auth=(username, password), headers=headers, json=data) 

print(response.content)