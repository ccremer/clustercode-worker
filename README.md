# clustercode-worker

Worker microservice for clustercode [WIP]

## Building

    go mod vendor

## Running

    # Run
    go build main.go
    ./clustercode-worker
    # OR
    go run main.go

## Configuring for local development

- Copy `defaults.yaml` to `config.yaml`
- Create the directories `input`, `output` in the current working dir.
- Change the input and output dirs in `config.yaml`
- Delete other values in `config.yaml` that are already default, leaving the custom ones left.

## Testing

Unit tests:

    go test ./...
    
Integration tests:

    docker network create clustercode
    docker-compose up -d
    # You may then send some messages to rabbitmq
    test/task.py
    test/slice.py
    test/cancel.py

# Concept

- Split video with `ffmpeg -i movie.mp4 -c copy -map 0 -segment_time 120 -f segment movie_segment_%d.mkv` on shovel node
- Transcode each segment with `ffmpeg -i movie_segment_1.mp4 -c:v copy -c:a copy movie_segment_1.mkv`
  on compute node (with whatever parameters)
- Create the concat file: `echo "file movie_segment_1.mkv" >> concat.txt` (make sure they are sorted!)
- Merge the segments back into 1 file: `ffmpeg -f concat -i concat.txt -c copy movie_out.mkv`
