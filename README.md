### docker-compose-file-tool (cft)

A tool for modifying docker-compose file via CLI

```bash
Tool for modifying docker-compose files via CLI and some additional neat automations

Usage:
  cft [command]

Available Commands:
  switch      Switches comments on image and build commands
  tag         Changes tags on images in docker-compose files

Flags:
  -c, --compose-file string   docker-compose file to change, if none set $CFT_COMPOSE will be used
  -f, --force                 Skips security confirmation prompts

Use "cft [command] --help" for more information about a command.
```

## switching services from images to build and vice versa
```bash
$ cat docker-compose.yml
version: '2'
services:
    mysql:
#		image: mysql
        build: /path/to/mysql
    mongo:
#		image: mongo
		build: /path/to/mongo

$ cft -c docker-compose.yml switch mysql
Changes:
- #		image: mysql
-         build: /path/to/mysql
+ 		image: mysql
+ #        build: /path/to/mysql

$ cat docker-compose.yml
version: '2'
services:
    mysql:
		image: mysql
#        build: /path/to/mysql
    mongo:
#		image: mongo
		build: /path/to/mongo

```

## tagging images with a specific tag or remove all tags
```bash
$ cat docker-compose.yml
version: '2'
services:
    mysql:
		image: mysql
#        build: /path/to/mysql
    mongo:
#		image: mongo
		build: /path/to/mongo

$ cft -c docker-compose.yml tag mysql -t latest
Changes:
- 		image: mysql
+ 		image: mysql:latest

$ cat docker-compose.yml
version: '2'
services:
    mysql:
		image: mysql:latest
#        build: /path/to/mysql
    mongo:
#		image: mongo
		build: /path/to/mongo

$ cft -c docker-compose.yml tag
No tag nor image pattern given, really remove all tags from all images? [y/n]
y
Changes:
- 		image: mysql:latest
+ 		image: mysql

$ cat docker-compose.yml
version: '2'
services:
    mysql:
		image: mysql
#        build: /path/to/mysql
    mongo:
#		image: mongo
		build: /path/to/mongo
```


### SEE ALSO in the docs
* [cft](doc/cft.md) - compose file tool
* [cft tag](doc/cft_tag.md)	 - Changes tags on images in docker-compose files
* [cft switch](doc/cft_switch.md)	 - Switches comments on image and build commands
