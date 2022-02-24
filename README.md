# Noisy

A simple python script that generates random HTTP/DNS traffic noise in the background while you go about your regular
web browsing, to make your web traffic data less valuable for selling and for extra obscurity.

Tested on MacOS High Sierra, Ubuntu 16.04 and Raspbian Stretch and is compatable with both Python 2.7 and 3.6

## Getting Started

These instructions will get you a copy of the project up and running on your local machine

### Dependencies

Install `requests` if you do not have it already installed, using `pip`:

```shell
pip install requests
```

### Usage

Clone the repository

```shell
git clone https://github.com/1tayH/noisy.git
```

Navigate into the `noisy` directory

```shell
cd noisy
```

Run the script

```shell
python noisy.py --config config.json
```

The program can accept a number of command line arguments:

```shell
$ python noisy.py --help
usage: noisy.py [-h] [--log -l] --config -c [--timeout -t]

optional arguments:
  -h, --help    show this help message and exit
  --log -l      logging level
  --config -c   config file
  --timeout -t  for how long the crawler should be running, in seconds
```

only the config file argument is required.

### Output

```shell
$ docker run -it noisy --config config.json --log debug
DEBUG:urllib3.connectionpool:Starting new HTTP connection (1): 4chan.org:80
DEBUG:urllib3.connectionpool:http://4chan.org:80 "GET / HTTP/1.1" 301 None
DEBUG:urllib3.connectionpool:Starting new HTTP connection (1): www.4chan.org:80
DEBUG:urllib3.connectionpool:http://www.4chan.org:80 "GET / HTTP/1.1" 200 None
DEBUG:root:found 92 links
INFO:root:Visiting http://boards.4chan.org/s4s/
DEBUG:urllib3.connectionpool:Starting new HTTP connection (1): boards.4chan.org:80
DEBUG:urllib3.connectionpool:http://boards.4chan.org:80 "GET /s4s/ HTTP/1.1" 200 None
INFO:root:Visiting http://boards.4chan.org/s4s/thread/6850193#p6850345
DEBUG:urllib3.connectionpool:Starting new HTTP connection (1): boards.4chan.org:80
DEBUG:urllib3.connectionpool:http://boards.4chan.org:80 "GET /s4s/thread/6850193 HTTP/1.1" 200 None
INFO:root:Visiting http://boards.4chan.org/o/
DEBUG:urllib3.connectionpool:Starting new HTTP connection (1): boards.4chan.org:80
DEBUG:urllib3.connectionpool:http://boards.4chan.org:80 "GET /o/ HTTP/1.1" 200 None
DEBUG:root:Hit a dead end, moving to the next root URL
DEBUG:urllib3.connectionpool:Starting new HTTPS connection (1): www.reddit.com:443
DEBUG:urllib3.connectionpool:https://www.reddit.com:443 "GET / HTTP/1.1" 200 None
DEBUG:root:found 237 links
INFO:root:Visiting https://www.reddit.com/user/Saditon
DEBUG:urllib3.connectionpool:Starting new HTTPS connection (1): www.reddit.com:443
DEBUG:urllib3.connectionpool:https://www.reddit.com:443 "GET /user/Saditon HTTP/1.1" 200 None
...
```

## Build Using Docker

### 1. Build the image

`docker build -t noisy .`

**Or** if you'd like to build it for a **Raspberry Pi** (running Raspbian stretch):

`docker build -f Dockerfile.pi -t noisy .`

### 2. Create the container and run:

`docker run -it noisy --config config.json`

## Run multiple containers using `docker-compose`

`docker-compose` is useful if you want to run more than one container at the same time, to generate more noise. To do
so, simply run the following commands:

```shell
cd docker-compose
docker-compose build
docker-compose up --scale noisy=<number-of-containers>
```

## Set noisy to run automatically via systemd

You can use systemd to start noisy.py automatically on every boot. The provided
example service assumes that you have the script copied to /opt/noisy and that
noisy.py and config.json are readable by the 'noisy' user. You can change these
values to suit your needs.

To configure the service:

```shell
sudo cp examples/systemd/noisy.service /etc/systemd/system
sudo systemctl daemon-reload
sudo systemctl enable noisy && sudo systemctl start noisy
```

You can view the script's output by running:

```shell
journalctl -f -n noisy
```

## Authors

* **Itay Hury** - *Initial work* - [1tayH](https://github.com/1tayH)
* **Michael Savin** - *Personal customization* - [jtprogru](https://github.com/jtprogru)

See also the list of [contributors](https://github.com/1tayH/Noisy/contributors) who participated in this project.

## License

This project is licensed under the GNU GPLv3 License - see the [LICENSE.md](LICENSE.md) file for details

## Acknowledgments

This project has been inspired by
* [RandomNoise](http://www.randomnoise.us)
* [web-traffic-generator](https://github.com/ecapuano/web-traffic-generator)

