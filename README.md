## Description

GoFrontline is an HTTP proxy server (similar to a mini-version of NGINX)
whose primary purpose currently is to multiplex incoming HTTP requests to 
different backend servers according to virtual host.

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
