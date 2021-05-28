# A router to localize p2p application traffic

## Install dependencies

- [Install bcc](https://github.com/iovisor/bcc/blob/master/INSTALL.md#arch---binary)
```bash
#Install bcc on arch linux
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

- [termui](https://github.com/gizak/termui)

- [go-echart](https://github.com/go-echarts/go-echarts)