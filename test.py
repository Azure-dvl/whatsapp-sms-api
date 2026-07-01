import requests
from flask import Flask, request

app = Flask(__name__)

register_webhook = requests.post("http://localhost:9050/webhook", json={
    "phone": "53013028",
    "url": "http://localhost:8000/receivedSMS",
    "secret": "pepe",
})

j = register_webhook.json()
print(j)

@app.route('/receivedSMS')
def sms():
    print(request)
    return ""

if __name__ == '__main__':
    app.run(port=8000)

    delete_webhook = requests.delete("http://localhost:9050/webhook?phone=53013028")
    j = delete_webhook.json()
    print(j)