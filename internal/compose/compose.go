package compose

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type File struct {
	Services map[string]Service `yaml:"services"`
}

type Service struct {
	Ports []Port
}

type Port struct {
	Service   string
	Raw       string
	HostIP    string
	Published int
	Target    int
}

func Load(path string) (File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return File{}, err
	}
	return Parse(data)
}

func Parse(data []byte) (File, error) {
	var raw struct {
		Services map[string]struct {
			Ports []yaml.Node `yaml:"ports"`
		} `yaml:"services"`
	}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return File{}, err
	}
	file := File{Services: map[string]Service{}}
	for name, svc := range raw.Services {
		service := Service{}
		for _, node := range svc.Ports {
			port, ok := parsePortNode(name, node)
			if ok {
				service.Ports = append(service.Ports, port)
			}
		}
		file.Services[name] = service
	}
	return file, nil
}

func parsePortNode(service string, node yaml.Node) (Port, bool) {
	switch node.Kind {
	case yaml.ScalarNode:
		return parsePortString(service, node.Value)
	case yaml.MappingNode:
		return parsePortMapping(service, node)
	default:
		return Port{}, false
	}
}

func parsePortString(service, raw string) (Port, bool) {
	parts := strings.Split(raw, ":")
	if len(parts) < 2 {
		return Port{Service: service, Raw: raw}, false
	}
	target, err := strconv.Atoi(stripProtocol(parts[len(parts)-1]))
	if err != nil {
		return Port{Service: service, Raw: raw}, false
	}
	published, err := strconv.Atoi(parts[len(parts)-2])
	if err != nil {
		return Port{Service: service, Raw: raw}, false
	}
	port := Port{Service: service, Raw: raw, Published: published, Target: target}
	if len(parts) == 3 {
		port.HostIP = parts[0]
	}
	return port, true
}

func parsePortMapping(service string, node yaml.Node) (Port, bool) {
	port := Port{Service: service}
	for i := 0; i+1 < len(node.Content); i += 2 {
		key := node.Content[i].Value
		value := node.Content[i+1].Value
		switch key {
		case "host_ip":
			port.HostIP = value
		case "published":
			port.Published, _ = strconv.Atoi(value)
		case "target":
			port.Target, _ = strconv.Atoi(value)
		}
	}
	port.Raw = fmt.Sprintf("%d:%d", port.Published, port.Target)
	return port, port.Published > 0
}

func stripProtocol(value string) string {
	if port, _, ok := strings.Cut(value, "/"); ok {
		return port
	}
	return value
}
