# 配置项的默认值是 :3000，取消注释这个参数可以定义自己要监听的地址和端口
# listen = "0.0.0.0:3000"

proxy = "socks5://127.0.0.1:7890"
proxy_enabled = false

# 日志级别参考Fiber框架的日志定义
# 常用的日志级别有：debug | info | warn | error | trace
log_level = "debug"

# 匹配地址的正则表达式规则是按Go语言的正则写的
# 可以按需增加，重启进程就会读取
rules = [
  "^(?:https?://)?github\\.com/([^/]+)/([^/]+)/(?:releases|archive)/.*$",
  "^(?:https?://)?github\\.com/([^/]+)/([^/]+)/(?:blob|raw)/.*$",
  "^(?:https?://)?github\\.com/([^/]+)/([^/]+)/(?:info|git-).*$",
  "^(?:https?://)?raw\\.github(?:usercontent|)\\.com/([^/]+)/([^/]+)/.+?/.*$",
  "^(?:https?://)?gist\\.(?:githubusercontent|github)\\.com/([^/]+)/([^/]+).*$",
  "^(?:https?://)?api\\.github\\.com/repos/([^/]+)/([^/]+)/.*$",
  "^(?:https?://)?huggingface\\.co(?:/spaces)?/([^/]+)/(.*)$",
  "^(?:https?://)?cdn-lfs\\.hf\\.co(?:/spaces)?/([^/]+)/([^/]+)(?:/(.*))?",
  "^(?:https?://)?download\\.docker\\.com/([^/]+)/.*\\.(tgz|zip)",
  "^(?:https?://)?(github|opengraph)\\.githubassets\\.com/([^/]+)/.*$",
  "^(?:https?://)?release-assets\\.githubusercontent\\.com/([^/]+)/([^/]+).*$",
  "^(?:https?://)?codeload\\.github\\.com/([^/]+)/([^/]+)/.*$",
]

