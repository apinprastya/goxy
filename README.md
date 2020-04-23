# goxy
Simple golang reverse proxy with virtual host. Replacement for apache or nginx reverse proxy with very simple config and setup. Certbot included here so you don't need to worry about certificate.

### Compile
Checkout the repo and run this command on terminal / command line :
make sure you Golang version minimum is 1.14
```bash
$ go build -o goxy main.go`
```

### Config
Create a file with name : .goxy.json
Use .goxy.json.example as an example of the json content
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

### Running
As the application is running on port 80 and 443, so you will need root to run the app.
```bash
# ./goxy
```
To run it on background use 
```bash
# ./goxy &
```