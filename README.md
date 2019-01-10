## TufTree - TUF + OSTree with a Dash of "Personality"

This project is a simple command line client targeted for embedded
systems based on [OSTree](https://ostree.readthedocs.io/en/latest/).
It compliments OSTree by providing some security benefits of
[The Update Framework](https://theupdateframework.github.io/), TUF.
Lastly it adds an optional ability to configure a "personality" for
a device by applying a docker-compose file which is also backed
by TUF.

## Why Invent Something New?

You might say we aren't. We are integrating three really nice
technologies into one simple wrapper.

However, there are alternatives that could be viewed as competitors. In
general they either lack TUF, do A/B partitioning, or aren't open source.

## Why Not Aktualizr?

Aktualizr is great, but its TUF + Uptane. Uptane isn't needed in many
use cases making its complexity not worth the effort.

## How It Works

A TUF compliant server such as Docker Notary provides a device with two
types of targets files, one for the base image backed by OSTree and one
two specify the "personality". The personality is currently backed
by docker-compose, but the design is flexible enough for alternate
approaches.

### OSTREE type targets
~~~
  {
    "v38-hikey": { //one target per hardware platform
      "custom": {
        "ostree": "https://api.foundries.io/lmp/treehub/release/api/v2/",
        "targetFormat": "OSTREE",
        "uri": "https://app.foundries.io/mp/38"
      }
      "length": 0
      "hashes": {"sha256": "ostree hash for device"}
    }
  }...
~~~

### DOCKER_COMPOSER type targets
~~~
  {
    "v38": {
      "custom": {
        "compose-env": {
          "TAG": "38",  # enviroment options to pass to docker-compose
        },
        "compose-files": ["optional list of files if not docker-compose.yml"],
        "targetFormat": "DOCKER_COMPOSE",
        "tgz": "https://github.com/foundriesio/gateway-containers/archive/mp-37.tar.gz",
        "tgzLeadingDir": true,  # Removing leading directory in tgz file
        "uri": "https://app.foundries.io/mp/38"
      }
      "length": 0
      "hashes": {"sha256": "hash of tarball"}
    }
  }...
~~~

## Deploying Your Own System

Look at the [example-backend](example-backend/README.md) for instructions.
