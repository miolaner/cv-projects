# Train Network Pathfinding Tool
### Overview
This program is designed to read a map of train stations and connections, then find and display paths from a specified start station to an end station. It implements Dijkstra's algorithm to calculate the shortest paths and moves a specified number of trains along these paths.

### Features
##### Map Reading: Parses a map file containing station coordinates and connections.
##### Error Handling: Detects and reports various errors in the map file (e.g., duplicate stations, invalid connections, negative coordinates).
##### Pathfinding: Uses Dijkstra's algorithm to find the shortest paths between stations.
##### Train Movement: Simulates moving trains along the computed paths.
##### Command Line Interface: Run the program with command-line arguments to specify the map file, start station, end station, and number of trains.
### Usage
##### Prerequisites
Go programming language installed on your machine.
##### Running the Program
To run the program, use the following command:
###### "go run main.go <path_to_map_file> <start_station> <end_station> <number_of_trains>"
##### Example:
###### "go run main.go maps/london.txt waterloo st_pancras 4"

This command reads the map from maps/london.txt, finds paths from waterloo to st_pancras, and simulates moving 4 trains along these paths.


#### Command Line Arguments
* <path_to_map_file>: Path to the map file containing station data and connections.
* <start_station>: Name of the starting station.
* <end_station>: Name of the ending station.
* <number_of_trains>: Number of trains to move from the start station to the end station.


#### Stations Section
Begins with stations: on a new line.
Each station is defined by a line containing the station name, x-coordinate, and y-coordinate separated by commas.
##### Example:
stations:
waterloo,1,2
st_pancras,4,6

#### Connections Section
Begins with connections: on a new line.
Each connection is defined by a line containing two station names separated by a hyphen.
##### Example:
connections:
waterloo-st_pancras

### Testing

A bash script is provided to run multiple test cases.
make it executable, and run it: 
###### "chmod +x run_tests.sh"
###### "./run_tests.sh"
 
### Key Functions

##### Mapreader(mapfile string, start string, end string): Reads the map file and returns station and connection data.
##### Dijkstra(stations map[string]Station, connections map[string][]string, start, end string): Implements Dijkstra's algorithm to find paths.
##### Pathbuilder(trains map[string]*Traininfo, paths [][]string, stations map[string]Station, start string): Simulates moving trains along the paths.

### Error Handling

The program includes extensive error checking for various potential issues, such as:

##### Start and end stations being the same.
##### Invalid station names or coordinates.
##### Duplicate stations or connections.
##### Exceeding the maximum number of allowed stations (10,000).


### Output
The program prints the paths and train movements to the standard output, using colored text for better readability.