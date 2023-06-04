# tune

Tune is a utility for forwarding local ports to private resources in a VPC using bastion hosts and [SSM Session Manager](https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager.html). Features include:

 * using **existing AWS credentials** for authentication and access control
 * support for bastion hosts in private subnets with **no open inbound ports**
 * **autodiscovery** for SSM-enabled bastion hosts and (eventually) services
 * authentication using **ephemeral SSH certificates** sent via [EC2 Instance Connect](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/Connect-using-EC2-Instance-Connect.html)

## Prerequisites

Tune requires a working installation of `awscli` with the [`session-manager-plugin`](https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-install-plugin.html) installed.

For bastion host autodiscovery, Tune requires at least one EC2 instance to be [configured for Session Manager](https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-getting-started.html) and tagged with `TuneJumpHost`.

## Installation

Download the latest release from [Releases](https://github.com/bfogarty/tune/releases). Extract the binary and add it to your `PATH`.

## Usage Examples

### Connecting to remote database

The following forwards `localhost:5433` to `my.db.com:5432` inside the VPC.

    tune to my.db.com 5432 --localPort 5433

### Using an assumed IAM Role

Tune respects AWS credentials set in `~/.aws/credentials` as well as environment variables such as `AWS_PROFILE`.

    AWS_PROFILE=qa tune to my.db.com 5432 --localPort 5433
