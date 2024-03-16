# Remote control

### **Build and run**
```
go build -o ./remote-control
chmod +x remote-control
./remote-control
```

### **Configure**

Config file name ```config.yaml```
Sample config with comments:
```yaml
authToken: 1234569870 # Authentication token
ipAddress: 127.0.0.1
port: 8080
commands:
    - name: ping # Human-readable command name
      uri: /cmd/os/ping # URI for call command
      cmd: ping -c 5 8.8.8.8 # Command such as OS 
    - name: pong
      uri: /cmd/os/pong
      cmd: pong
```

### Example

```bash
> curl -X GET -H "Authorization-token: 1234569870" http://localhost:8080/cmd/os/ping                                         4s
{"message": "Successfully executed. Command: ping_8_8_8_8"}
```
