#!/usr/bin/env python
import json
import pika

connection = pika.BlockingConnection(pika.ConnectionParameters(host='localhost'))
channel = connection.channel()

channel.exchange_declare(exchange='task-cancelled',
                         exchange_type='fanout',
                         durable=True)

message = json.dumps({
    "job_id": "asdf"
})
channel.basic_publish(exchange='task-cancelled',
                      routing_key='',
                      body=message)
print(" [x] Sent %r" % message)
connection.close()
