#!/usr/bin/env python3
import json
import pika

connection = pika.BlockingConnection(pika.ConnectionParameters(host='localhost'))
channel = connection.channel()

channel.exchange_declare(exchange='task-cancelled',
                         exchange_type='fanout',
                         durable=True)

message = "<TaskCancelledEvent><JobId>620b8251-52a1-4ecd-8adc-4fb280214bba</JobId></TaskCancelledEvent>"

channel.basic_publish(exchange='task-cancelled',
                      routing_key='',
                      body=message)
print(" [x] Sent %r" % message)
connection.close()
