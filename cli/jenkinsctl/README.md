**jenkinsctl** is a jenkins CLI based on [gojenkins](https://github.com/bndr/gojenkins) library. 🚀

:one: Generate a token for the username that will manage the jenkins.

- Log in to Jenkins.
- Click you name (upper-right corner).
- Click Configure (left-side menu).
- Use "Add new Token" button to generate a new one then name it.
- You must copy the token when you generate it as you cannot view the token afterwards.

:two: Create the `configuration directory` and the `config.json file`
```
$ mkdir -p ~/.config/jenkinsctl/
$ pushd ~/.config/jenkinsctl/
    $ vi config.json 
    {
        "Server": "https://jenkins.mydomain.com",
        "JenkinsUser": "jenkins-operator",
        "Token": "1152e8e7a88f6c7ef605844b35t5y6i"
    }
$ popd
```

:three: Build the jenkinsctl

```
$ git clone https://github.com/dougsland/jenkinsctl.git
$ cd jenkinsctl
$ make
```

```
$ ./jenkinsctl
Client for jenkins, manage resources by the jenkins

Usage:
  jenkinsctl [command]

Available Commands:
  create      Create a resource in Jenkins
  delete      Delete a resource from Jenkins
  disable     Disable a resource in Jenkins
  download    download related commands
  enable      Enable a resource in Jenkins
  get         Get a resource from Jenkins
  help        Help about any command
  plugins     Commands related to plugins

Flags:
      --config string   Path to config file
  -h, --help            help for jenkinsctl
  -v, --version         version for jenkinsctl

Use "jenkinsctl [command] --help" for more information about a command.
```

:rocket: :rocket: :rocket: :rocket:
