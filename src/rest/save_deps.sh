#/bin/bash

docker run --rm -v $PWD:/go/src/app src_rest /bin/bash -c "go-wrapper download && godep save -v"

