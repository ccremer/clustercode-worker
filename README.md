# clustercode-worker

Worker microservice for clustercode [WIP]

## Building

    sudo snap install go --classic # or whatever package manager you use
    sudo apt-get install g++
    go mod vendor
    go build

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

## Branches

* `master`: Development branch.
* `release-2.x`: Release branch of a 2.x version. All 2.x.y bugfixes go in here too after merging.

## Docker Tags

* `master`: Upon building the master branch travis will build a docker images with the tag `master` and push to Docker
  Hub. This should be considered experimental and not particularly stable.
* `master-armhf`: Same as master, but built for ARM32 architecture (Raspberry Pi's, etc.).
* `master-arm64`: Same as master, but built for ARM64 architecture (also called aarch64).
* `master-amd64`: Same as master, built for x86 or AMD64 architecture.
* `latest`: The latest stable image tag from the latest release branch. If you don't know which to pick, use this.
* `latest-armhf`, `latest-arm64`, `latest-amd64`: Same schema as the master applies here. 
* `2.x`, `2.x-armhf`, `2.x-arm64`, `2.x-amd64`: Particular release branch for a specific arch. Bugfixes go here too.

# Concept

- Split video with `ffmpeg -i movie.mp4 -c copy -map 0 -segment_time 120 -f segment movie_segment_%d.mkv` on shovel node
- Transcode each segment with `ffmpeg -i movie_segment_1.mp4 -c:v copy -c:a copy movie_segment_1.mkv`
  on compute node (with whatever parameters)
- Create the concat file: `echo "file movie_segment_1.mkv" >> concat.txt` (make sure they are sorted!)
- Merge the segments back into 1 file: `ffmpeg -f concat -i concat.txt -c copy movie_out.mkv`
