## Webdata Datakit Framework (WDF) Usage Doc

---

### 关于

- 此项目作为 cloude-care dataway 关联构建，作为中间层负责数据暂存。

- 数据存储中间件使用 [NSQ ](https://nsq.io/)，用户需要配置 NSQ lookupd host 和`topic`等参数。

### 版本

> Version: v2.0.2

### 运行环境

> 推荐使用 go version 1.12.13 及以上版本

### 安装和编译

> 执行命令 `go get github.com/CloudCare/web-datakit-framework`
> `go build main.go`

### 配置文件

程序会自动加载当前目录下的`wdf.conf`，或使用`-cfg`参数指定。
使用命令行参数`-newcfg`，程序会在当前目录新建一个名为`wdf.conf.example`的示例配置文件。

配置文件的格式是**toml**，其中`callback`为数组，可以拥有多个。

``` toml
## 全局配置
[global]

    ## 日志路径
    log_path = "log"

    ## 监听端口
    listen = ":8080"

    ## nsq lookupd 地址，需要添加协议头"http://"
    nsq_lookupd_addr = "http://127.0.0.1:4161"

    ## 定时器发送的topic，为空值时将不启动定时器
    timer_topic = "timer"

    ## 定时器周期，单位是秒，值小于等于 0 时不启动定时器
    timer_cycle = 60

## 回调组，可以有多个
[[callback]]

    ## 接收 url，即“路由”，首字符的斜线不能少
    callback_url = "/xxx"

    ## 验证脚本，除非需要验证，否则为空值
    # verify_bash = "/tmp/test.sh"
```
    
### 使用方式

#### 用户自行注册回调

此方式适用于较为复杂的回调情景，当注册回调和接收数据时需要大量的交互（包括权限验证、加密解密等），相关功能需要由用户自行完成，wdf 将作为数据队列来使用。

以使用钉钉回调为例，[官网文档](https://ding-doc.dingtalk.com/doc#/serverapi2/pwz3r5)，[官方demo](https://github.com/opendingtalk/eapp-corp-project/blob/master/src/main/java/com/controller/CallbackController.java)

钉钉官方 SDK 中，负责注册回调的文件是`CallbackController.java`，在完成注册回调之后，将接收到的 event 数据做解密操作，并发送给 wdf 的 http 服务器

发送示例：
    
``` java
	// private static final String WDF_URL = "http://IP:PORT/dingtalk";

	// public Map<string, string> Callback(xxxxx) 
	// 根据回调数据类型做不同的业务处理

	String eventType = obj.getString("EventType");
	
	if (! eventType.equals("check_url")) {
		RestTemplate client = new RestTemplate();

		//新建Http头，add方法可以添加参数
		HttpHeaders headers = new HttpHeaders();
		//设置请求发送方式
		HttpMethod method = HttpMethod.POST;
		// 以表单的方式提交
		headers.setContentType(MediaType.APPLICATION_JSON_UTF8);
		//将请求头部和参数合成一个请求
		HttpEntity<MultiValueMap<String, String>> requestEntity = new HttpEntity(plainText, headers);
		//执行HTTP请求，将返回的结构使用String 类格式化（可设置为对应返回值格式的类）
		ResponseEntity<String> response = client.exchange(WDF_URL, method, requestEntity, String.class);
		
		if (response.getStatusCode() != HttpStatus.OK) {
		    mainLogger.error("callback event send to wdf failed");
		}
	}
```

流程简述：

0. 启动 wdf 程序，配置相应的 callback url
1. 用户使用回调的官方 SDK 进行注册和验证，注意此时的注册 callback 地址是 SDK 的 http 服务，和 wdf 无关
2. 在成功注册，并能顺利接受和解析数据后，可以对数据进行添加或修改操作，然后创建 http 请求，将数据作为 body 发送给 wdf
3. wdf 在接收到数据后，会以 callback url 为 topic 存入 NSQ 服务中（默认忽略url的斜线，例如'/dingtalk'，那么 topic 就是'dingtalk'）
4. 数据在存入 NSQ 时会添加时间戳，用户对 NSQ 消费时可再做进一步处理

#### 简单的回调验证

对于一些简单的回调，用户只需进行注册，数据的接收由 wdf 完成，例如：

有回调服务器 cb，用户将 wdf 中配置的 callback url 作为回调接收地址，注册到 cb，按照一般步骤，cb 会向回调地址发送 http 请求，进行验证。

假设 cb 的验证方式很简单，只需要回调地址给它回复一条特殊字符串即可（比如'hello,wordl"），那么可以在 wdf 的 callback 项，配置`verify_bash`参数。

``` bash
echo "hello,world"
```

当此 url 收到请求后，会执行 bash 脚本，并将标准输出作为 body 返回给 http 的发起方，用以简单验证。

**注意：执行 bash 脚本并将标准输出作为返回的情况，只执行一次，此 url 从收到的第 2 条消息开始，便不再执行这段验证操作**

### 定时器

当定时器启动后，会每隔一段时间往 NSQ 的特定 topic 中发送一个字节的无意义数据，当 NSQ 的消费端订阅这个 topic 后，会定期收到这段数据达到定时功能。

#### 使用方法

修改配置文件`global`中的`timer_topic`和`timer_cycle`项：

> timer_topic：定时器在 NSQ 中生产和消费的 topic

> timer_cycle: 周期，单位是秒

注，当 “topic” 为空，或 “cycle” 小于等于 0 时，则认为不开启定时器功能，会在 log 中记录此事件。

### consumer NSQ

NSQ 拥有众多编程语言的实现库，可自行查阅。项目实例中使用 python3 pynsq 进行演示。


