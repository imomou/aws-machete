# aws-machete
Like a swiss army knife, but with more oomph

## Build Status
This software is currently in beta / pre-rc. Not "Google beta" kind of beta but "still testing" kind of beta! Use at your own risk!

[![concourse.bitstormy.io](https://concourse.bitstormy.io/api/v1/pipelines/aws-machete/jobs/build-cloudformation/badge)](https://concourse.bitstormy.io/teams/main/pipelines/aws-machete) s3 util

[![concourse.bitstormy.io](https://concourse.bitstormy.io/api/v1/pipelines/aws-machete/jobs/build-s3/badge)](https://concourse.bitstormy.io/teams/main/pipelines/aws-machete) cloudformation util

[![concourse.bitstormy.io](https://concourse.bitstormy.io/api/v1/pipelines/aws-machete/jobs/docker-build/badge)](https://concourse.bitstormy.io/teams/main/pipelines/aws-machete) docker pack

## Usage

Put the following file named "aws-machete" in /usr/bin (or anywhere in $PATH)

~~~
#!/bin/bash

docker run -it --rm -e AWS_REGION=$AWS_REGION \
    -v ~/.aws:/root/.aws \
    -v $PWD:/bcbin/ \
    bitclouded/aws-machete $@
~~~

You may then invoke the utility:

~~~
$ aws-machete cloudformation help

CloudFormation cli utility that can can do more sophisticated actions awscli cannot, such as copy a stack, etc.

Usage:
  cloudformation [flags]
  cloudformation [command]

Available Commands:
  delete-all  delete-all
  ensure      ensure
  help        Help about any command
  update      update

Flags:
      --config-format string   Format of the configuration file. (default "yaml")
  -c, --config-path string     Config file to supply flags / parameters with.
  -h, --help                   help for cloudformation
  -m, --mode string            Modes of command execution. Valid options are: noninteractive, changesetonly, dry, interactive. (default "interactive")
  -w, --wait int               Time out in seconds to wait for the operation to complete. -1 means wait forever. (default -1)

Use "cloudformation [command] --help" for more information about a command.
~~~


## Commands

### delete-all

Deletes all cloudformation stacks in all regions except the ones with termination protection enabled.

### update

Update a stack with only have to specify new / updated parameters. Parameters not specified will use previous values. It also trims the unused parameters. This is so ci config / commands don't error when the template has parameters are removed.

### ensure

Creates a new stack if one does not exist. If it exist, update it. Similar story with update command in terms of parameter.

## Config

Values can be passed in via cli flags.

Alternatively, you can use yaml files. See src/cloudformation/test-assets for examples.

Environment variable is also accepted. A prefix of "AWS_MACHETE_" is needed. e.g. AWS_MACHETE_MODE=dry

## Under the hood

This project uses cobra + viper.
