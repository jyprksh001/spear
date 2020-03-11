package config

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

//GetSections returns a list of sections with the given name
func (config *Configuration) GetSections(name string) []*Section {
	sections := []*Section{}
	for _, v := range *config {
		if v.Name == name {
			sections = append(sections, v)
		}
	}
	return sections
}

//ParseFile parse a spear config file to Configuration
func ParseFile(path string) (*Configuration, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	p := praser{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if err := p.readLine(scanner.Text()); err != nil {
			return nil, err
		}
	}
	p.export("")

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return p.config, nil
}

type praser struct {
	currentSection string
	buffer         map[string]string
	config         *Configuration
}

func (p *praser) readLine(line string) error {
	line = strings.TrimSpace(line)
	if len(line) == 0 {
		return nil
	}

	if line[0] == '[' && line[len(line)-1] == ']' {
		p.export(line[1 : len(line)-1])
		return nil
	}

	if p.config == nil {
		return errors.New("Line: " + line + " is not under any section.")
	}

	pair := strings.SplitN(line, "=", 2)
	if len(pair) != 2 {
		return errors.New("Line: " + line + " is not a valid key=value pair.")
	}

	key := strings.ToLower(strings.TrimSpace(pair[0]))
	value := strings.TrimSpace(pair[1])

	if _, ok := p.buffer[key]; ok {
		return errors.New("Section " + p.currentSection + " already has value for key " + key)
	}

	p.buffer[key] = value
	return nil
}

func (p *praser) export(newSection string) {
	oldSection := p.currentSection
	p.currentSection = newSection
	if p.config != nil {
		//Export section to config
		newConfig := append(*p.config, &Section{
			Name:    strings.ToLower(oldSection),
			Content: p.buffer,
		})
		p.config = &newConfig
	} else {
		//Create config
		p.config = &Configuration{}
	}

	p.buffer = map[string]string{}
}

//ParseAddr turns a string in ipv4:port format to a UDPAddr
func ParseAddr(str string) (*net.UDPAddr, error) {
	result := strings.Split(str, ":")

	var ipstr, portstr string
	if len(result) == 2 {
		ipstr = result[0]
		portstr = result[1]
	} else {
		return nil, fmt.Errorf("Failed to parse Address: %#v, length of split result is %d", str, len(result))
	}

	ip := net.ParseIP(ipstr)
	if ip == nil {
		return nil, fmt.Errorf("Failed to parse IP: %#v", ipstr)
	}

	port, err := strconv.Atoi(portstr)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse Port: %#v", portstr)
	}

	return &net.UDPAddr{IP: ip, Port: port}, nil
}
