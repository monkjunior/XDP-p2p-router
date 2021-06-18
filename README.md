# A router to localize p2p application traffic

## Install dependencies

- Install golang v1.16.4

- Install gcc
```bash
# Ubuntu
$ sudo apt-get instlal gcc

```  

- [Install bcc](https://github.com/iovisor/bcc/blob/master/INSTALL.md#arch---binary)
```bash
# Install bcc on ubuntu
$ sudo apt-get install bpfcc-tools linux-headers-$(uname -r)
$ sudo apt-key adv --keyserver keyserver.ubuntu.com --recv-keys 4052245BD4284CDD
$ echo "deb https://repo.iovisor.org/apt/bionic bionic main" | sudo tee /etc/apt/sources.list.d/iovisor.list
$ sudo apt-get update
$ sudo apt-get install bcc-tools libbcc-examples linux-headers-$(uname -r)

# Install bcc on arch linux
$ pacman -S bcc bcc-tools python-bcc
```

- Download geolite2 database

  Login to [Maxmind account](https://www.maxmind.com/) to have access to download geolite2 database.
```bash
$ tar -zxvf GeoLite2-ASN.tar.gz -C GeoLite2-ASN
$ tar -zxvf GeoLite2-City.tar.gz -C GeoLite2-City
$ tar -zxvf GeoLite2-Country.tar.gz -C GeoLite2-Country
```

## Usage
```markdown
go install

XDP-p2p-router -h

XDP-p2p-router start --device <your-network-device>
```

## Libraries to build UI and charts

- Public apis for getting public ip

  - [ipify](https://api.ipify.org)
  
  - [ip-api](http://ip-api.com/json/)

- [termui](https://github.com/gizak/termui)

  - [wiki](https://github.com/gizak/termui/wiki)
 
- [go-echart](https://github.com/go-echarts/go-echarts)

# References

- Concurrency patterns in Go

  - [Concurrency patterns in Go](https://youtu.be/YEKjSzIwAdA)

  - [Pipelines and cancellation](https://blog.golang.org/pipelines)