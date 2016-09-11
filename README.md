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

### SEE ALSO in the docs
* [cft](doc/cft.md) - compose file tool
* [cft tag](doc/cft_tag.md)	 - Changes tags on images in docker-compose files
* [cft switch](doc/cft_switch.md)	 - Switches comments on image and build commands
