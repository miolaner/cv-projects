package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	Underline = "\033[4m"
	Reset     = "\033[0m"
	Green     = "\033[32m"
	Blue      = "\033[1;34m"
	bold      = "\033[1m"
	Red       = "\x1b[31m"
)

type Traininfo struct {
	Location *string
}

type Station struct {
	Name string
	X    int
	Y    int
}

var pathFound bool
var errorsfound bool
var conflictExists bool

func Mapreader(mapfile string, start string, end string) (map[string]Station, map[string][]string) {
	if start == end {
		fmt.Fprintf(os.Stderr, "Error: Start and end stations are same (%s)\n", start)
		os.Exit(0)
	}
	file, err := os.OpenFile(mapfile, os.O_RDONLY, 0644)
	if err != nil {
		fmt.Println("error reading the map:", err)
		os.Exit(0)
	}

	pattern := `^[a-z_0-9]+$`
	match := regexp.MustCompile(pattern)
	stations := make(map[string]Station)
	connections := make(map[string][]string)
	scanner := bufio.NewScanner(file)
	occCoords := make(map[string]bool)
	section := ""
	hasConnections := false
	hasStations := false
	startExists := false
	endExists := false

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.ReplaceAll(line, " ", "")
		parts := strings.Split(line, "#")
		if len(parts) > 0 {
			line = parts[0]
		}
		if strings.HasPrefix(line, "stations:") {
			section = "stations"
			hasStations = true
			continue
		}
		if strings.HasPrefix(line, "connections:") {
			section = "connections"
			hasConnections = true
			continue
		}
		if line == "" {
			continue
		}

		if section == "stations" {
			parts := strings.Split(line, ",")
			if len(parts) == 3 {
				station := parts[0]
				if station == start {
					startExists = true
				}
				if station == end {
					endExists = true
				}
				if !match.MatchString(station) {
					fmt.Fprintf(os.Stderr, "Error: Station (%s) should be composed by only lowercase, numbers and underscore characters\n", station)
					errorsfound = true
				}
				if _, exists := stations[station]; exists {
					fmt.Fprintf(os.Stderr, "Error: Station %s defined more than once\n", station)
					errorsfound = true
				}
				if strings.Contains(parts[1], "-") || strings.Contains(parts[2], "-") {
					fmt.Fprintf(os.Stderr, "Error: Station "+station+" contains negative coordinates\n")
					errorsfound = true
				}
				x, _ := strconv.Atoi(parts[1])
				y, _ := strconv.Atoi(parts[2])
				stations[station] = Station{
					Name: station,
					X:    x,
					Y:    y,
				}
				if !occCoords[(parts[1] + " " + parts[2])] {
					occCoords[(parts[1] + " " + parts[2])] = true
				} else {
					fmt.Fprintf(os.Stderr, "Error: Station %s tried to occupy coordinates %s,%s which are already occupied\n", station, parts[1], parts[2])
					errorsfound = true
				}
				if len(stations) > 10000 {
					fmt.Fprintf(os.Stderr, "Error: Train map exceeded the maximum number(10,000) of allowed stations, exiting...\n")
					errorsfound = true
					os.Exit(0)
				}
			} else {
				fmt.Fprintf(os.Stderr, "Error: Insufficient variables for station in %s\n", parts)
				errorsfound = true
			}
		} else if section == "connections" {
			parts := strings.Split(line, "-")
			if len(parts) == 2 {
				station1 := parts[0]
				if _, exists := stations[station1]; !exists {
					fmt.Fprintf(os.Stderr, "Error: Tried to make connection to %s, which is not specified within stations section\n", station1)
					errorsfound = true
				}
				station2 := parts[1]
				if _, exists := stations[station2]; !exists {
					fmt.Fprintf(os.Stderr, "Error: Tried to make connection to %s, which is not specified within stations section\n", station2)
					errorsfound = true
				}
				if contains(connections[station1], station2) || contains(connections[station2], station1) {
					fmt.Fprintf(os.Stderr, "Error: duplicate line between "+station1+" and "+station2+"\n")
					errorsfound = true
				}
				connections[station1] = append(connections[station1], station2)
				connections[station2] = append(connections[station2], station1)
			}
		}
	}
	if !hasConnections {
		fmt.Fprintf(os.Stderr, "Error: Train map does not contain connections\n")
	}
	if !hasStations {
		fmt.Fprintf(os.Stderr, "Error: Train map does not contain stations\n")
	}
	if !startExists {
		fmt.Fprintf(os.Stderr, "Error: Start station (%s) was not found within the train map\n", start)
	}
	if !endExists {
		fmt.Fprintf(os.Stderr, "Error: End station (%s) was not found within the train map\n", end)
	}
	if !hasConnections || !hasStations || !startExists || !endExists {
		os.Exit(0)
	}
	return stations, connections
}

