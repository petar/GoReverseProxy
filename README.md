## Description

GoReverseProxy is an HTTP reverse proxy (similar to a simpler version of NGINX)
whose primary purpose currently is to multiplex incoming HTTP requests to 
different backend servers according to virtual host, and to safeguard
against malicious HTTP requests (GoReverseProxy will likely break before allowing
a backend server to break).

## Features

* Pipelining 
* Keepalive connections
* Multiplexing by virtual hosts, specified in a config file
* File-descriptor limiting
* Connection timeouts

## Installation

To install, simply run

	make
	make install

## About

GoReverseProxy is maintained by [Petar Maymounkov](http://pdos.csail.mit.edu/~petar/). 
