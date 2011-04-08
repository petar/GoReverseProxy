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

## Maturity

I am running GoReverseProxy in production (albeit a small production) in front
of my blog [Population Algorithms](http://popalg.org) and it has been working fine.
The blog requests are generally pretty heavy (since they pull in a lot of resource
files and things). Keepalive and pipelining have been working correctly.

Nevertheless, it is still early to say that GoReverseProxy is truly production-ready.

## Installation

To install, simply run

	git clone git://github.com/petar/GoReverseProxy.git GoReverseProxy-git
	cd GoReverseProxy-git
	make
	make install

There is an example config file in the subdirectory {reverseproxy} which is
simple and self-explanatory.

## About

GoReverseProxy is maintained by [Petar Maymounkov](http://pdos.csail.mit.edu/~petar/). 