func contains(station []string, connection string) bool {

	for _, s := range station {
		if s == connection {
			return true
		}
	}
	return false
}

func distance(s1, s2 Station) int {
	return int(math.Sqrt(float64((s1.X-s2.X)*(s1.X-s2.X) + (s1.Y-s2.Y)*(s1.Y-s2.Y))))
}

func Dijkstra(stations map[string]Station, connections map[string][]string, start, end string, conflicts []string) []string {

	dist := make(map[string]int)
	prev := make(map[string]string)
	unvisited := make(map[string]bool)

	for station := range stations {
		dist[station] = int(^uint(0) >> 1)
		unvisited[station] = true
	}
	dist[start] = 0

	for len(unvisited) > 0 {

		var currentStation string
		smallestDistance := int(^uint(0) >> 1)
		for station := range unvisited {

			if dist[station] < smallestDistance {
				smallestDistance = dist[station]
				currentStation = station
			}
		}

		if len(connections[currentStation]) == 0 {

			if pathFound {

				return []string{}
			} else {
				fmt.Fprintf(os.Stderr, "Error: no valid path between %s and %s\n", start, end)
				os.Exit(0)
			}
		}
		if currentStation == "" && !pathFound {
			fmt.Fprintf(os.Stderr, "Error: no valid path between %s and %s\n", start, end)
			errorsfound = true
		}
		if currentStation == end {
			break
		}

		delete(unvisited, currentStation)

		for _, neighbor := range connections[currentStation] {

			if containsString(conflicts, neighbor) && currentStation != start && len(connections[neighbor]) >= 1 && neighbor != prev[currentStation] {
				alt := dist[currentStation] + 1
				if len(connections[currentStation]) == 1 && prev[currentStation] == start || len(connections[currentStation]) == 2 && prev[currentStation] != start {

					if alt < dist[neighbor] {

						dist[neighbor] = alt
						prev[neighbor] = currentStation
						continue
					} else {
						delete(stations, currentStation)
					}
				}
				if len(connections[currentStation]) > 1 {

					connections[neighbor] = sever(connections[neighbor], currentStation)
					connections[currentStation] = sever(connections[currentStation], neighbor)

					continue
				}
				if len(connections[neighbor]) > 2 {

					connections[neighbor] = sever(connections[neighbor], currentStation)
					connections[currentStation] = sever(connections[currentStation], neighbor)

					if len(connections[currentStation]) == 0 {
						delete(stations, currentStation)
					}
					continue
				}

				continue

			}
			if !unvisited[neighbor] {
				continue
			}

			alt := dist[currentStation] + 1
			if alt < dist[neighbor] {
				dist[neighbor] = alt
				prev[neighbor] = currentStation
			}

		}
	}

	path := []string{}
	for u := end; u != ""; u = prev[u] {
		path = append([]string{u}, path...)
		if u == start {
			pathFound = true
			break
		}
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	return path
}

func containsString(slice []string, str string) bool {

	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func containsStringInSlices(slice [][]string, str string) bool {
	for _, innerSlice := range slice {
		if containsString(innerSlice, str) {
			return true
		}
	}
	return false
}

func sever(connection []string, cut string) []string {

	var result []string
	for _, station := range connection {
		if station != cut {
			result = append(result, station)
		}
	}
	return result
}

func main() {
	if len(os.Args) != 5 {
		fmt.Fprintf(os.Stderr, "Error: incorrect number of arguments (%d), should be 5\n", len(os.Args))
		fmt.Println(Green, " To run the tool:")
		fmt.Println("  go run . <path to file containing network map> <start station> <end station> <numeric amount of trains>", Reset)
		os.Exit(0)
	}

	start := os.Args[2]
	end := os.Args[3]
	if strings.HasPrefix(os.Args[4], "-") {
		fmt.Fprintf(os.Stderr, "Error: train value(%s) negative\n", os.Args[4])
		errorsfound = true
	}
	traincount, err := strconv.Atoi(os.Args[4])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: unable to convert train numbers(%s) to integers\n", os.Args[4])
		os.Exit(0)
	}

	//Mapreader reads the map, checks most error scenarios and returns two mapy, on contains stations and coordinates
	//and other stations and their connections.
	stations, connections := Mapreader(os.Args[1], start, end)
	startcon := connections[start]
	var conflicts []string
	paths, conflicts := pathPlanner(connections, stations, start, end, nil)

	count := 1
	for conflictExists {
		count++
		connections[start] = startcon
		paths, conflicts = pathPlanner(connections, stations, start, end, conflicts)
		if len(conflicts) == 0 {
			conflictExists = false
		}
	}

	//Trainnames simply creates a map which is used to separate trains from others and hold current location
	trains := Trainnames(traincount, start)

	if !errorsfound {
		Pathbuilder(trains, paths, stations, start)
	} else {
		fmt.Println(Red, "Please fix listed errors", Reset)
	}

}

func newLocation(location string) *string {
	loc := location
	return &loc
}

func pathPlanner(connections map[string][]string, stations map[string]Station, start, end string, conflicts []string) ([][]string, []string) {
	//Dijkstra calculates distances between stations and returns viable paths from start to end
	var paths [][]string
	connections2 := connections
	for len(stations) > 0 {

		path := Dijkstra(stations, connections2, start, end, conflicts)

		if len(path) == 0 {
			break
		}

		connections2[start] = sever(connections2[start], path[len(path)-2])
		connections2[path[len(path)-2]] = sever(connections2[path[len(path)-2]], start)

		paths = append(paths, path)

		conflicts = findConflicts(paths, start, end)
		if len(conflicts) > 0 {
			conflictExists = true
		}

	}
	return paths, conflicts
}

func Trainnames(n int, start string) map[string]*Traininfo {
	traininfo := make(map[string]*Traininfo)
	for i := 1; i <= n; i++ {
		trainName := "T" + strconv.Itoa(i) + "-"
		traininfo[trainName] = &Traininfo{
			Location: newLocation(start),
		}
	}
	return traininfo
}

func findConflicts(slice [][]string, start, end string) []string {
	occurrence := make(map[string]int)

	// Count the occurrences of each string across all slices
	for _, innerSlice := range slice {
		uniqueStrings := make(map[string]bool)
		for _, str := range innerSlice {
			uniqueStrings[str] = true
		}
		for str := range uniqueStrings {
			occurrence[str]++
		}
	}

	// Collect strings that are in conflict (appear in more than one slice)
	var conflicts []string
	for str, count := range occurrence {
		if count > 1 && str != start && str != end {
			conflicts = append(conflicts, str)
		}
	}

	return conflicts
}

func Pathbuilder(trains map[string]*Traininfo, paths [][]string, stations map[string]Station, start string) {

	occupied := make(map[string]bool)

	shortestpath := int(^uint(0) >> 1)
	for _, path := range paths {
		for _, station := range path[:len(path)-1] {
			occupied[station] = false
			if len(path) < shortestpath {
				shortestpath = len(path)
			}
		}
	}

	count := len(trains)
	departed := len(trains)
	var turn string
	occupied[start] = true
	severed := false

	for count > 0 {
		for _, path := range paths {
			for i, station := range path {
				if occupied[station] {
					for name, train := range trains {
						if station == *train.Location {
							if path[i-1] == path[0] {
								if station == start {
									if !severed {
										turn += name + path[i-1] + " "
										delete(trains, name)
										count--
										departed--
										severed = true
										continue
									} else {
										continue
									}
								}
								turn += name + path[i-1] + " "
								occupied[*train.Location] = false
								count--
								delete(trains, name)
							} else if !occupied[path[i-1]] {
								if station == start && (len(path)-shortestpath) > departed {
									continue
								}
								occupied[*train.Location] = false
								*train.Location = path[i-1]
								turn += name + *train.Location + " "
								occupied[*train.Location] = true
								if station == start {
									departed--
									if departed > 0 {
										occupied[start] = true
										continue
									}
								}
							}
							break
						}
					}
				} else {
					continue
				}
			}
		}
		fmt.Println(Blue, turn, Reset)
		severed = false
		turn = ""
	}
}
