## 说明（readme）

tcp_service_v1 简单的实现了，HTTP 协议、自定义 Stream 协议、WebSocket 协议。

为方便理解分布式应用的底层细节 tcp_service_v1 作为的示例代码就不动了。

拷贝一份代码出来到 tcp_service_v2，后续将在 tcp_service_v2 上实现一个简单地分布式应用。

简单的实现了：

- API Gateway
- 内部用户服务

基本架构为：

- 内部服务启动后自动到 API Gateway 注册服务和接口。
- 内部服务和 API Gateway 之间通过自定义 Stream 协议以长链接的方式访问。
- 内部服务和 API Gateway 之间建立心跳检测机制。
- 外部流量通过 HTTP 1.1 以短链接的方式访问 API Gateway，然后转发到内部服务。