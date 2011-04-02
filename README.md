## Description

GoFrontline is an HTTP proxy server (similar to a simpler version of NGINX)
whose primary purpose currently is to multiplex incoming HTTP requests to 
different backend servers according to virtual host, and to safeguard
against malicious HTTP requests (GoFrontline will break before allowing
a backend server to break).

## Features

* Pipelining 
* Keepalive connections
* Multiplexing by virtual hosts, specified in a config file

## Installation

To install, simply run

	make
	make install

## About

GoFrontline is maintained by [Petar Maymounkov](http://pdos.csail.mit.edu/~petar/). 
