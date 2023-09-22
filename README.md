# Load Balancer with multiple algorithms

## Description

Load Balancer written in Golang, with logrus to log every request information, active & passive backend URL check

## Getting Started

### Dependencies

* logrus

### Installing

* Clone this project
* Run command *make build-and-run*
* Run binary file

### Executing program

* Run command in Makefile
```
make build-and-run
```

## Help

* Create folder logfile/
* Create config.json from example.config.json

## Authors

Trần Sỹ Bảo
...

## Version History
* 0.5
    * Bug fix
    * Add LVS RR
    * Add Smooth RR
* 0.3
    * Bug fix
    * Improve Round Robin algorithm
    * Restructure algo/ folder for more algorithm implement
    * Check isAlive for 1 second and redirect request to another alive server.
* 0.1
    * Initial Release with default Round Robin


## License

This project is licensed under the [NAME HERE] License - see the LICENSE.md file for details