  #!/bin/bash

docker build -t drmaa/wflexample:latest .
docker run --rm -it drmaa/wflexample:latest