
# SuperHost

SuperHost 是一个功能强大的开源工具，允许用户扩展并自定义 `hosts` 文件解析规则。通过 SuperHost，可以灵活地控制域名解析，包括普通解析、泛域名解析、别名解析、指定 DNS 解析等，帮助用户更方便地管理本地网络环境。以增强系统自带hosts的不够强大的问题.

<br>

## 特性

- **普通解析**：标准的 IP 地址与域名的映射。
- **泛解析**：支持通配符域名解析。
- **别名解析**：允许将一个域名解析到另一个域名。
- **指定 DNS 解析**：为指定域名设置自定义 DNS 解析服务器。
- **DNS 别名解析**：支持将域名解析请求以其他域名的 DNS 服务器解析。

<br>

## 构建

 ```bash
go build
 ```

<br>

## 运行 SuperHost：

```bash
./superhost 127.0.0.1:8081 /mnt/hosts.txt debug=on/off
```
	
- superhost会在本地启动一个http proxy端口，只需把浏览器 或系统的http proxy设置为127.0.0.1:8081，即可使用superhosts的增加hosts功能。
	
- /mnt/hosts.txt为手动指定的hosts文件位置，可以参考hosts.example

<br>

### 启用 doh
配置doh证书。


 ```bash
openssl req -x509 -nodes -newkey rsa:2048 -days 365 -keyout doh.key -out doh.crt
```
由于doh被设计为必须启用证书，否则不被兼容。superhosts的doh证书预期设计为只在本机localhost环境中使用，可浏览器手动检查证书指纹是否为自己生成的

**请确保证书是自行生成的安全证书，不要使用他人生成的证书**。

<br>

## 用法

### 优先级

优先级暂时未定义，可能出现预料之外的情况

<br>

### 1. 普通 Hosts 解析

标准的 `hosts` 文件解析格式：

```text
127.0.0.1 localhost
::1 localhost
192.168.1.1 www.example.com
192.168.1.2 web2.example.com
````

### 2. 泛域名解析

```text
127.0.0.1 *.example.com
```

### 3. 别名解析

可将一个域名解析到另一个域名。例如，`web1.example.com` 将解析到 `web2.example.com`：

```text
@web2.example.com web1.example.com
```

### 4. 指定 DNS 解析

可为特定的域名指定使用特定的 DNS 解析服务器。支持普通 DNS 和 DoH（DNS over HTTPS）。

* 使用普通 DNS：

```text
!dns=8.8.8.8:53 web3.example.com
```

* 使用 DoH：

```text
!dns=https://doh.example.com/dns-query web5.example.com
```

### 5. DNS 别名解析

可指定一个域名使用另一个域名的 DNS 解析服务器。比如，`web4.example.com` 使用 `web3.example.com` 的 DNS 解析：

```text
@!web3.example.com web4.example.com
```

<br>

## 示例配置

superhosts配置示例，可根据自己的需求修改hosts配置：

```text
# 标准hosts
127.0.0.1 localhost
::1 localhost
192.168.1.1 www.example.com
192.168.1.2 web2.example.com

# 泛解析
127.0.0.1 *.example.com

# 别名解析
@web2.example.com web1.example.com

# 指定 DNS 解析
!dns=8.8.8.8:53 web3.example.com
!dns=https://doh.example.com/dns-query web5.example.com

# DNS 别名解析
@!web3.example.com web4.example.com
```

<br>


## 开发

软件还在开发阶段，可能有一些功能未到达预期要求。可根据自己情况手动增强。
