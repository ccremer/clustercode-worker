#!/usr/bin/env python
import json
import sys

import pika

connection = pika.BlockingConnection(pika.ConnectionParameters(host='localhost'))
channel = connection.channel()

channel.queue_declare(queue='slice-added', durable=True)

nr = int(sys.argv[1]) or 0

message = json.dumps({
    "args": [
        "-t", "200",
        "-s", "640x480",
        "-f", "rawvideo",
        "-pix_fmt", "rgb24",
        "-r", "25",
        "-i", "/dev/zero",
        "${output_dir}/empty.mp4",
    ],
    "file": "vendor/empty.mpeg",
    "job_id": "asdf",
    "slice_nr": nr
})
channel.basic_publish(exchange='',
                      routing_key='slice-added',
                      body=message,
                      properties=pika.BasicProperties(
                          delivery_mode=2,  # make message persistent
                          content_type="application/json"
                      ))
print(" [x] Sent %r" % message)
connection.close()
