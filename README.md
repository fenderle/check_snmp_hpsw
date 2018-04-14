# check_snmp_hpsw

check_snmp_hpsw is a nagios plugin to check a HP Procurve/Aruba switches for
errors using SNMP queries. The plugin is build in golang to make it dependency
free and easy to deploy.

## Getting started

These instructions will get you a copy of the project up and running on your
local machine for development and testing purposes. See deployment for notes
on how to deploy the project on a live system.

### Prerequisites

The project is developed using Visual Studio Code and thus provides settings
for VS Code. You can choose to use them or not.

The project uses vendoring with `glide` and of course requires a recent version of Go.

### Installing

```bash
go get https://github.com/fenderle/check_snmp_hpsw
```

You will find the compiled binary in your `$GOPATH/bin` folder.
