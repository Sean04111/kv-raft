#!/bin/bash

#!/bin/bash

count=0
for i in {1..100}
do
    output=$(go test -run 2A | tail -n 1)
    if [[ $output != ok* ]]; then
        echo "Failed at round $i 😥"
        break
    else
        echo "Success at round $i 😊"
    fi
done