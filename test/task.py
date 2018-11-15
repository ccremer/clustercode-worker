#!/usr/bin/env python
import json
import pika

connection = pika.BlockingConnection(pika.ConnectionParameters(host='localhost'))
channel = connection.channel()

channel.queue_declare(queue='task-added', durable=True)

message = json.dumps({
    "priority": 0,
    "slize_size": 120,
    "args": [
        "-hide_banner",
        "-y",
        "-i",
        "${input_dir}/empty.mp4",
        "-c", "copy",
        "-map", "0",
        "-segment_time", "120",
        "-f", "segment",
        "${tmp_dir}/empty_segment_%d.mp4"],
    "file": "empty.mpeg",
    "job_id": "asdf"
})

channel.basic_publish(exchange='',
                      routing_key='task-added',
                      body=message,
                      properties=pika.BasicProperties(
                          delivery_mode=2,  # make message persistent
                          content_type="application/json"
                      ))
print(" [x] Sent %r" % message)
connection.close()
