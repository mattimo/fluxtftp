# fluxtftp

A small tftp server that supports adding files via a client.

## Usage

To use the server you may start it in deamon mode
```
$> fluxtftp -d &
```
and then Update the file that shall be server with
```
$> fluxtftp <filename>
```

## Limitation

- Currently we only serve on Port 6969 
- The server should neve be run as root. 
- No configuration, everything is hardcoded

## Thanks

Thanks to epeli, author of [hookftp](https://github.com/epeli/hooktftp) from 
where the tftp code was taken.
