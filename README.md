<img src="https://raw.githubusercontent.com/Nineluj/cloudcrackr/master/.github/images/banner.png" alt="logo"/>

Cloudcrackr is a command line tool orchestrating password cracking on AWS using Docker containers and ECS.

## Requirements
* AWS account with configured profile on your local machine
    * Easiest way is to install and use the AWS CLI . See [here](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html) for instructions
* Docker
    * Install instructions [here](https://docs.docker.com/engine/install/)

## Configuration
Upon running the CLI you will be prompted to configure the program. A configuration file will be
created at `~/.cloudcrackr.yaml`. Additionally, a custom configuration file can be using
by using the `--config` flag. See the `config` command for further information.

## Running - WIP
After the first release there will be more instructions on how to run the program.


## Rationale
For a security classes that I took as an undergraduate,
me and my classmates had the task of cracking passwords. However, many students were left disadvantaged
by not having the means or space to own a powerful computer that would have made this homework trivial.
The purpose of this project is to provide password cracking to those that, for any reason,
find themselves needing a more accessible option.

I also hope that this project can promote a better understanding of password cracking procedures
and encourage better practices in selecting passwords. 

## Contributing / Help
This project is in active development. Feel free to open up an issue or
contact me if you have any questions or want to get involved.