# t11c-reset

This command line tool provides a quick and easy way to force reconnect the
modem on the Zyxel AMG1302-T11C ADSL wireless router. All interaction is
carried out by forging HTTP requests made by the configuration UI.

## Usage

Full usage instructions are available by running:

```sh
t11c-reset help
```

To immediately disconnect and reconnect the modem:

```sh
t11c-reset reconnect --username=admin --password=hunter2 --hostname=192.168.1.1
```

## Configuration

Login credentials and the hostname may be provided via a YAML configuration
file. The tool will attempt to automatically load from `$HOME/.t11c-reset.yaml`,
if it exists.

```yaml
username: AzureDiamond
password: hunter2
hostname: 192.168.1.1
```

## Build

This project uses Go Modules, so compilation is extremely straightforward.
Go `1.15` or a compatible release is required.

```sh
go build # By default will create the executable t11c-reset
```

## SystemD Usage

To run as a service on a Linux machine, you can use the provided SystemD
unit file in the `systemd` directory (supported by the majority of recent
mainstream distributions).

To install, first install the built executable somewhere sensible on the
machine, e.g.

```sh
sudo cp t11c-reset /usr/local/bin/t11c-reset
```

Then install the systemd unit file:

```sh
sudo cp systemd/t11c-reset.service /etc/systemd/system/t11c-reset.service
sudo chmod 644 /etc/systemd/system/t11c-reset.service
```

You should consider running the service [as a non-privileged user](https://www.freedesktop.org/software/systemd/man/systemd.exec.html#User=).

And create a configuration file with the login details (or edit the unit file
to configure via environment/command line):

```sh
# Create this file by hand. See the "Configuration" section above for details
sudo vim /usr/local/etc/t11c-reset.yaml
```

Finally, enable the service on boot, and start it:

```sh
sudo systemctl enable t11c-reset
sudo systemctl start t11c-reset
```

Once running, you can check the logs using `journalctl`:

```sh
journalctl -u t11c-reset -f
```
