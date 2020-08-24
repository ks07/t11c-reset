# t11c-reset

This command line tool provides a quick and easy way to force reconnect the
modem on the Zyxel AMG1302-T11C ADSL wireless router. All interaction is
carried out by forging HTTP requests made by the configuration UI.

## Usage

Full usage instructions are available by running:

```
t11c-reset help
```

To immediately disconnect and reconnect the modem:

```
t11c-reset reconnect --username=admin --password=hunter2 --hostname=192.168.1.1
```

## Configuration

Login credentials and the hostname may be provided via a YAML configuration
file. The tool will attempt to automatically load from `$HOME/.t11c-reset.yaml`,
if it exists.

```
username: AzureDiamond
password: hunter2
hostname: 192.168.1.1
```
