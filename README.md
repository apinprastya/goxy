# goxy
Simple golang reverse proxy with virtual host. Replacement for apache or nginx reverse proxy with very simple config and setup. Let's Encrypt included here so you don't need to worry about certificate.

### Binary file
Check the release link : [release](https://github.com/apinprastya/goxy/releases) 
### Compile
Checkout the repo and run this command on terminal / command line :
make sure you Golang version minimum is 1.14
```bash
$ go build -o goxy main.go
```

### Config
Create a file at the same folder as the binary file with name : 
```bash
.goxy.json
```
Use .goxy.json.example as an example of the json content
```bash
$ cp .goxy.json.example .goxy.json
```
And edit the file as your needs
```json
{
  "cache": "/root/cert",
  "vhosts": [
    {
      "domain": [
        "lekapin.com",
        "www.lekapin.com"
      ],
      "target": "http://127.0.0.1:8110"
    },
    {
      "domain": [
        "test2.lekapin.com"
      ],
      "target": "http://172.0.1.15:6300"
    }
  ]
}
```
- cache: is used to store the certificate from Let's Encrypt
- domain: the list of domain you want server for the target
- target: where is the connection will be reverse proxied

### Running
As the application is running on port 80 and 443, so you will need **ROOT** to run the app.
```bash
# ./goxy
```
To run it on background use 
```bash
# ./goxy &
```
