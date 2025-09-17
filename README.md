# ghproxy

结合 [HubProxy](https://github.com/sky22333/hubproxy) 和 [gh-proxy](https://github.com/hunshcn/gh-proxy) 项目编写的一个轻量的Github加速服务，支持命令行clone操作

这个重写的项目是用 [Fiber](https://gofiber.io/) 框架编写的，主要靠fiber自身封装的Client连接池实现的高性能处理反向代理任务，通过配置文件定义匹配的地址规则，作为提供服务的反向代理白名单。


**设计思路**

> 这个软件的主要原理是通过正则表达式匹配接收到的HTTP请求的路径，作为白名单的功能专门加速指定的那些网站，
> 这个软件一般是运行在国际上某个带宽比较高的IDC中的服务器上面，原链接加到这个服务的地址后面，就可以把请求转发到源站了。


## 用法

安装到当前目录

```bash
# 添加一些build参数，用来去除二进制文件中的调试符合，减小文件尺寸
GOBIN=`pwd`/bin CGO_ENABLE=0 go install -gcflags="all=-N -l" -ldflags='-w -s' -trimpath https://github.com/gnuos/ghproxy@latest
```

从仓库里下载cfg.hcl文件，和安装的二进制文件放到一起

```bash
# 给文件设置可执行权限
chmod +x ghproxy
```

## 使用说明

这个项目主要的特点就是足够简单，并且基本不用配置，默认用的是3000端口，如果要使用80端口，应该使用systemd服务文件添加一下内核权限。

如果是打算部署到公网上面给自己开发项目用的那些github包做加速，建议搭配Caddy一起使用，Caddy可以做反向代理并且能够自动申请签发Let's Encrypt的SSL证书。

项目使用的框架Fiber是支持自动多进程启动的，能够大幅提高并发能力，但是默认参数是没有启动的，可以在配置文件中开启。

内部还对http和socks5代理做了适配，如果是放到局域网内做资源加速，这些代理就是在对外转发请求的时候起作用的，Fiber框架用的是fasthttp自身实现的客户端封装。

项目有一部分代码是复制了HubProxy的代码，主要是处理重定向链接的那部分，还有处理powershell脚本和shell脚本的那部分，用于方便用户直接请求获取单个文件的。


## 配置启动服务

```systemd
[Unit]
Description=ghproxy server
After=network.target

[Service]
Type=notify
ExecStart=/usr/local/bin/ghproxy
TimeoutStopSec=5s
LimitNOFILE=1048576
LimitNPROC=512
PrivateTmp=true
ProtectHome=true
ProtectSystem=full
AmbientCapabilities=CAP_NET_BIND_SERVICE

KillMode=mixed
KillSignal=SIGINT

[Install]
WantedBy=multi-user.target
```


