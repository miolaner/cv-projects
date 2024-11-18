#!/bin/bash

# Define an array of commands
commands=(
    "go run main.go maps/dubRoutes.txt waterloo st_pancras 4"
    "go run main.go maps/london.txt waterloo st_pancras 3"
    "go run main.go maps/london.txt waterloo st_pancras testi 3"
    "go run main.go maps/noConnect.txt waterloo st_pancras 4"
    "go run main.go maps/beet.txt beethoven part 9"
    "go run main.go maps/dubNames.txt waterloo st_pancras 4"
    "go run main.go maps/london.txt waterloo 4"
    "go run main.go maps/noWater.txt waterloo st_pancras 4"
    "go run main.go maps/noSaint.txt waterloo st_pancras 4"
    "go run main.go maps/madeupConnect.txt waterloo st_pancras 4"
    "go run main.go maps/noStation.txt waterloo st_pancras 4"
    "go run main.go maps/negativeCoo.txt waterloo st_pancras 4"
    "go run main.go maps/sizes.txt small large 9"
    "go run main.go maps/numbers.txt two four 4"
    "go run main.go maps/jungle.txt jungle desert 10"
    "go run main.go maps/london.txt waterloo st_pancras -4"
    "go run main.go maps/london.txt waterloo st_pancras 100"
    "go run main.go maps/madeupName.txt waterloo st_pancras 4"
    "go run main.go maps/bond.txt bond_square space_port 4"
    "go run main.go maps/sameCoo.txt waterloo st_pancras 4"
    "go run main.go maps/alpha.txt alpha zeta 60"
    "go run main.go maps/nu.txt alpha nu 70"
    "go run main.go maps/london.txt waterloo st_pancras 2"
    "go run main.go maps/begi.txt beginning terminus 20"
    "go run main.go maps/london.txt waterloo st_pancras 1"
    "go run main.go maps/london.txt waterloo st_pancras 4"
    "go run main.go maps/london.txt waterloo waterloo 4"
    "go run main.go maps/noPath.txt waterloo st_pancras 4"
    "go run maps/test_large_map.go"
)

# Loop through the commands and execute each one
for cmd in "${commands[@]}"; do
    echo "Executing: $cmd"
    $cmd
    if [ $? -ne 0 ]; then
        echo "Command failed: $cmd"
        exit 1
    fi
done

echo "All commands executed successfully."