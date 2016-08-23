# WinRM for Go

This is a Go library (and command-line executable) to execute remote commands on Windows machines through
the use of WinRM/WinRS.

_Note_: this library doesn't support domain users (it doesn't support GSSAPI nor Kerberos). It's primary target is to execute remote commands on EC2 windows machines.

[![Build Status](https://travis-ci.org/masterzen/winrm.svg?branch=master)](https://travis-ci.org/masterzen/winrm)
[![Coverage Status](https://coveralls.io/repos/masterzen/winrm/badge.png)](https://coveralls.io/r/masterzen/winrm)

## Contact

- Bugs: https://github.com/masterzen/winrm/issues


## Getting Started
WinRM is available on Windows Server 2008 and up. This project natively supports basic authentication for local accounts, see the steps in the next section on how to prepare the remote Windows machine for this scenario. The authentication model is pluggable, see below for an example on using Negotiate/NTLM authentication (e.g. for connecting to vanilla Azure VMs).

### Preparing the remote Windows machine for Basic authentication
This project supports only basic authentication for local accounts (domain users are not supported). The remote windows system must be prepared for winrm:

_For a PowerShell script to do what is described below in one go, check [Richard Downer's blog](http://www.frontiertown.co.uk/2011/12/overthere-control-windows-from-java/)_

On the remote host, open a Command Prompt (not a PowerShell prompt!) using the __Run as Administrator__ option and paste in the following lines:

		winrm quickconfig
		y
		winrm set winrm/config/service/Auth @{Basic="true"}
		winrm set winrm/config/service @{AllowUnencrypted="true"}
		winrm set winrm/config/winrs @{MaxMemoryPerShellMB="1024"}

__N.B.:__ The Windows Firewall needs to be running to run this command. See [Microsoft Knowledge Base article #2004640](http://support.microsoft.com/kb/2004640).

__N.B.:__ Do not disable Negotiate authentication as the `winrm` command itself uses this for internal authentication, and you risk getting a system where `winrm` doesn't work anymore.

__N.B.:__ The `MaxMemoryPerShellMB` option has no effects on some Windows 2008R2 systems because of a WinRM bug. Make sure to install the hotfix described [Microsoft Knowledge Base article #2842230](http://support.microsoft.com/kb/2842230) if you need to run commands that uses more than 150MB of memory.

For more information on WinRM, please refer to <a href="http://msdn.microsoft.com/en-us/library/windows/desktop/aa384426(v=vs.85).aspx">the online documentation at Microsoft's DevCenter</a>.

### Building the winrm go and executable

You can build winrm from source:

```sh
git clone https://github.com/masterzen/winrm
cd winrm
make
```

This will generate a binary in the base directory called `./winrm`.

_Note_: this winrm code doesn't depend anymore on [Gokogiri](https://github.com/moovweb/gokogiri) which means it is now in pure Go.

_Note_: you need go 1.1+. Please check your installation with

```
go version
```

## Command-line usage

Once built, you can run remote commands like this:

```sh
./winrm -hostname remote.domain.com -username "Administrator" -password "secret" "ipconfig /all"
```

## Library Usage

**Warning the API might be subject to change.**

For the fast version (this doesn't allow to send input to the command):

```go
import "github.com/masterzen/winrm/winrm"

client := winrm.NewClient("localhost", "Administrator", "secret")
client.Run("ipconfig /all", os.Stdout, os.Stderr)
```

or
```go
import (
  "github.com/masterzen/winrm/winrm"
  "fmt"
  "os"
)

client, err := winrm.NewClient(&winrm.Endpoint{Host: "localhost", Port: 5985, HTTPS: false, Insecure: false}, "Administrator", "secret")
if err != nil {
	fmt.Println(err)
}

run, err := client.RunWithInput("ipconfig /all", os.Stdout, os.Stderr, os.Stdin)
if err != nil {
	fmt.Println(err)
}

fmt.Println(run)
```

For a more complex example, it is possible to call the various functions directly:

```go
import (
  "github.com/masterzen/winrm/winrm"
  "fmt"
  "bytes"
  "os"
)

stdin := bytes.NewBufferString("ipconfig /all")

client := winrm.NewClient(&winrm.Endpoint{Host: "localhost", Port: 5985, HTTPS: false, Insecure: false}, "Administrator", "secret")
shell, err := client.CreateShell()
if err != nil {
  fmt.Printf("Impossible to create shell %s\n", err)
  os.Exit(1)
}
var cmd *Command
cmd, err = shell.Execute("cmd.exe")
if err != nil {
  fmt.Printf("Impossible to create Command %s\n", err)
  os.Exit(1)
}

go io.Copy(cmd.Stdin, &stdin)
go io.Copy(os.Stdout, cmd.Stdout)
go io.Copy(os.Stderr, cmd.Stderr)

cmd.Wait()
shell.Close()
```

### Pluggable authentication example: Negotiate/NTLM authentication
Using the winrm.Parameters.TransportDecorator, it is possible to wrap or completely replace the outgoing http.RoundTripper. For example, use github.com/paulmey/go-ntlmssp to plug in Negotiate/NTLM authentication :

```go
import (
  "github.com/masterzen/winrm/winrm"
  "github.com/paulmey/go-ntlmssp"
)

params := winrm.DefaultParameters()
params.TransportDecorator = func(t *http.Transport) http.RoundTripper { return ntlmssp.Negotiator{t} }
client, err := winrm.NewClientWithParameters(..., params)
```

## Developing on WinRM

If you wish to work on `winrm` itself, you'll first need [Go](http://golang.org)
installed (version 1.1+ is _required_). Make sure you have Go properly installed,
including setting up your [GOPATH](http://golang.org/doc/code.html#GOPATH).

For some additional dependencies, Go needs [Mercurial](http://mercurial.selenic.com/)
and [Bazaar](http://bazaar.canonical.com/en/) to be installed.
Winrm itself doesn't require these, but a dependency of a dependency does.

Next, clone this repository into `$GOPATH/src/github.com/masterzen/winrm` and
then just type `make`. In a few moments, you'll have a working `winrm` executable:

```
$ make
...
$ bin/winrm
...
```
You can run tests by typing `make test`.

If you make any changes to the code, run `make format` in order to automatically
format the code according to Go standards.

When new dependencies are added to winrm you can use `make updatedeps` to
get the latest and subsequently use `make` to compile and generate the `winrm` binary.