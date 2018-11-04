# clustercode-worker
Worker microservice for clustercode [WIP]

## Building

    go mod vendor

## Running

    go build main.go
    ./clustercode-worker

## Configuring for local development

- Copy `defaults.yaml` to `config.yaml`
- Create the directories `input`, `output` in the current working dir.
- Change the input and output dirs in `config.yaml`
- Delete other values in `config.yaml` that are already default, leaving the custom ones left.

## Testing

not sure yet
