#!/bin/bash

#!/bin/bash

count=0
for i in {1..200}
do
    output=$(go test -run 2A | tail -n 1)
    if [[ $output != ok* ]]; then
        echo "Failed at round $i 😥"
        count=$((count + 1))
    else
        echo "Success at round $i 😊"
    fi
done
echo "Failure rate $((100 * count / 200)) %"